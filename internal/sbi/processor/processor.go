package processor

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/adrf/internal/logger"
	"github.com/free5gc/adrf/internal/store"
	"github.com/free5gc/openapi/models"
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
	logger.ProcLog.Warn("GetDataStoreRecords is not implemented")
	p.writeNotImplemented(c, "GET_DATA_STORE_RECORDS")
}

func (p *Processor) HandleDeleteDataRetrievalSubscription(c *gin.Context) {
	logger.ProcLog.Warn("DeleteDataRetrievalSubscription is not implemented")
	p.writeNotImplemented(c, "DELETE_DATA_RETRIEVAL_SUBSCRIPTION")
}

func (p *Processor) writeNotImplemented(c *gin.Context, cause string) {
	c.JSON(http.StatusNotImplemented, models.ProblemDetails{
		Status: http.StatusNotImplemented,
		Cause:  cause,
		Detail: "The requested ADRF operation is not implemented in this build.",
	})
}
