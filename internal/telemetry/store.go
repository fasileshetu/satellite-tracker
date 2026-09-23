// Package telemetry stores and retrieves satellite telemetry readings in
// DynamoDB. Telemetry is a very different access pattern from the
// components data in Postgres: readings are write-heavy, always accessed
// by satellite + recent time range, and never need joins or transactions
// across satellites — a textbook fit for a single-table NoSQL design
// instead of a relational schema.
package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// TableName is the DynamoDB table telemetry readings are stored in.
//
// Partition key: satellite_id (string)   -- one partition per satellite
// Sort key:      timestamp_ms (number)   -- unix millis, so "most recent
//
//	N readings" is a single Query with
//	ScanIndexForward=false and a Limit.
const TableName = "telemetry_readings"

// Reading is the storage-layer representation of one telemetry sample.
// It mirrors the gRPC TelemetryReading message but is independent of the
// generated protobuf types, so the storage layer doesn't need to know
// anything about gRPC.
type Reading struct {
	SatelliteID        string  `dynamodbav:"satellite_id"`
	TimestampMs        int64   `dynamodbav:"timestamp_ms"`
	BatteryVoltage     float64 `dynamodbav:"battery_voltage"`
	TemperatureCelsius float64 `dynamodbav:"temperature_celsius"`
	SignalStrengthDbm  float64 `dynamodbav:"signal_strength_dbm"`
	Status             string  `dynamodbav:"status"`
}

// Store wraps a DynamoDB client scoped to the telemetry table.
type Store struct {
	client    *dynamodb.Client
	tableName string
}

// NewStore builds a Store from an AWS config. Pass cfg from
// config.LoadDefaultConfig(ctx) for real AWS, or one built with a custom
// endpoint resolver for DynamoDB Local during development (see
// NewLocalStore).
func NewStore(client *dynamodb.Client) *Store {
	return &Store{client: client, tableName: TableName}
}

// PutReading writes a single telemetry reading. DynamoDB has no schema to
// migrate — unlike the Postgres `components` table, there's no CREATE
// TABLE step baked into the app; the table itself is provisioned once
// (see EnsureTable, and terraform/dynamodb.tf for the real AWS version).
func (s *Store) PutReading(ctx context.Context, r Reading) error {
	item, err := attributevalue.MarshalMap(r)
	if err != nil {
		return fmt.Errorf("marshal reading: %w", err)
	}
	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("put item: %w", err)
	}
	return nil
}

// RecentReadings returns up to `limit` readings for a satellite, most
// recent first. This is a single-partition Query (cheap, fast) rather
// than a Scan (which would read the whole table) — the access pattern
// the table's key schema was designed around.
func (s *Store) RecentReadings(ctx context.Context, satelliteID string, limit int32) ([]Reading, error) {
	if limit <= 0 {
		limit = 50
	}
	out, err := s.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		KeyConditionExpression: aws.String("satellite_id = :sid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":sid": &types.AttributeValueMemberS{Value: satelliteID},
		},
		ScanIndexForward: aws.Bool(false), // descending by sort key (timestamp)
		Limit:            aws.Int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("query readings: %w", err)
	}

	readings := make([]Reading, 0, len(out.Items))
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &readings); err != nil {
		return nil, fmt.Errorf("unmarshal readings: %w", err)
	}
	return readings, nil
}

// EnsureTable creates the telemetry table if it doesn't already exist.
// Called once at startup for local development against DynamoDB Local;
// in real AWS the table is provisioned by Terraform instead (see
// terraform/dynamodb.tf) and this is a harmless no-op check.
func (s *Store) EnsureTable(ctx context.Context) error {
	_, err := s.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(s.tableName),
	})
	if err == nil {
		return nil // already exists
	}

	_, err = s.client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(s.tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("satellite_id"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("timestamp_ms"), AttributeType: types.ScalarAttributeTypeN},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("satellite_id"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("timestamp_ms"), KeyType: types.KeyTypeRange},
		},
		BillingMode: types.BillingModePayPerRequest, // no capacity planning needed for a portfolio-scale table
	})
	if err != nil {
		return fmt.Errorf("create table: %w", err)
	}

	waiter := dynamodb.NewTableExistsWaiter(s.client)
	return waiter.Wait(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(s.tableName)}, 30*time.Second)
}
