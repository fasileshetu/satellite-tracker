// The happy-path click-through from the components list to a real,
// live telemetry page (against a real grpc-server + DynamoDB) is
// covered in components.cy.ts. This spec isolates the telemetry page's
// three UI states -- empty, populated, and backend-unreachable -- by
// intercepting its own API route (/api/telemetry/[satelliteId]), the
// same boundary the real page calls (see components/TelemetryTable.tsx).
// That keeps these three assertions deterministic regardless of whether
// grpc-server happens to be running or has data for this satellite ID,
// while still exercising the actual page and component code, not a
// mock of the whole page.

describe("telemetry page", () => {
  const satelliteId = "sat-01";

  it("shows the empty state when there are no readings yet", () => {
    cy.intercept("GET", `/api/telemetry/${satelliteId}*`, { readings: [] }).as("getTelemetry");
    cy.visit(`/telemetry/${satelliteId}`);
    cy.wait("@getTelemetry");
    cy.contains("No telemetry readings yet.").should("be.visible");
  });

  it("renders real-shaped readings in the table", () => {
    const now = Date.now();
    cy.intercept("GET", `/api/telemetry/${satelliteId}*`, {
      readings: [
        {
          satellite_id: satelliteId,
          timestamp_ms: now,
          battery_voltage: 12.02,
          temperature_celsius: 22.8,
          signal_strength_dbm: -64.2,
          status: "nominal",
        },
        {
          satellite_id: satelliteId,
          timestamp_ms: now - 1000,
          battery_voltage: 11.9,
          temperature_celsius: 21.6,
          signal_strength_dbm: -68.3,
          status: "nominal",
        },
      ],
    }).as("getTelemetry");

    cy.visit(`/telemetry/${satelliteId}`);
    cy.wait("@getTelemetry");

    cy.get("table tbody tr").should("have.length", 2);
    cy.contains("td", "12.02").should("be.visible");
    cy.contains("td", "22.8").should("be.visible");
    cy.contains("td", "-64.2").should("be.visible");
    cy.contains("td", "nominal").should("be.visible");
  });

  it("shows a clear error when grpc-server is unreachable", () => {
    cy.intercept("GET", `/api/telemetry/${satelliteId}*`, {
      statusCode: 502,
      body: { error: "gRPC call failed: connection refused" },
    }).as("getTelemetry");

    cy.visit(`/telemetry/${satelliteId}`);
    cy.wait("@getTelemetry");

    cy.contains("connection refused").should("be.visible");
    cy.contains("Is grpc-server running and reachable").should("be.visible");
  });
});
