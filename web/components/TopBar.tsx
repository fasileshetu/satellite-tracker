"use client";

import { useAuth } from "@/lib/auth-context";

export function TopBar() {
  const { session, loading, configured, login, logout } = useAuth();

  return (
    <header className="topbar">
      <div>
        <h1>satellite-tracker</h1>
        <div className="sub">component tracking &amp; telemetry</div>
      </div>
      <div>
        {!configured ? (
          <span className="muted">Cognito not configured -- writes disabled</span>
        ) : loading ? null : session ? (
          <button className="secondary" onClick={logout}>
            Log out
          </button>
        ) : (
          <button onClick={login}>Log in</button>
        )}
      </div>
    </header>
  );
}
