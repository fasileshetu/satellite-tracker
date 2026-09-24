// End-to-end coverage of the components workflow (roadmap item 8).
//
// Targets the dashboard running against local docker-compose (Postgres +
// API), the same setup documented in the top-level README's "Run with
// Docker Compose" section. With no Cognito env vars set, the API runs
// with no auth (see internal/api/router.go), so the dashboard's own
// canWrite check (ComponentsPanel.tsx) allows writes without logging in
// -- these tests exercise the same unauthenticated local-dev path a
// contributor would use day to day, not the Cognito login flow itself
// (that needs a real browser redirect to Cognito's Hosted UI, which is
// out of scope for this suite; it's already been verified manually
// against real AWS -- see README).
//
// Each test uses a satellite ID unique to that test run (timestamp-based)
// so re-running the suite against a database that already has data from
// previous runs never produces false failures from stale rows.

function uniqueSatelliteId() {
  return `CY-${Date.now()}`;
}

describe("components workflow", () => {
  beforeEach(() => {
    cy.visit("/");
  });

  it("loads the dashboard shell", () => {
    cy.contains("h1", "satellite-tracker").should("be.visible");
    cy.contains("h2", "Register a component").should("be.visible");
    cy.contains("h2", "Components").should("be.visible");
  });

  it("registers a new component and lists it", () => {
    const satelliteId = uniqueSatelliteId();

    cy.get('[data-cy="satellite-id-input"]').type(satelliteId);
    cy.get('[data-cy="name-input"]').type("Star Tracker Assembly");
    cy.get('[data-cy="part-number-input"]').type("ST-900");
    cy.get('[data-cy="submit-button"]').click();

    cy.get(`[data-cy="component-row"][data-satellite-id="${satelliteId}"]`)
      .should("be.visible")
      .within(() => {
        cy.contains("Star Tracker Assembly");
        cy.contains("ST-900");
        cy.get('[data-cy="status-badge"]').should("contain.text", "received");
      });

    // the form clears after a successful submit, ready for the next entry
    cy.get('[data-cy="satellite-id-input"]').should("have.value", "");
  });

  it("filters the components list by satellite ID", () => {
    const satelliteId = uniqueSatelliteId();

    cy.get('[data-cy="satellite-id-input"]').type(satelliteId);
    cy.get('[data-cy="name-input"]').type("Reaction Wheel");
    cy.get('[data-cy="part-number-input"]').type("RW-4");
    cy.get('[data-cy="submit-button"]').click();
    cy.get(`[data-cy="component-row"][data-satellite-id="${satelliteId}"]`).should("exist");

    // an unrelated filter shouldn't show this row
    cy.get('[data-cy="satellite-filter-input"]').type("no-such-satellite-xyz");
    cy.get(`[data-cy="component-row"][data-satellite-id="${satelliteId}"]`).should("not.exist");
    cy.get('[data-cy="empty-state"]').should("be.visible");

    // filtering by its own ID brings it back, and only it
    cy.get('[data-cy="satellite-filter-input"]').clear().type(satelliteId);
    cy.get('[data-cy="component-row"]').should("have.length", 1);
    cy.get(`[data-cy="component-row"][data-satellite-id="${satelliteId}"]`).should("exist");
  });

  it("moves a component through its status lifecycle", () => {
    const satelliteId = uniqueSatelliteId();

    cy.get('[data-cy="satellite-id-input"]').type(satelliteId);
    cy.get('[data-cy="name-input"]').type("Avionics Board");
    cy.get('[data-cy="part-number-input"]').type("AV-2201");
    cy.get('[data-cy="submit-button"]').click();

    const row = () => cy.get(`[data-cy="component-row"][data-satellite-id="${satelliteId}"]`);

    row().find('[data-cy="status-badge"]').should("have.text", "received");

    row().find('[data-cy="status-select"]').select("in_test");
    row().find('[data-cy="status-badge"]').should("have.text", "in_test");

    row().find('[data-cy="status-select"]').select("pass");
    row().find('[data-cy="status-badge"]').should("have.text", "pass");

    row().find('[data-cy="status-select"]').select("flight_ready");
    row().find('[data-cy="status-badge"]').should("have.text", "flight_ready");
  });

  it("navigates to a satellite's telemetry page from the components list", () => {
    const satelliteId = uniqueSatelliteId();

    cy.get('[data-cy="satellite-id-input"]').type(satelliteId);
    cy.get('[data-cy="name-input"]').type("Signal Amplifier");
    cy.get('[data-cy="part-number-input"]').type("SA-7");
    cy.get('[data-cy="submit-button"]').click();

    cy.get(`[data-cy="component-row"][data-satellite-id="${satelliteId}"]`)
      .find('[data-cy="telemetry-link"]')
      .click();

    cy.url().should("include", `/telemetry/${satelliteId}`);
    cy.contains("h2", `Telemetry -- ${satelliteId}`).should("be.visible");

    // grpc-server may or may not be running with data for a satellite
    // that was just created -- either a populated table or the explicit
    // empty-state message is a correct outcome; a stuck "Loading..." or
    // an uncaught page crash is not.
    cy.contains(/No telemetry readings yet\.|BATTERY \(V\)/, { timeout: 10000 }).should(
      "be.visible"
    );
  });
});
