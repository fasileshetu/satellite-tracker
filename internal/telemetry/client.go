package telemetry

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// NewClient builds a DynamoDB client.
//
//   - endpoint == ""  -> real AWS DynamoDB, credentials/region come from
//     the normal AWS credential chain (env vars, ~/.aws/credentials, or
//     the pod's IAM role when running in EKS).
//   - endpoint != ""  -> DynamoDB Local (docker-compose), which doesn't
//     check credentials at all but the SDK still requires *some* static
//     values to be present, so we pass dummy ones.
func NewClient(ctx context.Context, region, endpoint string) (*dynamodb.Client, error) {
	if endpoint == "" {
		cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
		if err != nil {
			return nil, fmt.Errorf("load AWS config: %w", err)
		}
		return dynamodb.NewFromConfig(cfg), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("local", "local", "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load AWS config for local endpoint: %w", err)
	}

	return dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	}), nil
}
