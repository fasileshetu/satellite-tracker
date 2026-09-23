export const cognitoConfig = {
  domain: process.env.NEXT_PUBLIC_COGNITO_DOMAIN ?? "",
  region: process.env.NEXT_PUBLIC_COGNITO_REGION ?? "us-west-2",
  clientId: process.env.NEXT_PUBLIC_COGNITO_CLIENT_ID ?? "",
  redirectUri:
    process.env.NEXT_PUBLIC_COGNITO_REDIRECT_URI ?? "http://localhost:3000/callback",
};

export function isCognitoConfigured(): boolean {
  return Boolean(cognitoConfig.domain && cognitoConfig.clientId);
}

export function hostedUiBaseUrl(): string {
  return `https://${cognitoConfig.domain}.auth.${cognitoConfig.region}.amazoncognito.com`;
}

export function buildAuthorizeUrl(codeChallenge: string): string {
  const params = new URLSearchParams({
    client_id: cognitoConfig.clientId,
    response_type: "code",
    scope: "openid email profile",
    redirect_uri: cognitoConfig.redirectUri,
    code_challenge_method: "S256",
    code_challenge: codeChallenge,
  });
  return `${hostedUiBaseUrl()}/login?${params.toString()}`;
}

export function buildLogoutUrl(): string {
  const params = new URLSearchParams({
    client_id: cognitoConfig.clientId,
    logout_uri: cognitoConfig.redirectUri.replace(/\/callback$/, ""),
  });
  return `${hostedUiBaseUrl()}/logout?${params.toString()}`;
}
