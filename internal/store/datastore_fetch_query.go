package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/free5gc/adrf/internal/logger"
	"github.com/free5gc/util/mongoapi"
)

// GetDataStoreRecordByStoreTransID reads one store record by store transaction
// identifier.
//
// The lookup key is the same identifier distributed in RetrievalNotify
// fetch instructions. Returning mongo.ErrNoDocuments is intentional so the
// processor can map missing records to HTTP 204 per TS 29.575 behavior.
func (r *DataStoreRepository) GetDataStoreRecordByStoreTransID(
	ctx context.Context,
	storeTransID string,
) (*NadrfDataStoreRecordDocument, error) {
	storeTransID = strings.TrimSpace(storeTransID)
	if storeTransID == "" {
		return nil, errors.New("store transaction identifier is empty")
	}
	if mongoapi.Client == nil {
		return nil, errors.New("mongodb client is not initialized")
	}

	filter := bson.M{"storeTransId": storeTransID}
	var doc NadrfDataStoreRecordDocument
	findErr := r.collection().FindOne(ctx, filter).Decode(&doc)
	if findErr != nil {
		if errors.Is(findErr, mongo.ErrNoDocuments) {
			return nil, mongo.ErrNoDocuments
		}
		return nil, fmt.Errorf("failed to query data store record: %w", findErr)
	}

	logger.StoreLog.Debugf("Store record fetched: storeTransId=%s", storeTransID)
	return &doc, nil
}
