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

const MLModelStoreRecordsCollectionName = "mlmodel_store_records"

type MLModelInfoDoc struct {
	ModelUniqueID     int64                `bson:"modelUniqueId"`
	MlFileAddr        string               `bson:"mlFileAddr"`
	MlStorageSize     int64                `bson:"mlStorageSize"`
	AllowConsumerList []AllowedConsumerDoc `bson:"allowConsumerList,omitempty"`
}

type MLModelDoc struct {
	ModelUniqueID int64  `bson:"modelUniqueId"`
	MlModel       []byte `bson:"mlModel"`
}

type AllowedConsumerDoc struct {
	NfInstanceID string `bson:"nfInstanceId,omitempty"`
	NfSetID      string `bson:"nfSetId,omitempty"`
}

type ModelStoreResultDoc struct {
	ModelUniqueID int64  `bson:"modelUniqueId"`
	StoreResult   string `bson:"storeResult"`
}

type MLModelStoreRecordDocument struct {
	ID               string              `bson:"_id,omitempty"`
	StoreTransID     string              `bson:"storeTransId"`
	NfInstanceID     string              `bson:"nfInstanceId,omitempty"`
	NfSetID          string              `bson:"nfSetId,omitempty"`
	MlModelInfo      []MLModelInfoDoc    `bson:"mlModelInfo,omitempty"`
	MlModels         []MLModelDoc        `bson:"mlModels,omitempty"`
	ModelUniqueID    int64               `bson:"modelUniqueId"`
	MlFileAddr       string              `bson:"mlFileAddr"`
	SourceAddr       string              `bson:"sourceAddr"`
	StorageSize      int64               `bson:"storageSize"`
	ModelStoreResult ModelStoreResultDoc `bson:"modelStoreResult"`
	ExpiryTime       time.Time           `bson:"expiryTime,omitempty"`
	CreatedAt        time.Time           `bson:"createdAt"`
	UpdatedAt        time.Time           `bson:"updatedAt"`
}

type MLModelRepository struct {
	dbName string
}

func NewMLModelRepository(dbName string) *MLModelRepository {
	return &MLModelRepository{dbName: dbName}
}

func (r *MLModelRepository) EnsureIndexes(ctx context.Context) error {
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
				SetName("idx_mlmodel_store_trans_id").
				SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "modelUniqueId", Value: 1}},
			Options: options.Index().
				SetName("idx_mlmodel_unique_id").
				SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "nfInstanceId", Value: 1}},
			Options: options.Index().
				SetName("idx_mlmodel_nf_instance_id"),
		},
		{
			Keys: bson.D{{Key: "nfSetId", Value: 1}},
			Options: options.Index().
				SetName("idx_mlmodel_nf_set_id"),
		},
		{
			Keys: bson.D{{Key: "expiryTime", Value: 1}},
			Options: options.Index().
				SetName("idx_mlmodel_expiry_time"),
		},
	}

	logger.StoreLog.Infof("Ensuring Mongo indexes for collection: %s", MLModelStoreRecordsCollectionName)
	_, err := r.collection().Indexes().CreateMany(ctx, models)
	if err != nil {
		return fmt.Errorf("failed to create mlmodel store indexes: %w", err)
	}

	logger.StoreLog.Infof("ADRF Mongo indexes ensured for collection: %s", MLModelStoreRecordsCollectionName)
	return nil
}

func (r *MLModelRepository) InsertMLModelStoreRecord(ctx context.Context, doc *MLModelStoreRecordDocument) error {
	if doc == nil {
		return errors.New("mlmodel store document is nil")
	}
	if mongoapi.Client == nil {
		return errors.New("mongodb client is not initialized")
	}

	logger.StoreLog.Debugf("Inserting mlmodel record: storeTransId=%s modelUniqueId=%d", doc.StoreTransID, doc.ModelUniqueID)
	_, err := r.collection().InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to insert mlmodel store record: %w", err)
	}

	logger.StoreLog.Debugf("MLModel record inserted: storeTransId=%s", doc.StoreTransID)
	return nil
}

func (r *MLModelRepository) GetMLModelStoreRecord(ctx context.Context, storeTransId string) (*MLModelStoreRecordDocument, error) {
	if mongoapi.Client == nil {
		return nil, errors.New("mongodb client is not initialized")
	}

	var doc MLModelStoreRecordDocument
	err := r.collection().FindOne(ctx, bson.M{"storeTransId": storeTransId}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get mlmodel store record: %w", err)
	}
	return &doc, nil
}

func (r *MLModelRepository) GetMLModelStoreRecords(ctx context.Context, modelUniqueIds []int64) ([]*MLModelStoreRecordDocument, error) {
	if mongoapi.Client == nil {
		return nil, errors.New("mongodb client is not initialized")
	}

	filter := bson.M{}
	if len(modelUniqueIds) > 0 {
		filter["modelUniqueId"] = bson.M{"$in": modelUniqueIds}
	}

	cursor, err := r.collection().Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list mlmodel store records: %w", err)
	}
	defer cursor.Close(ctx)

	var results []*MLModelStoreRecordDocument
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode mlmodel store records: %w", err)
	}
	return results, nil
}

func (r *MLModelRepository) UpdateMLModelStoreRecord(ctx context.Context, storeTransId string, doc *MLModelStoreRecordDocument) error {
	if doc == nil {
		return errors.New("mlmodel store document is nil")
	}
	if mongoapi.Client == nil {
		return errors.New("mongodb client is not initialized")
	}

	filter := bson.M{"storeTransId": storeTransId}
	update := bson.M{
		"$set": bson.M{
			"nfInstanceId":     doc.NfInstanceID,
			"nfSetId":          doc.NfSetID,
			"mlModelInfo":      doc.MlModelInfo,
			"mlModels":         doc.MlModels,
			"modelUniqueId":    doc.ModelUniqueID,
			"mlFileAddr":       doc.MlFileAddr,
			"sourceAddr":       doc.SourceAddr,
			"storageSize":      doc.StorageSize,
			"modelStoreResult": doc.ModelStoreResult,
			"expiryTime":       doc.ExpiryTime,
			"updatedAt":        time.Now().UTC(),
		},
	}

	_, err := r.collection().UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update mlmodel store record: %w", err)
	}
	return nil
}

func (r *MLModelRepository) DeleteMLModelStoreRecord(ctx context.Context, storeTransId string) error {
	if mongoapi.Client == nil {
		return errors.New("mongodb client is not initialized")
	}

	result, err := r.collection().DeleteOne(ctx, bson.M{"storeTransId": storeTransId})
	if err != nil {
		return fmt.Errorf("failed to delete mlmodel store record: %w", err)
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// DeleteExpiredModels purges ML models whose expiryTime has passed.
func (r *MLModelRepository) DeleteExpiredModels(ctx context.Context, now time.Time) (int64, error) {
	if mongoapi.Client == nil {
		return 0, errors.New("mongodb client is not initialized")
	}

	filter := bson.M{
		"expiryTime": bson.M{"$gt": time.Time{}, "$lte": now},
	}
	res, err := r.collection().DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired mlmodel store records: %w", err)
	}
	return res.DeletedCount, nil
}

func (r *MLModelRepository) collection() *mongo.Collection {
	return mongoapi.Client.Database(r.dbName).Collection(MLModelStoreRecordsCollectionName)
}
