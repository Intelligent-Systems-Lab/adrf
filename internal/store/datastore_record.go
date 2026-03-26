package store

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DataStoreRecordsCollectionName is the Mongo collection for ADRF storage records.
const DataStoreRecordsCollectionName = "data_store_records"

// NadrfDataStoreRecordDocument is the persisted representation of one
// NadrfDataStoreRecord payload in ADRF.
//
// Design notes:
//  1. storeTransId is stored explicitly and indexed uniquely because it is the
//     external fetch correlation key used by retrieval flows.
//  2. ingestedAt captures ADRF acceptance time (not UPF measurement time) so
//     snapshot cutoff logic can be implemented deterministically later.
//  3. supi is denormalized for fast retrieval filtering and index locality.
type NadrfDataStoreRecordDocument struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	StoreTransID string             `bson:"storeTransId"`
	Supi         string             `bson:"supi"`
	IngestedAt   time.Time          `bson:"ingestedAt"`
	DataSub      []bson.M           `bson:"dataSub"`
	DataNotif    bson.M             `bson:"dataNotif"`
}

// NewStoreTransID generates a new store transaction identifier.
//
// UUID is chosen to avoid coordination between ADRF instances and keeps ID
// generation lock-free under concurrent write load.
func NewStoreTransID() string {
	return uuid.NewString()
}

// NewDataStoreRecordDocument creates a document with internal metadata filled.
func NewDataStoreRecordDocument(
	supi string,
	dataSub []bson.M,
	dataNotif bson.M,
	ingestedAt time.Time,
) *NadrfDataStoreRecordDocument {
	if ingestedAt.IsZero() {
		ingestedAt = time.Now().UTC()
	}

	return &NadrfDataStoreRecordDocument{
		StoreTransID: NewStoreTransID(),
		Supi:         supi,
		IngestedAt:   ingestedAt,
		DataSub:      dataSub,
		DataNotif:    dataNotif,
	}
}
