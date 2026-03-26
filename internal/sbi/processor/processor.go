package processor

import (
	"context"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/adrf/internal/store"
)

// Processor owns ADRF business behavior behind SBI handlers.
//
// Implemented operations are routed to dedicated handler modules. Operations
// that are not supported yet return a structured "not implemented" response.
type Processor struct {
	dataStoreRepo dataStoreRecordWriter

	retrievalSubMu sync.RWMutex
	retrievalSubs  map[string]*retrievalSubscriptionState

	retrievalNotifier retrievalNotificationSender
}

// dataStoreRecordWriter is the persistence dependency needed by the store API.
//
// A narrow interface keeps the processor testable without a live Mongo instance,
// while still allowing the concrete repository to be injected in production.
type dataStoreRecordWriter interface {
	InsertDataStoreRecord(ctx context.Context, doc *store.NadrfDataStoreRecordDocument) error
	GetDataStoreRecordByStoreTransID(
		ctx context.Context,
		storeTransID string,
	) (*store.NadrfDataStoreRecordDocument, error)
	ListStoreTransIDsBySnapshot(
		ctx context.Context,
		supi string,
		snapshotCutoff time.Time,
		windowStart time.Time,
		windowStop time.Time,
	) ([]string, error)
}

func NewProcessor(dataStoreRepo dataStoreRecordWriter) *Processor {
	return &Processor{
		dataStoreRepo:     dataStoreRepo,
		retrievalSubs:     make(map[string]*retrievalSubscriptionState),
		retrievalNotifier: newHTTPRetrievalNotificationSender(),
	}
}

func (p *Processor) HandleCreateDataStoreRecord(c *gin.Context) {
	p.handleCreateDataStoreRecord(c)
}

func (p *Processor) HandleCreateDataRetrievalSubscription(c *gin.Context) {
	p.handleCreateDataRetrievalSubscription(c)
}

func (p *Processor) HandleGetDataStoreRecords(c *gin.Context) {
	p.handleGetDataStoreRecords(c)
}

func (p *Processor) HandleDeleteDataRetrievalSubscription(c *gin.Context) {
	p.handleDeleteDataRetrievalSubscription(c)
}
