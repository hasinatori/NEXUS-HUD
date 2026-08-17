// Zuletzt geaendert: 2026-08-17
// VoIP-Anruf-Status und Mapping. Reine Funktion -> ohne HTTP testbar.

export type CallStatus = "queued" | "ringing" | "in-progress" | "completed" | "failed" | "busy" | "no-answer" | "canceled";

export interface CallState {
  callId: string;
  to: string;
  from: string;
  status: CallStatus;
  durationSec?: number;
  startTime?: string;
  endTime?: string;
  errorCode?: number;
  errorMessage?: string;
}

interface RawCall {
  sid?: string;
  to?: string;
  from?: string;
  status?: string;
  duration?: string;
  start_time?: string;
  end_time?: string;
  error_code?: number;
  error_message?: string;
}

const VALID_STATUSES: CallStatus[] = ["queued", "ringing", "in-progress", "completed", "failed", "busy", "no-answer", "canceled"];

/** Mappt die rohe Twilio-API-Antwort auf CallState. */
export function toCallState(raw: RawCall): CallState | null {
  if (!raw.sid || !raw.to || !raw.from) {
    return null;
  }
  const status = VALID_STATUSES.includes(raw.status as CallStatus)
    ? (raw.status as CallStatus)
    : "failed";

  const state: CallState = {
    callId: raw.sid,
    to: raw.to,
    from: raw.from,
    status,
  };

  if (raw.duration) {
    const d = Number(raw.duration);
    if (Number.isFinite(d) && d >= 0) {
      state.durationSec = d;
    }
  }
  if (raw.start_time) {
    state.startTime = raw.start_time;
  }
  if (raw.end_time) {
    state.endTime = raw.end_time;
  }
  if (typeof raw.error_code === "number" && raw.error_code !== 0) {
    state.errorCode = raw.error_code;
  }
  if (raw.error_message) {
    state.errorMessage = raw.error_message;
  }

  return state;
}
