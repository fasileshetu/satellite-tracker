import { NextRequest, NextResponse } from "next/server";
import { getTelemetryHistory } from "@/lib/grpc-telemetry";

export async function GET(
  req: NextRequest,
  { params }: { params: Promise<{ satelliteId: string }> }
) {
  const { satelliteId } = await params;
  const limitParam = req.nextUrl.searchParams.get("limit");
  const limit = limitParam ? parseInt(limitParam, 10) : 20;

  try {
    const readings = await getTelemetryHistory(satelliteId, limit);
    return NextResponse.json({ readings });
  } catch (err) {
    const message = err instanceof Error ? err.message : "gRPC call failed";
    // grpc-server (ClusterIP-only in EKS, or docker-compose locally) is
    // unreachable most often because it isn't running / isn't
    // port-forwarded -- surface that as a 502, not a 500.
    return NextResponse.json({ error: message }, { status: 502 });
  }
}
