"use client";

import { Suspense, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useAuth } from "@/lib/auth-context";

function CallbackInner() {
  const router = useRouter();
  const params = useSearchParams();
  const { completeLogin } = useAuth();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const code = params.get("code");
    const oauthError = params.get("error_description") ?? params.get("error");
    if (oauthError) {
      setError(oauthError);
      return;
    }
    if (!code) {
      setError("no authorization code in the callback URL");
      return;
    }
    completeLogin(code)
      .then(() => router.replace("/"))
      .catch((err) => setError(err.message ?? "login failed"));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [params]);

  return (
    <div className="panel">
      <h2>Signing you in...</h2>
      {error ? (
        <>
          <p className="error">{error}</p>
          <a href="/">Back to dashboard</a>
        </>
      ) : (
        <p className="muted">Exchanging your Cognito authorization code for tokens.</p>
      )}
    </div>
  );
}

// Cognito's redirect back to /callback?code=... only carries a search
// param, so this reads it with useSearchParams -- which Next.js requires
// wrapping in Suspense so the page can still be statically analyzed.
export default function CallbackPage() {
  return (
    <Suspense fallback={<div className="panel">Loading...</div>}>
      <CallbackInner />
    </Suspense>
  );
}
