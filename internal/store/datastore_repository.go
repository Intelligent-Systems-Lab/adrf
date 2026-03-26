package store

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/free5gc/adrf/internal/logger"
	"github.com/free5gc/util/mongoapi"
)

// DataStoreRepository owns Mongo persistence primitives for ADRF store records.
type DataStoreRepository struct {
	dbName string
}

func NewDataStoreRepository(dbName string) *DataStoreRepository {
	return &DataStoreRepository{dbName: dbName}
}

// EnsureIndexes provisions required indexes for store and retrieval paths.
//
// Required indexes:
//  1. unique(storeTransId): retrieval key uniqueness guarantee.
//  2. (supi, ingestedAt): primary access pattern for UE and time-window filters.
//  3. ingestedAt: supporting index for time-based maintenance and scans.
func (r *DataStoreRepository) EnsureIndexes(ctx context.Context) error {
	if r.dbName == "" {
		return errors.New("mongodb database name is empty")
	}
	if mongoapi.Client == nil {
		return errors.New("mongodb client is not initialized")
	}

	models := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "storeTransId", Value: 1}},
			Options: options.Index().
				SetName("idx_store_trans_id").
				SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "supi", Value: 1}, {Key: "ingestedAt", Value: -1}},
			Options: options.Index().
				SetName("idx_supi_ingested_at"),
		},
		{
			Keys: bson.D{{Key: "ingestedAt", Value: -1}},
			Options: options.Index().
				SetName("idx_ingested_at"),
		},
	}

	logger.StoreLog.Infof("Ensuring Mongo indexes for collection: %s", DataStoreRecordsCollectionName)
	_, err := r.collection().Indexes().CreateMany(ctx, models)
	if err != nil {
		return fmt.Errorf("failed to create data store indexes: %w", err)
	}

	logger.StoreLog.Infof("ADRF Mongo indexes ensured for collection: %s", DataStoreRecordsCollectionName)
	return nil
}

// InsertDataStoreRecord persists one store record document into Mongo.
func (r *DataStoreRepository) InsertDataStoreRecord(
	ctx context.Context,
	doc *NadrfDataStoreRecordDocument,
) error {
	if doc == nil {
		return errors.New("data store document is nil")
	}
	if mongoapi.Client == nil {
		return errors.New("mongodb client is not initialized")
	}

	logger.StoreLog.Debugf("Inserting store record: storeTransId=%s supi=%s", doc.StoreTransID, doc.Supi)
	_, err := r.collection().InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to insert data store record: %w", err)
	}

	logger.StoreLog.Infof("Store record inserted: storeTransId=%s", doc.StoreTransID)
	return nil
}

func (r *DataStoreRepository) collection() *mongo.Collection {
	return mongoapi.Client.Database(r.dbName).Collection(DataStoreRecordsCollectionName)
}
