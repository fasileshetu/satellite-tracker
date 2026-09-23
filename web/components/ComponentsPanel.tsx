"use client";

import { useEffect, useState } from "react";
import { useAuth } from "@/lib/auth-context";
import { ApiError, createComponent, listComponents, updateComponentStatus } from "@/lib/api";
import type { Component, ComponentStatus } from "@/lib/types";

const STATUSES: ComponentStatus[] = ["received", "in_test", "pass", "fail", "flight_ready"];

export function ComponentsPanel() {
  const { session } = useAuth();
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
    if (!session) return;
    setSubmitting(true);
    setError(null);
    try {
      await createComponent(
        { satellite_id: satelliteId, name, part_number: partNumber },
        session.accessToken
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
    if (!session) return;
    setError(null);
    try {
      await updateComponentStatus(id, status, session.accessToken);
      await refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "failed to update status");
    }
  }

  return (
    <>
      <div className="panel">
        <h2>Register a component</h2>
        {!session ? (
          <p className="muted">Log in to register components or update their status.</p>
        ) : (
          <form className="inline" onSubmit={handleCreate}>
            <label>
              Satellite ID
              <input
                required
                value={satelliteId}
                onChange={(e) => setSatelliteId(e.target.value)}
                placeholder="K2-GRAVITAS-2"
              />
            </label>
            <label>
              Name
              <input
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Avionics Board Rev C"
              />
            </label>
            <label>
              Part number
              <input
                required
                value={partNumber}
                onChange={(e) => setPartNumber(e.target.value)}
                placeholder="AV-2201"
              />
            </label>
            <button type="submit" disabled={submitting}>
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
            />
          </label>
        </form>

        {loading ? (
          <p className="muted">Loading...</p>
        ) : components.length === 0 ? (
          <p className="muted">No components yet.</p>
        ) : (
          <table>
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
                <tr key={c.id}>
                  <td>
                    <a href={`/telemetry/${encodeURIComponent(c.satellite_id)}`}>
                      {c.satellite_id}
                    </a>
                  </td>
                  <td>{c.name}</td>
                  <td>{c.part_number}</td>
                  <td>
                    <span className={`badge ${c.status}`}>{c.status}</span>
                  </td>
                  <td className="muted">{new Date(c.updated_at).toLocaleString()}</td>
                  <td>
                    {session && (
                      <select
                        value=""
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
