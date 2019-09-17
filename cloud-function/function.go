// Package cloudfunction contains a Pub/Sub Cloud Function.
package cloudfunction

import (
	"context"
	"errors"
	"os"

	snapshot "github.com/n0madic/gce-snapshot"
)

// PubSubMessage is the payload of a Pub/Sub event.
type PubSubMessage struct {
	Days  int `json:"days"`
	Month int `json:"month"`
}

// SnapshotPubSub consumes a Pub/Sub message.
func SnapshotPubSub(ctx context.Context, m PubSubMessage) error {
	if m.Days < 1 || m.Month < 1 {
		return errors.New("Snapshot cleaning periods not set")
	}
	project := os.Getenv("GCP_PROJECT")
	snapshot.Create(ctx, project)
	snapshot.Purge(ctx, project, m.Days, m.Month)
	return nil
}
