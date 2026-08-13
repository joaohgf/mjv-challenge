package adapter

import (
	"errors"
	"testing"
	"time"

	errs "github.com/joaohgf/mjv-challenge/internal/core/error"
	"github.com/joaohgf/mjv-challenge/internal/enum"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestClaimUpdateAssignsLeaseToken(t *testing.T) {
	now := time.Date(2026, time.August, 13, 12, 0, 0, 0, time.UTC)
	update := claimUpdate(now, 15*time.Second, "lease-2")
	set := update["$set"].(bson.M)

	if set["status"] != enum.OutboxProcessing || set["lease_token"] != "lease-2" || !set["locked_until"].(time.Time).Equal(now.Add(15*time.Second)) {
		t.Fatalf("expected processing lease update, got %#v", set)
	}
}

func TestClaimedRequiresCurrentLeaseToken(t *testing.T) {
	filter := claimed("event-1", "lease-2")

	if filter["event_id"] != "event-1" || filter["status"] != enum.OutboxProcessing || filter["lease_token"] != "lease-2" {
		t.Fatalf("expected ownership filter, got %#v", filter)
	}
}

func TestRequireLeaseRejectsLostOrMissingOwnership(t *testing.T) {
	for _, result := range []*mongo.UpdateResult{nil, {}, {MatchedCount: 0}} {
		if !errors.Is(requireLease(result), errs.ErrLeaseLost) {
			t.Fatalf("expected lease loss for %#v", result)
		}
	}
	if err := requireLease(&mongo.UpdateResult{MatchedCount: 1}); err != nil {
		t.Fatalf("expected owned lease, got %v", err)
	}
}

func TestUnlockFieldsRemovesLeaseMetadata(t *testing.T) {
	fields := unlockFields()

	if fields["locked_until"] != "" || fields["lease_token"] != "" {
		t.Fatalf("expected lease fields to be unset, got %#v", fields)
	}
}
