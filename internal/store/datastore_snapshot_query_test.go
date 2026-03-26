package store

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestDataNotifHasStartTimeInWindowStringTimestamp(t *testing.T) {
	windowStart := mustParseRFC3339(t, "2026-03-26T09:00:00Z")
	windowStop := mustParseRFC3339(t, "2026-03-26T09:30:00Z")

	dataNotif := bson.M{
		"upfEventNotifs": []any{
			bson.M{
				"notificationItems": []any{
					bson.M{"startTime": "2026-03-26T09:05:00Z"},
				},
			},
		},
	}

	if !dataNotifHasStartTimeInWindow(dataNotif, windowStart, windowStop) {
		t.Fatal("expected dataNotif to match time window")
	}
}

func TestDataNotifHasStartTimeInWindowPrimitiveDateTime(t *testing.T) {
	windowStart := mustParseRFC3339(t, "2026-03-26T09:00:00Z")
	windowStop := mustParseRFC3339(t, "2026-03-26T09:30:00Z")

	ts := primitive.NewDateTimeFromTime(time.Date(2026, 3, 26, 9, 10, 0, 0, time.UTC))
	dataNotif := bson.M{
		"upfEventNotifs": primitive.A{
			bson.M{
				"notificationItems": primitive.A{
					bson.M{"startTime": ts},
				},
			},
		},
	}

	if !dataNotifHasStartTimeInWindow(dataNotif, windowStart, windowStop) {
		t.Fatal("expected primitive datetime to match time window")
	}
}

func TestDataNotifHasStartTimeInWindowNoMatch(t *testing.T) {
	windowStart := mustParseRFC3339(t, "2026-03-26T09:00:00Z")
	windowStop := mustParseRFC3339(t, "2026-03-26T09:30:00Z")

	dataNotif := bson.M{
		"upfEventNotifs": []any{
			bson.M{
				"notificationItems": []any{
					bson.M{"startTime": "2026-03-26T10:05:00Z"},
				},
			},
		},
	}

	if dataNotifHasStartTimeInWindow(dataNotif, windowStart, windowStop) {
		t.Fatal("expected dataNotif to be outside time window")
	}
}

func mustParseRFC3339(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("failed to parse RFC3339 time %q: %v", value, err)
	}
	return parsed
}
