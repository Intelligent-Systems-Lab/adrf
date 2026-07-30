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
	mlModelRepo   mlModelRecordWriter

	retrievalSubMu sync.RWMutex
	retrievalSubs  map[string]*retrievalSubscriptionState

	retrievalNotifier retrievalNotificationSender
}

// dataStoreRecordWriter is the persistence dependency needed by the store API.
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
	SearchRecordsByFilter(
		ctx context.Context,
		supi string,
		startTime, endTime *time.Time,
		limit, offset int64,
	) ([]*store.NadrfDataStoreRecordDocument, int64, error)
}

type mlModelRecordWriter interface {
	InsertMLModelStoreRecord(ctx context.Context, doc *store.MLModelStoreRecordDocument) error
	GetMLModelStoreRecord(ctx context.Context, storeTransId string) (*store.MLModelStoreRecordDocument, error)
	GetMLModelStoreRecords(ctx context.Context, modelUniqueIds []string) ([]*store.MLModelStoreRecordDocument, error)
	UpdateMLModelStoreRecord(ctx context.Context, storeTransId string, doc *store.MLModelStoreRecordDocument) error
	DeleteMLModelStoreRecord(ctx context.Context, storeTransId string) error
}

func NewProcessor(dataStoreRepo dataStoreRecordWriter) *Processor {
	return &Processor{
		dataStoreRepo:     dataStoreRepo,
		retrievalSubs:     make(map[string]*retrievalSubscriptionState),
		retrievalNotifier: newHTTPRetrievalNotificationSender(),
	}
}

func (p *Processor) SetMLModelRepo(mlModelRepo mlModelRecordWriter) {
	p.mlModelRepo = mlModelRepo
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

func (p *Processor) HandleDownloadDataSnapshot(c *gin.Context) {
	p.handleDownloadDataSnapshot(c)
}
