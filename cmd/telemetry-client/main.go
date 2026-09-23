// Command telemetry-client is a small demo/load-test client: it opens a
// streaming connection to the telemetry gRPC server and pushes a batch
// of simulated readings, then queries the history back. Useful for
// demoing the gRPC service without needing a real satellite ground
// station, and for the interview walkthrough.
package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "satellite-tracker/internal/telemetrypb"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	addr := getenv("GRPC_ADDR", "localhost:9090")
	satelliteID := getenv("SATELLITE_ID", "sat-01")

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to %s: %v", addr, err)
	}
	defer conn.Close()

	client := pb.NewTelemetryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := client.IngestTelemetry(ctx)
	if err != nil {
		log.Fatalf("failed to open ingest stream: %v", err)
	}

	const numReadings = 20
	for i := 0; i < numReadings; i++ {
		reading := &pb.TelemetryReading{
			SatelliteId:        satelliteID,
			Timestamp:          timestamppb.New(time.Now()),
			BatteryVoltage:     11.5 + rand.Float64(),
			TemperatureCelsius: 18 + rand.Float64()*6,
			SignalStrengthDbm:  -70 + rand.Float64()*10,
			Status:             "nominal",
		}
		if err := stream.Send(reading); err != nil {
			log.Fatalf("failed to send reading %d: %v", i, err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	summary, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("failed to close stream: %v", err)
	}
	log.Printf("ingested %d readings for %s", summary.GetReadingsReceived(), summary.GetSatelliteId())

	history, err := client.GetTelemetryHistory(context.Background(), &pb.TelemetryHistoryRequest{
		SatelliteId: satelliteID,
		Limit:       5,
	})
	if err != nil {
		log.Fatalf("failed to fetch history: %v", err)
	}
	log.Printf("most recent %d readings for %s:", len(history.GetReadings()), satelliteID)
	for _, r := range history.GetReadings() {
		log.Printf("  %s  battery=%.2fV  temp=%.1fC  signal=%.1fdBm  status=%s",
			r.GetTimestamp().AsTime().Format(time.RFC3339),
			r.GetBatteryVoltage(), r.GetTemperatureCelsius(), r.GetSignalStrengthDbm(), r.GetStatus())
	}
}
