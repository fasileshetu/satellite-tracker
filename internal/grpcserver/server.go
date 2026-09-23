// Package grpcserver implements the TelemetryService gRPC API defined in
// proto/telemetry.proto, backed by DynamoDB.
package grpcserver

import (
	"context"
	"io"
	"log"

	"google.golang.org/protobuf/types/known/timestamppb"

	"satellite-tracker/internal/telemetry"
	pb "satellite-tracker/internal/telemetrypb"
)

// Server implements pb.TelemetryServiceServer.
type Server struct {
	pb.UnimplementedTelemetryServiceServer
	store *telemetry.Store
}

func New(store *telemetry.Store) *Server {
	return &Server{store: store}
}

// IngestTelemetry receives a stream of readings from a single caller
// (a satellite ground-station relay, or in this project a load-test
// client) and writes each one to DynamoDB as it arrives. Streaming
// keeps one connection open for many readings instead of paying a new
// TCP + TLS + HTTP handshake per reading, which matters when a ground
// station is pushing telemetry every few seconds for hours at a time.
func (s *Server) IngestTelemetry(stream pb.TelemetryService_IngestTelemetryServer) error {
	ctx := stream.Context()
	var count int32
	var satelliteID string

	for {
		reading, err := stream.Recv()
		if err == io.EOF {
			// Client closed its send side -- send the final summary and return.
			return stream.SendAndClose(&pb.IngestSummary{
				ReadingsReceived: count,
				SatelliteId:      satelliteID,
			})
		}
		if err != nil {
			return err
		}

		satelliteID = reading.GetSatelliteId()

		record := telemetry.Reading{
			SatelliteID:        reading.GetSatelliteId(),
			TimestampMs:        toMillis(reading.GetTimestamp()),
			BatteryVoltage:     reading.GetBatteryVoltage(),
			TemperatureCelsius: reading.GetTemperatureCelsius(),
			SignalStrengthDbm:  reading.GetSignalStrengthDbm(),
			Status:             reading.GetStatus(),
		}

		if err := s.store.PutReading(ctx, record); err != nil {
			log.Printf("failed to store reading for %s: %v", record.SatelliteID, err)
			continue // don't kill the whole stream over one bad write
		}
		count++
	}
}

// GetTelemetryHistory returns the most recent readings for one satellite.
func (s *Server) GetTelemetryHistory(ctx context.Context, req *pb.TelemetryHistoryRequest) (*pb.TelemetryHistoryResponse, error) {
	readings, err := s.store.RecentReadings(ctx, req.GetSatelliteId(), req.GetLimit())
	if err != nil {
		return nil, err
	}

	resp := &pb.TelemetryHistoryResponse{
		Readings: make([]*pb.TelemetryReading, 0, len(readings)),
	}
	for _, r := range readings {
		resp.Readings = append(resp.Readings, &pb.TelemetryReading{
			SatelliteId:        r.SatelliteID,
			Timestamp:          timestamppb.New(fromMillis(r.TimestampMs)),
			BatteryVoltage:     r.BatteryVoltage,
			TemperatureCelsius: r.TemperatureCelsius,
			SignalStrengthDbm:  r.SignalStrengthDbm,
			Status:             r.Status,
		})
	}
	return resp, nil
}
