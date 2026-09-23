import type { Component, NewComponentInput } from "./types";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {},
  accessToken?: string | null
): Promise<T> {
  const headers = new Headers(options.headers);
  headers.set("Content-Type", "application/json");
  if (accessToken) {
    headers.set("Authorization", `Bearer ${accessToken}`);
  }

  const res = await fetch(`${API_URL}${path}`, { ...options, headers });
  if (!res.ok) {
    let message = res.statusText;
    try {
      const body = await res.json();
      if (body?.error) message = body.error;
    } catch {
      // response wasn't JSON -- fall back to statusText
    }
    throw new ApiError(res.status, message);
  }
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

export function listComponents(satelliteId?: string): Promise<Component[]> {
  const qs = satelliteId ? `?satellite_id=${encodeURIComponent(satelliteId)}` : "";
  return request<Component[]>(`/components${qs}`);
}

export function getComponent(id: number): Promise<Component> {
  return request<Component>(`/components/${id}`);
}

// Both writes require a Cognito access token -- internal/api/router.go puts
// auth.Verifier.Middleware on these two routes only. Reads stay open.
export function createComponent(
  input: NewComponentInput,
  accessToken: string
): Promise<Component> {
  return request<Component>(
    "/components",
    { method: "POST", body: JSON.stringify(input) },
    accessToken
  );
}

export function updateComponentStatus(
  id: number,
  status: string,
  accessToken: string
): Promise<Component> {
  return request<Component>(
    `/components/${id}/status`,
    { method: "PATCH", body: JSON.stringify({ status }) },
    accessToken
  );
}

export { ApiError };
