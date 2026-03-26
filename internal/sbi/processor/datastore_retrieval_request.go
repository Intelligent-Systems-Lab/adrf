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
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/free5gc/adrf/internal/logger"
	"github.com/free5gc/openapi/models"
)

const (
	dataStoreFetchTimeout = 5 * time.Second

	storeTransIDQueryKey      = "store-trans-id"
	fetchCorrelationIDsKey    = "fetch-correlation-ids"
	dataSetIDQueryKey         = "data-set-id"
	maxFetchCorrelationIDPerQ = 1
)

func (p *Processor) handleGetDataStoreRecords(c *gin.Context) {
	if p.dataStoreRepo == nil {
		logger.ProcLog.Error("DataStore repository is not initialized")
		p.writeProblem(c, newProblemDetails(
			http.StatusInternalServerError,
			"SYSTEM_FAILURE",
			"DataStore repository is unavailable",
			nil,
		))
		return
	}

	fetchCorrID, problem := resolveFetchCorrelationIDFromQuery(c.Request)
	if problem != nil {
		logger.ProcLog.Warnf(
			"Rejected retrieval request by query validation: status=%d cause=%s detail=%s",
			problem.Status,
			problem.Cause,
			problem.Detail,
		)
		p.writeProblem(c, problem)
		return
	}

	fetchCtx, cancel := context.WithTimeout(c.Request.Context(), dataStoreFetchTimeout)
	defer cancel()

	doc, fetchErr := p.dataStoreRepo.GetDataStoreRecordByStoreTransID(fetchCtx, fetchCorrID)
	if fetchErr != nil {
		if errors.Is(fetchErr, mongo.ErrNoDocuments) {
			logger.ProcLog.Debugf(
				"RetrievalRequest no data: fetchCorrId=%s (mapped to 204)",
				summarizeIdentifier(fetchCorrID),
			)
			c.Status(http.StatusNoContent)
			return
		}

		logger.ProcLog.Errorf(
			"Failed to fetch data-store record: fetchCorrId=%s err=%v",
			summarizeIdentifier(fetchCorrID),
			fetchErr,
		)
		p.writeProblem(c, mapDataStoreFetchErrorToProblemDetails(fetchErr))
		return
	}
	if doc == nil {
		logger.ProcLog.Debugf(
			"RetrievalRequest no data: fetchCorrId=%s (nil document mapped to 204)",
			summarizeIdentifier(fetchCorrID),
		)
		c.Status(http.StatusNoContent)
		return
	}

	response := dataStoreRecordPayload{
		DataSub:   toMapList(doc.DataSub),
		DataNotif: map[string]any(doc.DataNotif),
	}
	logger.ProcLog.Debugf(
		"RetrievalRequest completed: fetchCorrId=%s status=200",
		summarizeIdentifier(fetchCorrID),
	)
	c.JSON(http.StatusOK, response)
}

func resolveFetchCorrelationIDFromQuery(req *http.Request) (string, *models.ProblemDetails) {
	if req == nil || req.URL == nil {
		return "", newProblemDetails(
			http.StatusInternalServerError,
			"SYSTEM_FAILURE",
			"Request context is unavailable",
			nil,
		)
	}

	queryValues := req.URL.Query()
	hasStoreTransID := hasQueryParam(queryValues, storeTransIDQueryKey)
	hasFetchCorrelationIDs := hasQueryParam(queryValues, fetchCorrelationIDsKey)
	hasDataSetID := hasQueryParam(queryValues, dataSetIDQueryKey)

	providedCount := 0
	if hasStoreTransID {
		providedCount++
	}
	if hasFetchCorrelationIDs {
		providedCount++
	}
	if hasDataSetID {
		providedCount++
	}

	if providedCount == 0 {
		return "", newProblemDetails(
			http.StatusBadRequest,
			"MANDATORY_IE_MISSING",
			"Exactly one retrieval query parameter is required",
			[]models.InvalidParam{
				invalidParam(fetchCorrelationIDsKey, "query parameter is required in ADRF V0"),
			},
		)
	}
	if providedCount > 1 {
		return "", newProblemDetails(
			http.StatusBadRequest,
			"MANDATORY_IE_MISSING",
			"Exactly one of store-trans-id, fetch-correlation-ids, and data-set-id shall be provided",
			[]models.InvalidParam{
				invalidParam(fetchCorrelationIDsKey, "only one retrieval query key is allowed"),
			},
		)
	}
	if hasStoreTransID || hasDataSetID {
		return "", newProblemDetails(
			http.StatusBadRequest,
			"MANDATORY_IE_MISSING",
			"ADRF V0 retrieval supports fetch-correlation-ids only",
			[]models.InvalidParam{
				invalidParam(fetchCorrelationIDsKey, "store-trans-id and data-set-id are not supported in ADRF V0"),
			},
		)
	}

	fetchCorrIDs := parseFetchCorrelationIDs(queryValues[fetchCorrelationIDsKey])
	if len(fetchCorrIDs) == 0 {
		return "", newProblemDetails(
			http.StatusBadRequest,
			"MANDATORY_IE_MISSING",
			"fetch-correlation-ids shall contain at least one identifier",
			[]models.InvalidParam{
				invalidParam(fetchCorrelationIDsKey, "at least one non-empty identifier is required"),
			},
		)
	}
	if len(fetchCorrIDs) > maxFetchCorrelationIDPerQ {
		return "", newProblemDetails(
			http.StatusBadRequest,
			"MANDATORY_IE_MISSING",
			"ADRF V0 supports one fetch-correlation-id per request",
			[]models.InvalidParam{
				invalidParam(fetchCorrelationIDsKey, "multiple identifiers are not supported in ADRF V0"),
			},
		)
	}

	return fetchCorrIDs[0], nil
}

func hasQueryParam(queryValues map[string][]string, key string) bool {
	if queryValues == nil {
		return false
	}
	_, exists := queryValues[key]
	return exists
}

// parseFetchCorrelationIDs supports both repeated query style and comma-
// separated form style (`explode=false`) from TS 29.575 OpenAPI.
func parseFetchCorrelationIDs(rawValues []string) []string {
	result := make([]string, 0, len(rawValues))
	for _, rawValue := range rawValues {
		for _, token := range strings.Split(rawValue, ",") {
			token = strings.TrimSpace(token)
			if token == "" {
				continue
			}
			result = append(result, token)
		}
	}
	return result
}

func toMapList(dataSub []bson.M) []map[string]any {
	result := make([]map[string]any, 0, len(dataSub))
	for _, dataSubEntry := range dataSub {
		if dataSubEntry == nil {
			result = append(result, map[string]any{})
			continue
		}
		result = append(result, map[string]any(dataSubEntry))
	}
	return result
}

func mapDataStoreFetchErrorToProblemDetails(fetchErr error) *models.ProblemDetails {
	switch {
	case fetchErr == nil:
		return nil
	case errors.Is(fetchErr, context.DeadlineExceeded):
		return newProblemDetails(
			http.StatusServiceUnavailable,
			"SYSTEM_FAILURE",
			"Data retrieval operation timed out",
			nil,
		)
	case errors.Is(fetchErr, context.Canceled):
		return newProblemDetails(
			http.StatusInternalServerError,
			"SYSTEM_FAILURE",
			"Data retrieval operation was canceled",
			nil,
		)
	default:
		return newProblemDetails(
			http.StatusInternalServerError,
			"SYSTEM_FAILURE",
			fmt.Sprintf("Failed to fetch data-store record: %v", fetchErr),
			nil,
		)
	}
}
