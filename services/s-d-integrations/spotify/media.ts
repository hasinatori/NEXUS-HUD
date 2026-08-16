// Zuletzt geändert: 2026-08-16
// Mapping der Spotify-Web-API-Antwort (GET /me/player/currently-playing) auf
// die event.media.state-Params. Reine Funktion -> ohne HTTP testbar.

export interface MediaState {
  playing: boolean;
  track?: string;
  artist?: string;
  album?: string;
  album_art_url?: string;
  duration_ms?: number;
  progress_ms?: number;
}

interface Artist {
  name: string;
}

interface Album {
  name: string;
  images?: Array<{ url: string }>;
}

interface TrackItem {
  name?: string;
  artists?: Artist[];
  album?: Album;
  duration_ms?: number;
}

interface RawCurrentlyPlaying {
  is_playing?: boolean;
  progress_ms?: number;
  item?: TrackItem | null;
}

function joinArtists(item: TrackItem): string | undefined {
  const names = item.artists?.map((a) => a.name).filter(Boolean);
  return names && names.length > 0 ? names.join(", ") : undefined;
}

/** Mappt die rohe API-Antwort auf MediaState. Gibt null zurück, wenn nichts
 *  abgespielt wird (item fehlt/leer). */
export function toMediaState(raw: RawCurrentlyPlaying): MediaState | null {
  const item = raw.item;
  if (!item) {
    return null;
  }
  const state: MediaState = {
    playing: Boolean(raw.is_playing),
    duration_ms: item.duration_ms,
    progress_ms: raw.progress_ms,
  };
  const track = item.name?.trim();
  if (track) {
    state.track = track;
  }
  const artist = joinArtists(item);
  if (artist) {
    state.artist = artist;
  }
  const album = item.album?.name?.trim();
  if (album) {
    state.album = album;
  }
  const art = item.album?.images?.[0]?.url;
  if (art) {
    state.album_art_url = art;
  }
  return state;
}
