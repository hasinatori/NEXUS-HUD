// Zuletzt geaendert: 2026-08-17
// Discord Rich Presence Status-Mapping. Reine Funktion -> ohne HTTP testbar.

export interface PresenceState {
  status: "online" | "idle" | "dnd" | "offline";
  activity?: string;
  activityType?: "playing" | "streaming" | "listening" | "watching" | "competing";
  details?: string;
  state?: string;
  largeImageKey?: string;
  largeImageText?: string;
}

interface RawPresence {
  status?: string;
  activities?: Array<{
    type?: number;
    name?: string;
    details?: string;
    state?: string;
    assets?: { large_image?: string; large_text?: string };
  }>;
}

const ACTIVITY_TYPE_MAP: Record<number, PresenceState["activityType"]> = {
  0: "playing",
  1: "streaming",
  2: "listening",
  3: "watching",
  5: "competing",
};

/** Mappt die rohe Discord-API-Antwort auf PresenceState. */
export function toPresenceState(raw: RawPresence): PresenceState | null {
  if (!raw.status) {
    return null;
  }
  const validStatuses: PresenceState["status"][] = ["online", "idle", "dnd", "offline"];
  const status = validStatuses.includes(raw.status as PresenceState["status"])
    ? (raw.status as PresenceState["status"])
    : "offline";

  const state: PresenceState = { status };

  const activity = raw.activities?.[0];
  if (activity) {
    if (activity.name) {
      state.activity = activity.name;
    }
    if (typeof activity.type === "number" && ACTIVITY_TYPE_MAP[activity.type]) {
      state.activityType = ACTIVITY_TYPE_MAP[activity.type];
    }
    if (activity.details) {
      state.details = activity.details;
    }
    if (activity.state) {
      state.state = activity.state;
    }
    if (activity.assets?.large_image) {
      state.largeImageKey = activity.assets.large_image;
    }
    if (activity.assets?.large_text) {
      state.largeImageText = activity.assets.large_text;
    }
  }

  return state;
}
