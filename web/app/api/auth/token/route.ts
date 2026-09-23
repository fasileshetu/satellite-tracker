import { NextRequest, NextResponse } from "next/server";
import { cognitoConfig, hostedUiBaseUrl } from "@/lib/cognito-config";

// Exchanges an Authorization Code flow `code` for tokens.
//
// This runs server-side (rather than the browser calling Cognito's token
// endpoint directly) purely to sidestep CORS -- Cognito's token endpoint
// doesn't need to allow our origin this way. Since the app is a public
// client (no client secret), this is otherwise the exact same request a
// browser could make itself; the PKCE code_verifier is what proves this
// request came from whoever started the flow, not the fact that it's
// server-side.
export async function POST(req: NextRequest) {
  const { code, codeVerifier } = await req.json();
  if (!code || !codeVerifier) {
    return NextResponse.json(
      { error: "code and codeVerifier are required" },
      { status: 400 }
    );
  }

  const body = new URLSearchParams({
    grant_type: "authorization_code",
    client_id: cognitoConfig.clientId,
    code,
    redirect_uri: cognitoConfig.redirectUri,
    code_verifier: codeVerifier,
  });

  const res = await fetch(`${hostedUiBaseUrl()}/oauth2/token`, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: body.toString(),
  });

  const data = await res.json();
  if (!res.ok) {
    return NextResponse.json(
      { error: data.error_description ?? data.error ?? "token exchange failed" },
      { status: res.status }
    );
  }

  // { access_token, id_token, refresh_token, expires_in, token_type }
  return NextResponse.json(data);
}
