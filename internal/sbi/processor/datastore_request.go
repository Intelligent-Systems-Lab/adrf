package processor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/free5gc/adrf/internal/logger"
	"github.com/free5gc/adrf/internal/store"
	"github.com/free5gc/adrf/pkg/factory"
	"github.com/free5gc/openapi/models"
)

const (
	dataStoreInsertTimeout = 5 * time.Second
)

// dataStoreRecordPayload models the TS 29.575 data-path store request shape.
//
// This implementation supports the `dataSub + dataNotif` variant from the
// oneOf schema. The analytics variant (`anaSub + anaNotifications`) is not
// handled by this handler yet.
type dataStoreRecordPayload struct {
	DataSub   []map[string]any `json:"dataSub"`
	DataNotif map[string]any   `json:"dataNotif"`
}

func (p *Processor) handleCreateDataStoreRecord(c *gin.Context) {
	if p.dataStoreRepo == nil {
		logger.ProcLog.Error("DataStore repository is not initialized")
		p.writeProblem(c, http.StatusInternalServerError, "SYSTEM_FAILURE", "DataStore repository is unavailable")
		return
	}

	var req dataStoreRecordPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.ProcLog.Errorf("Failed to parse data-store request: %v", err)
		p.writeProblem(c, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}
	logger.ProcLog.Infof("StorageRequest accepted: dataSubCount=%d", len(req.DataSub))

	if problem := validateDataStoreRecordPayload(&req); problem != nil {
		logger.ProcLog.Warnf("Invalid data-store request: %+v", problem)
		c.JSON(int(problem.Status), problem)
		return
	}

	supi, err := extractSupiFromDataSub(req.DataSub)
	if err != nil {
		logger.ProcLog.Warnf("Failed to extract SUPI from dataSub: %v", err)
		p.writeProblem(c, http.StatusBadRequest, "MANDATORY_IE_MISSING", "dataSub.smfDataSub.supi is required")
		return
	}
	logger.ProcLog.Infof("StorageRequest SUPI resolved: %s", supi)

	// IngestedAt is assigned at ADRF accept time so retrieval queries can apply
	// deterministic snapshot boundaries independently from notification timestamps.
	doc := store.NewDataStoreRecordDocument(
		supi,
		toBSONMapList(req.DataSub),
		bson.M(req.DataNotif),
		time.Now().UTC(),
	)

	insertCtx, cancel := context.WithTimeout(c.Request.Context(), dataStoreInsertTimeout)
	defer cancel()

	insertErr := p.dataStoreRepo.InsertDataStoreRecord(insertCtx, doc)
	if insertErr != nil {
		logger.ProcLog.Errorf("Failed to persist data-store record: %v", insertErr)
		p.writeProblem(c, http.StatusInternalServerError, "SYSTEM_FAILURE", "Failed to persist data-store record")
		return
	}

	location := fmt.Sprintf(
		"%s%s/%s",
		factory.AdrfDataManagementResUriPrefix,
		factory.AdrfDataStoreRecordsPath,
		doc.StoreTransID,
	)
	c.Header("Location", location)

	logger.StoreLog.Infof(
		"StorageRequest completed: storeTransId=%s supi=%s location=%s",
		doc.StoreTransID,
		doc.Supi,
		location,
	)
	c.JSON(http.StatusCreated, req)
}

func validateDataStoreRecordPayload(req *dataStoreRecordPayload) *models.ProblemDetails {
	if req == nil {
		return &models.ProblemDetails{
			Status: http.StatusBadRequest,
			Cause:  "MANDATORY_IE_MISSING",
			Detail: "Request body is required",
		}
	}
	if len(req.DataSub) == 0 {
		return &models.ProblemDetails{
			Status: http.StatusBadRequest,
			Cause:  "MANDATORY_IE_MISSING",
			Detail: "dataSub is required",
		}
	}
	if len(req.DataNotif) == 0 {
		return &models.ProblemDetails{
			Status: http.StatusBadRequest,
			Cause:  "MANDATORY_IE_MISSING",
			Detail: "dataNotif is required",
		}
	}
	return nil
}

func extractSupiFromDataSub(dataSub []map[string]any) (string, error) {
	for _, sub := range dataSub {
		if sub == nil {
			continue
		}

		smfDataSubRaw, ok := sub["smfDataSub"]
		if !ok {
			continue
		}

		smfDataSub, ok := smfDataSubRaw.(map[string]any)
		if !ok {
			continue
		}

		supiRaw, ok := smfDataSub["supi"]
		if !ok {
			continue
		}

		supi, ok := supiRaw.(string)
		if !ok {
			continue
		}

		supi = strings.TrimSpace(supi)
		if supi == "" {
			continue
		}

		return supi, nil
	}

	return "", errors.New("supi not found in dataSub.smfDataSub")
}

func toBSONMapList(dataSub []map[string]any) []bson.M {
	result := make([]bson.M, 0, len(dataSub))
	for _, sub := range dataSub {
		if sub == nil {
			result = append(result, bson.M{})
			continue
		}
		result = append(result, bson.M(sub))
	}
	return result
}

func (p *Processor) writeProblem(c *gin.Context, status int, cause, detail string) {
	c.JSON(status, models.ProblemDetails{
		Status: int32(status),
		Cause:  cause,
		Detail: detail,
	})
}
