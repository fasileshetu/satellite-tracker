"use client";

import { useEffect, useState } from "react";
import type { TelemetryReading } from "@/lib/types";

export function TelemetryTable({ satelliteId }: { satelliteId: string }) {
  const [readings, setReadings] = useState<TelemetryReading[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    fetch(`/api/telemetry/${encodeURIComponent(satelliteId)}?limit=20`)
      .then(async (res) => {
        const data = await res.json();
        if (!res.ok) throw new Error(data.error ?? "failed to load telemetry");
        return data.readings as TelemetryReading[];
      })
      .then((r) => !cancelled && setReadings(r))
      .catch((err) => !cancelled && setError(err.message));
    return () => {
      cancelled = true;
    };
  }, [satelliteId]);

  if (error) {
    return (
      <div className="error">
        {error}
        <div className="muted" style={{ marginTop: 6 }}>
          Is grpc-server running and reachable at the address in{" "}
          <code>GRPC_SERVER_ADDR</code>? Locally: <code>docker compose up grpc-server</code>.
          In EKS: this needs to run from inside the cluster (or via{" "}
          <code>kubectl port-forward</code>), since the Service is ClusterIP-only.
        </div>
      </div>
    );
  }

  if (!readings) return <p className="muted">Loading...</p>;
  if (readings.length === 0) return <p className="muted">No telemetry readings yet.</p>;

  return (
    <table>
      <thead>
        <tr>
          <th>Time</th>
          <th>Battery (V)</th>
          <th>Temp (&deg;C)</th>
          <th>Signal (dBm)</th>
          <th>Status</th>
        </tr>
      </thead>
      <tbody>
        {readings.map((r, i) => (
          <tr key={i}>
            <td>{new Date(r.timestamp_ms).toLocaleString()}</td>
            <td>{r.battery_voltage.toFixed(2)}</td>
            <td>{r.temperature_celsius.toFixed(1)}</td>
            <td>{r.signal_strength_dbm.toFixed(1)}</td>
            <td>{r.status}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
