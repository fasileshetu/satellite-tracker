# satellite-tracker

Internal tool for tracking satellite components through manufacturing and test,
from raw part to flight-ready status. Built as a portfolio project mapped to
the K2 Space Application Software Engineer qualifications: Go, REST/gRPC,
PostgreSQL, DynamoDB, Docker, Kubernetes, Terraform, CI/CD, OAuth, AWS.

## Status: Step 1 — REST API + Postgres, containerized

Currently implemented:
- `GET /health`
- `POST /components` — register a new component
- `GET /components` — list components, optional `?satellite_id=` filter
- `GET /components/{id}` — fetch one
- `PATCH /components/{id}/status` — move it through received → in_test → pass/fail → flight_ready

## Run locally with Docker Compose

```bash
docker compose up --build
```

This starts Postgres (with the schema in `migrations/001_init.sql` applied
automatically on first boot) and the API on `localhost:8080`.

Try it:

```bash
curl -X POST localhost:8080/components \
  -H "Content-Type: application/json" \
  -d '{"satellite_id":"K2-GRAVITAS-2","name":"Avionics Board Rev C","part_number":"AV-2201"}'

curl localhost:8080/components

curl -X PATCH localhost:8080/components/1/status \
  -H "Content-Type: application/json" \
  -d '{"status":"in_test"}'
```

## Run without Docker

Requires Go 1.22+ and a running Postgres instance.

```bash
go mod tidy
psql "$DATABASE_URL" -f migrations/001_init.sql
go run ./cmd/api
```

## Roadmap (see project plan)

1. ✅ Go REST API + Postgres, containerized
2. Deploy to local Kubernetes (minikube/kind): Deployment, Service, ConfigMap
3. Terraform for real AWS infra: EKS, RDS
4. gRPC telemetry ingestion service + DynamoDB
5. OAuth 2.0 / OIDC login
6. Next.js + TypeScript dashboard
7. GitHub Actions CI/CD: test → build → push to CodeArtifact → deploy to EKS
8. Cypress E2E tests
