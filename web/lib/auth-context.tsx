"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { generateCodeChallenge, generateCodeVerifier } from "./pkce";
import { buildAuthorizeUrl, buildLogoutUrl, isCognitoConfigured } from "./cognito-config";

interface Session {
  accessToken: string;
  idToken: string;
  expiresAt: number; // epoch ms
}

interface AuthContextValue {
  session: Session | null;
  loading: boolean;
  configured: boolean;
  login: () => Promise<void>;
  logout: () => void;
  completeLogin: (code: string) => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

// Demo-grade storage: sessionStorage, not an httpOnly cookie. Fine for a
// portfolio project exercising the Cognito flow end to end; a production
// app would keep tokens server-side (or in an httpOnly cookie) so they're
// never reachable from page JavaScript at all.
const STORAGE_KEY = "satellite-tracker.session";
const VERIFIER_KEY = "satellite-tracker.pkce-verifier";

function loadSession(): Session | null {
  if (typeof window === "undefined") return null;
  const raw = window.sessionStorage.getItem(STORAGE_KEY);
  if (!raw) return null;
  try {
    const parsed = JSON.parse(raw) as Session;
    if (parsed.expiresAt < Date.now()) {
      window.sessionStorage.removeItem(STORAGE_KEY);
      return null;
    }
    return parsed;
  } catch {
    return null;
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<Session | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setSession(loadSession());
    setLoading(false);
  }, []);

  const login = useCallback(async () => {
    const verifier = generateCodeVerifier();
    window.sessionStorage.setItem(VERIFIER_KEY, verifier);
    const challenge = await generateCodeChallenge(verifier);
    window.location.href = buildAuthorizeUrl(challenge);
  }, []);

  const completeLogin = useCallback(async (code: string) => {
    const codeVerifier = window.sessionStorage.getItem(VERIFIER_KEY);
    if (!codeVerifier) {
      throw new Error("missing PKCE verifier -- start the login flow again");
    }
    const res = await fetch("/api/auth/token", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ code, codeVerifier }),
    });
    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error ?? "login failed");
    }
    const next: Session = {
      accessToken: data.access_token,
      idToken: data.id_token,
      expiresAt: Date.now() + data.expires_in * 1000,
    };
    window.sessionStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    window.sessionStorage.removeItem(VERIFIER_KEY);
    setSession(next);
  }, []);

  const logout = useCallback(() => {
    window.sessionStorage.removeItem(STORAGE_KEY);
    setSession(null);
    window.location.href = buildLogoutUrl();
  }, []);

  const value = useMemo(
    () => ({ session, loading, configured: isCognitoConfigured(), login, logout, completeLogin }),
    [session, loading, login, logout, completeLogin]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
