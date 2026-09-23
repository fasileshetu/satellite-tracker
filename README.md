# satellite-tracker

Internal tool for tracking satellite components through manufacturing and test,
from raw part to flight-ready status. Built as a portfolio project mapped to
the K2 Space Application Software Engineer qualifications: Go, REST/gRPC,
PostgreSQL, DynamoDB, Docker, Kubernetes, Terraform, CI/CD, OAuth, AWS.

## Status: Step 2 — running on local Kubernetes

Currently implemented:
- `GET /health`
- `POST /components` — register a new component
- `GET /components` — list components, optional `?satellite_id=` filter
- `GET /components/{id}` — fetch one
- `PATCH /components/{id}/status` — move it through received → in_test → pass/fail → flight_ready

## Run with Docker Compose (simplest)

```bash
docker compose up --build
```

Starts Postgres and the API together on `localhost:8080`.

## Run on local Kubernetes (minikube)

This deploys the same containers, but through real Kubernetes objects:
a Deployment and Service for Postgres, a Deployment and Service for the API,
plus a ConfigMap, a Secret, and a PersistentVolumeClaim.

### 1. Start minikube

```bash
minikube start
```

### 2. Build the API image and load it into minikube

minikube runs its own Docker daemon, separate from your normal `docker`
command, so an image built normally isn't visible to it. `minikube image load`
copies it in.

```bash
docker build -t satellite-tracker-api:local .
minikube image load satellite-tracker-api:local
```

### 3. Apply the manifests

```bash
kubectl apply -f k8s/
```

This creates everything: namespace, secret, configmaps, PVC, both deployments,
both services.

### 4. Watch it come up

```bash
kubectl get pods -n satellite-tracker -w
```

Wait until both `postgres-...` and `api-...` pods show `Running` and `1/1`
ready. Postgres needs to pass its readiness probe before the API's first
requests will succeed, though the API pod itself will still start.

### 5. Reach the API

```bash
minikube service api -n satellite-tracker --url
```

This prints a URL. Use it the same way as before:

```bash
curl -X POST $(minikube service api -n satellite-tracker --url)/components \
  -H "Content-Type: application/json" \
  -d '{"satellite_id":"K2-GRAVITAS-2","name":"Avionics Board Rev C","part_number":"AV-2201"}'
```

### 6. Useful commands while you're learning

```bash
kubectl get all -n satellite-tracker        # see every object at once
kubectl describe pod <pod-name> -n satellite-tracker   # why is it not starting?
kubectl logs <pod-name> -n satellite-tracker           # app output
kubectl logs <pod-name> -n satellite-tracker -f        # follow logs live
```

### 7. Tear it down

```bash
kubectl delete namespace satellite-tracker
```

Deleting the namespace deletes everything inside it in one shot.

## What each manifest does

| File | Kind | Purpose |
|---|---|---|
| `00-namespace.yaml` | Namespace | Keeps everything grouped and isolated from other things in the cluster |
| `01-postgres-configmap.yaml` | ConfigMap | Holds the schema SQL, mounted into Postgres's init directory |
| `02-postgres-secret.yaml` | Secret | Postgres credentials, kept separate from plain config |
| `03-postgres-pvc.yaml` | PersistentVolumeClaim | Requests disk storage that survives pod restarts |
| `04-postgres-deployment.yaml` | Deployment | Runs the Postgres pod, mounts the volume and the init SQL |
| `05-postgres-service.yaml` | Service (ClusterIP) | Internal-only DNS name so the API can reach Postgres |
| `06-api-configmap.yaml` | ConfigMap | Non-secret API settings, `PORT` and `DATABASE_URL` |
| `07-api-deployment.yaml` | Deployment | Runs 2 replicas of the API, with health checks and resource limits |
| `08-api-service.yaml` | Service (NodePort) | Exposes the API outside the cluster for local testing |

## Run without any of this

Requires Go 1.22+ and a running Postgres instance.

```bash
go mod tidy
psql "$DATABASE_URL" -f migrations/001_init.sql
go run ./cmd/api
```

## OAuth 2.0 / OIDC (Cognito)

`terraform/cognito.tf` provisions an AWS Cognito User Pool as the OIDC
identity provider. `internal/auth/middleware.go` is the resource-server
side: it validates the `Authorization: Bearer <token>` header on
`POST /components` and `PATCH /components/{id}/status` against Cognito's
JWKS (its public signing keys), checking the token's signature, issuer,
expiry, `token_use`, and `client_id`. `GET` endpoints and `/health` stay
open.

