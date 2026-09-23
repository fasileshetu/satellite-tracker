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

	"google.golang.org/grpc"
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
		if err := store.EnsureTable(ctx); err != nil {
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

	log.Printf("telemetry gRPC server listening on :%s", port)
	if err := grpcSrv.Serve(lis); err != nil {
		log.Fatalf("grpc server failed: %v", err)
	}
}
