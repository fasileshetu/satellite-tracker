"use client";

import { useEffect, useState } from "react";
import { useAuth } from "@/lib/auth-context";
import { ApiError, createComponent, listComponents, updateComponentStatus } from "@/lib/api";
import type { Component, ComponentStatus } from "@/lib/types";

const STATUSES: ComponentStatus[] = ["received", "in_test", "pass", "fail", "flight_ready"];

export function ComponentsPanel() {
  const { session, configured } = useAuth();
  // Writes are only actually gated when Cognito is configured -- that's
  // what internal/api/router.go bases its own auth check on (a nil
  // verifier when OIDC_ISSUER_URL/OIDC_CLIENT_ID are unset, e.g. local
  // docker-compose). So local dev can still exercise the create/update
  // flows without a login; accessToken is simply omitted from the request
  // in that case, matching what the unauthenticated API expects.
  const canWrite = session || !configured;
  const [components, setComponents] = useState<Component[]>([]);
  const [satelliteFilter, setSatelliteFilter] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [satelliteId, setSatelliteId] = useState("");
  const [name, setName] = useState("");
  const [partNumber, setPartNumber] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function refresh() {
    setLoading(true);
    setError(null);
    try {
      setComponents(await listComponents(satelliteFilter || undefined));
    } catch (err) {
      setError(err instanceof Error ? err.message : "failed to load components");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    refresh();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [satelliteFilter]);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!canWrite) return;
    setSubmitting(true);
    setError(null);
    try {
      await createComponent(
        { satellite_id: satelliteId, name, part_number: partNumber },
        session?.accessToken ?? ""
      );
      setSatelliteId("");
      setName("");
      setPartNumber("");
      await refresh();
    } catch (err) {
      setError(
        err instanceof ApiError && err.status === 401
          ? "your session expired -- log in again"
          : err instanceof Error
          ? err.message
          : "failed to create component"
      );
    } finally {
      setSubmitting(false);
    }
  }

  async function handleStatusChange(id: number, status: string) {
    if (!canWrite) return;
    setError(null);
    try {
      await updateComponentStatus(id, status, session?.accessToken ?? "");
      await refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "failed to update status");
    }
  }

  return (
    <>
      <div className="panel">
        <h2>Register a component</h2>
        {!canWrite ? (
          <p className="muted">Log in to register components or update their status.</p>
        ) : (
          <form className="inline" onSubmit={handleCreate} data-cy="create-form">
            <label>
              Satellite ID
              <input
                required
                value={satelliteId}
                onChange={(e) => setSatelliteId(e.target.value)}
                placeholder="K2-GRAVITAS-2"
                data-cy="satellite-id-input"
              />
            </label>
            <label>
              Name
              <input
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Avionics Board Rev C"
                data-cy="name-input"
              />
            </label>
            <label>
              Part number
              <input
                required
                value={partNumber}
                onChange={(e) => setPartNumber(e.target.value)}
                placeholder="AV-2201"
                data-cy="part-number-input"
              />
            </label>
            <button type="submit" disabled={submitting} data-cy="submit-button">
              {submitting ? "Adding..." : "Add component"}
            </button>
          </form>
        )}
        {error && <div className="error">{error}</div>}
      </div>

      <div className="panel">
        <h2>Components</h2>
        <form className="inline" style={{ marginBottom: 14 }}>
          <label>
            Filter by satellite ID
            <input
              value={satelliteFilter}
              onChange={(e) => setSatelliteFilter(e.target.value)}
              placeholder="(all satellites)"
              data-cy="satellite-filter-input"
            />
          </label>
        </form>

        {loading ? (
          <p className="muted" data-cy="loading">
            Loading...
          </p>
        ) : components.length === 0 ? (
          <p className="muted" data-cy="empty-state">
            No components yet.
          </p>
        ) : (
          <table data-cy="components-table">
            <thead>
              <tr>
                <th>Satellite</th>
                <th>Name</th>
                <th>Part #</th>
                <th>Status</th>
                <th>Updated</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {components.map((c) => (
                <tr key={c.id} data-cy="component-row" data-satellite-id={c.satellite_id}>
                  <td>
                    <a
                      href={`/telemetry/${encodeURIComponent(c.satellite_id)}`}
                      data-cy="telemetry-link"
                    >
                      {c.satellite_id}
                    </a>
                  </td>
                  <td>{c.name}</td>
                  <td>{c.part_number}</td>
                  <td>
                    <span className={`badge ${c.status}`} data-cy="status-badge">
                      {c.status}
                    </span>
                  </td>
                  <td className="muted">{new Date(c.updated_at).toLocaleString()}</td>
                  <td>
                    {canWrite && (
                      <select
                        value=""
                        data-cy="status-select"
                        onChange={(e) => {
                          if (e.target.value) handleStatusChange(c.id, e.target.value);
                        }}
                      >
                        <option value="">move to...</option>
                        {STATUSES.filter((s) => s !== c.status).map((s) => (
                          <option key={s} value={s}>
                            {s}
                          </option>
                        ))}
                      </select>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </>
  );
}