The API only enables this when `OIDC_ISSUER_URL` and `OIDC_CLIENT_ID` are
both set (see `k8s/aws/api-configmap.yaml`) — left unset, e.g. for local
`docker-compose` dev, it falls back to no auth at all.

### Getting a real token to test with

No frontend exists yet, so the fastest way to get a real Cognito-issued
access token is the AWS CLI, using `USER_PASSWORD_AUTH` (enabled in
`cognito.tf` for exactly this purpose):

```bash
# one-time: create a test user and set a permanent password
aws cognito-idp admin-create-user \
  --user-pool-id $(terraform -chdir=terraform output -raw cognito_user_pool_id) \
  --username test@example.com \
  --message-action SUPPRESS

aws cognito-idp admin-set-user-password \
  --user-pool-id $(terraform -chdir=terraform output -raw cognito_user_pool_id) \
  --username test@example.com \
  --password 'TempPass123!' \
  --permanent

# get an access token
aws cognito-idp initiate-auth \
  --auth-flow USER_PASSWORD_AUTH \
  --client-id $(terraform -chdir=terraform output -raw cognito_app_client_id) \
  --auth-parameters USERNAME=test@example.com,PASSWORD='TempPass123!'
```

That last command's JSON response has an `AccessToken` field. Use it:

```bash
curl -X POST localhost:8080/components \
  -H "Authorization: Bearer <AccessToken value>" \
  -d '{"satellite_id":"sat-01","name":"Star Tracker","part_number":"ST-100"}'

# without the header, or with a bad/expired token:
curl -X POST localhost:8080/components -d '{...}'   # -> 401
```

A real frontend would instead redirect the user to Cognito's Hosted UI
(`terraform output cognito_hosted_ui_url`) and exchange the returned
`code` for tokens — the Authorization Code flow the `allowed_oauth_flows`
setting in `cognito.tf` supports. The CLI shortcut above exercises the
exact same token-validation code path without needing that frontend
built first.

## Dashboard (Next.js + TypeScript)

`web/` is a small App Router dashboard that talks to both backends:

- **Components** — lists/filters `GET /components`, registers new ones
  (`POST /components`), and moves a component through its status lifecycle
  (`PATCH /components/{id}/status`).
- **Login** — a real Authorization Code + PKCE flow against Cognito's
  Hosted UI (`/callback` exchanges the code for tokens via a server-side
  route, `app/api/auth/token`, so the token endpoint never needs to allow
  the browser's origin directly). Logged-out visitors can browse; the two
  write endpoints require a token, same as the API enforces server-side.
- **Telemetry** — clicking a satellite ID opens `/telemetry/[satelliteId]`,
  which calls the dashboard's own `/api/telemetry/[satelliteId]` route.
  That route is a small gRPC client (`@grpc/grpc-js` + `@grpc/proto-loader`,
  loading `web/proto/telemetry.proto` directly — no generated stubs needed
  in Node) that calls `grpc-server`'s `GetTelemetryHistory` RPC and returns
  plain JSON. This exists because browsers can't speak gRPC's HTTP/2
  framing directly; the dashboard's backend is the bridge, the same
  backend-for-frontend pattern a real gRPC-web setup would use.

### Running it locally

```bash
cd web
npm install
cp .env.local.example .env.local   # fill in Cognito values if testing login
npm run dev
```

Open `http://localhost:3000`. With no Cognito env vars set, the dashboard
still works for browsing (`GET` routes have no auth) — it just disables
the login button. Point `NEXT_PUBLIC_API_URL` / `GRPC_SERVER_ADDR` at
wherever the API and grpc-server are actually running (docker-compose,
minikube via `kubectl port-forward`, or the real EKS services).

The API now sends CORS headers (`internal/api/cors.go`) so the dashboard
can call it from a different origin; `ALLOWED_ORIGIN` controls which
origin is allowed (defaults to `*`, fine for a portfolio project — a real
deployment would pin it to the dashboard's actual URL).

## Roadmap (see project plan)

1. ✅ Go REST API + Postgres, containerized
2. ✅ Local Kubernetes (minikube): Deployment, Service, ConfigMap, Secret, PVC
3. ✅ Terraform for real AWS infra: EKS, RDS
4. ✅ gRPC telemetry ingestion service + DynamoDB (via IRSA)
5. ✅ OAuth 2.0 / OIDC login (Cognito, resource-server-side validation)
6. ✅ Next.js + TypeScript dashboard (components CRUD, Cognito login, gRPC-backed telemetry view)
7. GitHub Actions CI/CD: test → build → push to CodeArtifact → deploy to EKS
8. Cypress E2E tests
