import path from "path";
import * as grpc from "@grpc/grpc-js";
import * as protoLoader from "@grpc/proto-loader";
import type { TelemetryReading } from "./types";

// Server-side only (imported from a route handler). Browsers can't speak
// gRPC's HTTP/2 framing directly, so this is the dashboard's own
// backend-for-frontend: it proxies GetTelemetryHistory to the real
// grpc-server (internal/grpcserver) over a normal gRPC client connection
// and hands the browser plain JSON back.
//
// Loaded from the .proto source directly with @grpc/proto-loader instead
// of generated JS stubs -- Node doesn't need the Go-style codegen step
// used in cmd/grpc-server; the same telemetry.proto (copied into
// web/proto/) is enough.

const PROTO_PATH = path.join(process.cwd(), "proto", "telemetry.proto");

// protobufjs (a proto-loader dependency) ships the well-known types
// (google/protobuf/timestamp.proto, referenced by telemetry.proto) under
// its own package directory, so it's added as an include path rather than
// vendoring a copy here.
const WELL_KNOWN_INCLUDE = path.join(process.cwd(), "node_modules", "protobufjs");

let cachedClient: any = null;

function getClient() {
  if (cachedClient) return cachedClient;

  const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
    keepCase: true,
    longs: Number,
    enums: String,
    defaults: true,
    oneofs: true,
    includeDirs: [path.dirname(PROTO_PATH), WELL_KNOWN_INCLUDE],
  });
  const proto = grpc.loadPackageDefinition(packageDefinition) as any;

  const addr = process.env.GRPC_SERVER_ADDR ?? "localhost:9090";
  cachedClient = new proto.telemetry.TelemetryService(
    addr,
    grpc.credentials.createInsecure()
  );
  return cachedClient;
}

export function getTelemetryHistory(
  satelliteId: string,
  limit = 20
): Promise<TelemetryReading[]> {
  return new Promise((resolve, reject) => {
    const client = getClient();
    const deadline = new Date(Date.now() + 5000);
    client.GetTelemetryHistory(
      { satellite_id: satelliteId, limit },
      { deadline },
      (err: grpc.ServiceError | null, response: any) => {
        if (err) {
          reject(err);
          return;
        }
        const readings: TelemetryReading[] = (response.readings ?? []).map((r: any) => ({
          satellite_id: r.satellite_id,
          // proto-loader decodes google.protobuf.Timestamp as
          // { seconds, nanos } rather than a JS Date.
          timestamp_ms:
            Number(r.timestamp?.seconds ?? 0) * 1000 +
            Math.floor(Number(r.timestamp?.nanos ?? 0) / 1e6),
          battery_voltage: r.battery_voltage,
          temperature_celsius: r.temperature_celsius,
          signal_strength_dbm: r.signal_strength_dbm,
          status: r.status,
        }));
        resolve(readings);
      }
    );
  });
}
