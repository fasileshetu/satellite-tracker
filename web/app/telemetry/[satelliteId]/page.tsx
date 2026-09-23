import Link from "next/link";
import { TopBar } from "@/components/TopBar";
import { TelemetryTable } from "@/components/TelemetryTable";

export default async function TelemetryPage({
  params,
}: {
  params: Promise<{ satelliteId: string }>;
}) {
  const { satelliteId } = await params;

  return (
    <>
      <TopBar />
      <p className="muted" style={{ marginTop: -12 }}>
        <Link href="/">&larr; back to components</Link>
      </p>
      <div className="panel">
        <h2>Telemetry -- {decodeURIComponent(satelliteId)}</h2>
        <p className="muted">
          Proxied server-side from this dashboard's own backend
          (<code>/api/telemetry/[satelliteId]</code>) to grpc-server's{" "}
          <code>GetTelemetryHistory</code> RPC, since browsers can&apos;t speak
          gRPC directly.
        </p>
        <TelemetryTable satelliteId={satelliteId} />
      </div>
    </>
  );
}
