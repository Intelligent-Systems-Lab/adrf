package processor

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/adrf/internal/logger"
	"github.com/free5gc/openapi/models"
)

// Processor owns ADRF business behavior behind SBI handlers.
//
// In R01 this component intentionally returns NotImplemented for all data
// operations. The purpose is to lock down package boundaries, call flow, and
// error contract shape before implementing persistent logic in later rounds.
type Processor struct{}

func NewProcessor() *Processor {
	return &Processor{}
}

func (p *Processor) HandleCreateDataStoreRecord(c *gin.Context) {
	logger.ProcLog.Warn("CreateDataStoreRecord is not implemented in R01")
	p.writeNotImplemented(c, "CREATE_DATA_STORE_RECORD")
}

func (p *Processor) HandleCreateDataRetrievalSubscription(c *gin.Context) {
	logger.ProcLog.Warn("CreateDataRetrievalSubscription is not implemented in R01")
	p.writeNotImplemented(c, "CREATE_DATA_RETRIEVAL_SUBSCRIPTION")
}

func (p *Processor) HandleGetDataStoreRecords(c *gin.Context) {
	logger.ProcLog.Warn("GetDataStoreRecords is not implemented in R01")
	p.writeNotImplemented(c, "GET_DATA_STORE_RECORDS")
}

func (p *Processor) HandleDeleteDataRetrievalSubscription(c *gin.Context) {
	logger.ProcLog.Warn("DeleteDataRetrievalSubscription is not implemented in R01")
	p.writeNotImplemented(c, "DELETE_DATA_RETRIEVAL_SUBSCRIPTION")
}

func (p *Processor) writeNotImplemented(c *gin.Context, cause string) {
	c.JSON(http.StatusNotImplemented, models.ProblemDetails{
		Status: http.StatusNotImplemented,
		Cause:  cause,
		Detail: "ADRF R01 skeleton only; business logic will be added in next rounds.",
	})
}
