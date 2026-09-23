// PKCE (Proof Key for Code Exchange, RFC 7636) for the Authorization Code
// flow against Cognito's Hosted UI. The app is a public client (no client
// secret -- see `generate_secret = false` in terraform/cognito.tf), and
// Cognito requires PKCE for public clients using the code flow, so this
// isn't optional scaffolding: without it the token exchange is rejected.
//
// Runs in the browser only (uses window.crypto), ahead of the redirect to
// Cognito.

function base64UrlEncode(bytes: ArrayBuffer): string {
  const binary = String.fromCharCode(...new Uint8Array(bytes));
  return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}

export function generateCodeVerifier(): string {
  const bytes = new Uint8Array(32);
  crypto.getRandomValues(bytes);
  return base64UrlEncode(bytes.buffer);
}

export async function generateCodeChallenge(verifier: string): Promise<string> {
  const data = new TextEncoder().encode(verifier);
  const digest = await crypto.subtle.digest("SHA-256", data);
  return base64UrlEncode(digest);
}
