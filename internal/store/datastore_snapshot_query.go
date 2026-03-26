package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/free5gc/adrf/internal/logger"
	"github.com/free5gc/util/mongoapi"
)

// ListStoreTransIDsBySnapshot resolves fetch correlation IDs for one retrieval
// subscription snapshot.
//
// Snapshot rules:
//  1. UE scope: `supi` must match the retrieval subscription SUPI.
//  2. Snapshot boundary: only records ingested no later than `snapshotCutoff`.
//  3. Time window: at least one `notificationItems[*].startTime` must be inside
//     [windowStart, windowStop].
//
// The result order is deterministic (`ingestedAt` asc, `_id` asc) so callback
// and fetch behavior are repeatable across retries.
func (r *DataStoreRepository) ListStoreTransIDsBySnapshot(
	ctx context.Context,
	supi string,
	snapshotCutoff time.Time,
	windowStart time.Time,
	windowStop time.Time,
) ([]string, error) {
	supi = strings.TrimSpace(supi)
	if supi == "" {
		return nil, errors.New("supi is empty")
	}
	if snapshotCutoff.IsZero() {
		return nil, errors.New("snapshot cutoff is zero")
	}
	if windowStart.IsZero() || windowStop.IsZero() {
		return nil, errors.New("time window is incomplete")
	}
	if windowStart.After(windowStop) {
		return nil, errors.New("time window start is later than stop")
	}
	if mongoapi.Client == nil {
		return nil, errors.New("mongodb client is not initialized")
	}

	filter := bson.M{
		"supi":       supi,
		"ingestedAt": bson.M{"$lte": snapshotCutoff},
	}
	findOpts := options.Find().
		SetProjection(bson.M{
			"storeTransId": 1,
			"dataNotif":    1,
			"ingestedAt":   1,
		}).
		SetSort(bson.D{{Key: "ingestedAt", Value: 1}, {Key: "_id", Value: 1}})

	cursor, err := r.collection().Find(ctx, filter, findOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to query snapshot candidates: %w", err)
	}
	defer func() {
		if closeErr := cursor.Close(ctx); closeErr != nil {
			logger.StoreLog.Warnf("Failed to close snapshot candidate cursor: %v", closeErr)
		}
	}()

	matchedIDs := make([]string, 0, 64)
	candidateCount := 0
	for cursor.Next(ctx) {
		candidateCount++

		var doc snapshotCandidateDocument
		decodeErr := cursor.Decode(&doc)
		if decodeErr != nil {
			return nil, fmt.Errorf("failed to decode snapshot candidate: %w", decodeErr)
		}

		if doc.StoreTransID == "" {
			continue
		}
		if !dataNotifHasStartTimeInWindow(doc.DataNotif, windowStart, windowStop) {
			continue
		}

		matchedIDs = append(matchedIDs, doc.StoreTransID)
	}
	cursorErr := cursor.Err()
	if cursorErr != nil {
		return nil, fmt.Errorf("snapshot candidate cursor error: %w", cursorErr)
	}

	logger.StoreLog.Infof(
		"Snapshot query resolved: supi=%s candidates=%d matched=%d windowStart=%s windowStop=%s snapshotCutoff=%s",
		supi,
		candidateCount,
		len(matchedIDs),
		windowStart.Format(time.RFC3339Nano),
		windowStop.Format(time.RFC3339Nano),
		snapshotCutoff.Format(time.RFC3339Nano),
	)
	return matchedIDs, nil
}

type snapshotCandidateDocument struct {
	StoreTransID string `bson:"storeTransId"`
	DataNotif    bson.M `bson:"dataNotif"`
}

func dataNotifHasStartTimeInWindow(dataNotif bson.M, windowStart time.Time, windowStop time.Time) bool {
	if len(dataNotif) == 0 {
		return false
	}

	upfEventNotifsRaw, foundUpfEventNotifs := mapValue(dataNotif, "upfEventNotifs")
	if !foundUpfEventNotifs {
		return false
	}

	for _, upfEventNotifAny := range toAnySlice(upfEventNotifsRaw) {
		upfEventNotif, parsedUpfEventNotif := toAnyMap(upfEventNotifAny)
		if !parsedUpfEventNotif {
			continue
		}

		notificationItemsRaw, foundNotificationItems := mapValue(upfEventNotif, "notificationItems")
		if !foundNotificationItems {
			continue
		}
		for _, notificationItemAny := range toAnySlice(notificationItemsRaw) {
			notificationItem, parsedNotificationItem := toAnyMap(notificationItemAny)
			if !parsedNotificationItem {
				continue
			}

			startTimeRaw, foundStartTime := mapValue(notificationItem, "startTime")
			if !foundStartTime {
				continue
			}
			startTime, parsedStartTime := parseTimestamp(startTimeRaw)
			if !parsedStartTime {
				continue
			}

			startTime = startTime.UTC()
			if !startTime.Before(windowStart) && !startTime.After(windowStop) {
				return true
			}
		}
	}

	return false
}

func mapValue(anyMap map[string]any, key string) (any, bool) {
	if anyMap == nil {
		return nil, false
	}
	value, ok := anyMap[key]
	if !ok {
		return nil, false
	}
	return value, true
}

func toAnyMap(raw any) (map[string]any, bool) {
	switch value := raw.(type) {
	case map[string]any:
		return value, true
	case bson.M:
		return map[string]any(value), true
	default:
		return nil, false
	}
}

func toAnySlice(raw any) []any {
	switch value := raw.(type) {
	case []any:
		return value
	case primitive.A:
		return []any(value)
	default:
		return nil
	}
}

func parseTimestamp(raw any) (time.Time, bool) {
	switch value := raw.(type) {
	case time.Time:
		return value, true
	case *time.Time:
		if value == nil {
			return time.Time{}, false
		}
		return *value, true
	case primitive.DateTime:
		return value.Time(), true
	case string:
		if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
			return parsed, true
		}
		if parsed, err := time.Parse(time.RFC3339, value); err == nil {
			return parsed, true
		}
		return time.Time{}, false
	default:
		return time.Time{}, false
	}
}
