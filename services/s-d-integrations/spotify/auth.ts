// Zuletzt geändert: 2026-08-16
// OAuth2: Erneuern des Access-Tokens über den Refresh-Token-Flow
// (Authorization Code + PKCE beim Login; danach nur noch Refresh hier).
// Form-Codierung + Basic Auth wie von Spotify verlangt.

export interface TokenResponse {
  access_token: string;
  token_type: string;
  scope: string;
  expires_in: number;
}

export class SpotifyAuthError extends Error {}

const DEFAULT_TOKEN_URL = "https://accounts.spotify.com/api/token";

/** Tauscht einen Refresh-Token gegen ein frisches Access-Token. fetchImpl ist
 *  für Tests injizierbar; die Endpunkt-URL lässt sich per SPOTIFY_TOKEN_URL
 *  überschreiben (Tests/Staging). */
export async function refreshAccessToken(
  clientId: string,
  clientSecret: string,
  refreshToken: string,
  fetchImpl: typeof fetch = fetch,
): Promise<TokenResponse> {
  const tokenUrl = process.env.SPOTIFY_TOKEN_URL ?? DEFAULT_TOKEN_URL;
  const body = new URLSearchParams({
    grant_type: "refresh_token",
    refresh_token: refreshToken,
  });

  let res: Response;
  try {
    res = await fetchImpl(tokenUrl, {
      method: "POST",
      headers: {
        "Content-Type": "application/x-www-form-urlencoded",
        Authorization: `Basic ${Buffer.from(`${clientId}:${clientSecret}`).toString("base64")}`,
      },
      body,
    });
  } catch (err) {
    throw new SpotifyAuthError(`Token-Endpunkt nicht erreichbar: ${(err as Error).message}`);
  }

  if (!res.ok) {
    throw new SpotifyAuthError(`Token-Refresh fehlgeschlagen (HTTP ${res.status}).`);
  }

  const data = (await res.json()) as Partial<TokenResponse>;
  if (!data.access_token || typeof data.expires_in !== "number") {
    throw new SpotifyAuthError("Token-Antwort unvollständig (access_token/expires_in fehlen).");
  }
  return {
    access_token: data.access_token,
    token_type: data.token_type ?? "Bearer",
    scope: data.scope ?? "",
    expires_in: data.expires_in,
  };
}
