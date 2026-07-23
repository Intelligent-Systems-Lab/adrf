package store

import (
	"context"
	"errors"
	"fmt"
	"time"

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
		{
			Keys: bson.D{{Key: "expiryTime", Value: 1}},
			Options: options.Index().
				SetName("idx_expiry_time"),
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

	logger.StoreLog.Debugf("Store record inserted: storeTransId=%s", doc.StoreTransID)
	return nil
}

// SearchRecordsByFilter performs paginated timeWindow/supi searches.
func (r *DataStoreRepository) SearchRecordsByFilter(
	ctx context.Context,
	supi string,
	startTime, endTime *time.Time,
	limit, offset int64,
) ([]*NadrfDataStoreRecordDocument, int64, error) {
	if mongoapi.Client == nil {
		return nil, 0, errors.New("mongodb client is not initialized")
	}

	filter := bson.M{}
	if supi != "" {
		filter["supi"] = supi
	}
	if startTime != nil || endTime != nil {
		timeFilter := bson.M{}
		if startTime != nil {
			timeFilter["$gte"] = *startTime
		}
		if endTime != nil {
			timeFilter["$lte"] = *endTime
		}
		filter["ingestedAt"] = timeFilter
	}

	total, err := r.collection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count data store records: %w", err)
	}

	findOpts := options.Find().SetSort(bson.D{{Key: "ingestedAt", Value: -1}})
	if limit > 0 {
		findOpts.SetLimit(limit)
	}
	if offset > 0 {
		findOpts.SetSkip(offset)
	}

	cursor, err := r.collection().Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search data store records: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []*NadrfDataStoreRecordDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, 0, fmt.Errorf("failed to decode data store records: %w", err)
	}
	return docs, total, nil
}

// DeleteExpiredRecords purges records whose expiryTime is past.
func (r *DataStoreRepository) DeleteExpiredRecords(ctx context.Context, now time.Time) (int64, error) {
	if mongoapi.Client == nil {
		return 0, errors.New("mongodb client is not initialized")
	}

	filter := bson.M{
		"expiryTime": bson.M{"$gt": time.Time{}, "$lte": now},
	}
	res, err := r.collection().DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired data store records: %w", err)
	}
	return res.DeletedCount, nil
}

func (r *DataStoreRepository) collection() *mongo.Collection {
	return mongoapi.Client.Database(r.dbName).Collection(DataStoreRecordsCollectionName)
}
