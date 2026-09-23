// Command grpc-server runs the telemetry ingestion service: a gRPC server
// backed by DynamoDB, separate from the REST API (cmd/api) which is
// backed by Postgres. Two services, two databases, one codebase -- this
// is the "polyglot persistence" half of the project: structured
// relational data (components) in Postgres, high-volume time-series
// writes (telemetry) in DynamoDB.
package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"satellite-tracker/internal/grpcserver"
	"satellite-tracker/internal/telemetry"
	pb "satellite-tracker/internal/telemetrypb"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	ctx := context.Background()

	port := getenv("GRPC_PORT", "9090")
	region := getenv("AWS_REGION", "us-west-2")
	// DYNAMODB_ENDPOINT is set for local dev (docker-compose points this
	// at DynamoDB Local); left empty in EKS, where the SDK talks to real
	// AWS DynamoDB using the pod's IAM permissions.
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")

	client, err := telemetry.NewClient(ctx, region, endpoint)
	if err != nil {
		log.Fatalf("failed to build DynamoDB client: %v", err)
	}

	store := telemetry.NewStore(client)

	// Only auto-create the table for local dev. In real AWS the table is
	// provisioned by Terraform (terraform/dynamodb.tf) -- an app process
	// shouldn't be the thing creating production infrastructure.
	if endpoint != "" {
		if err := ensureTableWithRetry(ctx, store); err != nil {
			log.Fatalf("failed to ensure telemetry table exists: %v", err)
		}
		log.Println("DynamoDB Local table ready")
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", port, err)
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterTelemetryServiceServer(grpcSrv, grpcserver.New(store))
	reflection.Register(grpcSrv) // lets grpcurl/evans introspect the service without the .proto file

	// Kubernetes' native grpc: readiness/liveness probes (k8s/aws/grpc-server-deployment.yaml)
	// speak the standard gRPC health-checking protocol, not HTTP -- this
	// is what they're actually calling.
	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcSrv, healthSrv)

	log.Printf("telemetry gRPC server listening on :%s", port)
	if err := grpcSrv.Serve(lis); err != nil {
		log.Fatalf("grpc server failed: %v", err)
	}
}

// ensureTableWithRetry retries EnsureTable with backoff. DynamoDB Local
// starts as a separate container that docker-compose brings up in
// parallel -- by the time this process starts, the container exists but
// the JVM inside it may not have finished booting and bound port 8000
// yet. depends_on only guarantees container start order, not "ready to
// accept connections," so the app itself has to tolerate that gap.
func ensureTableWithRetry(ctx context.Context, store *telemetry.Store) error {
	const maxAttempts = 10
	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err = store.EnsureTable(ctx); err == nil {
			return nil
		}
		log.Printf("DynamoDB not ready yet (attempt %d/%d): %v", attempt, maxAttempts, err)
		time.Sleep(2 * time.Second)
	}
	return err
}
