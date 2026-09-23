package grpcserver

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// toMillis converts a protobuf Timestamp to unix milliseconds, the sort
// key type used in DynamoDB. A nil timestamp (client didn't set one)
// falls back to "now" rather than storing a zero-value reading.
func toMillis(ts *timestamppb.Timestamp) int64 {
	if ts == nil {
		return time.Now().UnixMilli()
	}
	return ts.AsTime().UnixMilli()
}

func fromMillis(ms int64) time.Time {
	return time.UnixMilli(ms)
}
