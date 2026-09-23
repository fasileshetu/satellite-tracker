// Mirrors internal/models.Component and NewComponentInput on the Go side.
export type ComponentStatus =
  | "received"
  | "in_test"
  | "pass"
  | "fail"
  | "flight_ready";

export interface Component {
  id: number;
  satellite_id: string;
  name: string;
  part_number: string;
  status: ComponentStatus;
  created_at: string;
  updated_at: string;
}

export interface NewComponentInput {
  satellite_id: string;
  name: string;
  part_number: string;
}

// Mirrors proto/telemetry.proto's TelemetryReading / TelemetryHistoryResponse.
export interface TelemetryReading {
  satellite_id: string;
  timestamp_ms: number;
  battery_voltage: number;
  temperature_celsius: number;
  signal_strength_dbm: number;
  status: string;
}
