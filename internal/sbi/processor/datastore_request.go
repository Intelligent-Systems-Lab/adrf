package processor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/free5gc/adrf/internal/logger"
	"github.com/free5gc/adrf/internal/store"
	"github.com/free5gc/adrf/pkg/factory"
	"github.com/free5gc/openapi/models"
)

const (
	dataStoreInsertTimeout = 5 * time.Second

	adrfMediaTypeJSON = "application/json"
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

type dataStoreSearchRequest struct {
	Supi      string `json:"supi,omitempty"`
	StartTime string `json:"startTime,omitempty"`
	EndTime   string `json:"endTime,omitempty"`
	Limit     int64  `json:"limit,omitempty"`
	Offset    int64  `json:"offset,omitempty"`
}

type dataStoreSearchResponse struct {
	Total   int64                           `json:"total"`
	Records []*store.NadrfDataStoreRecordDocument `json:"records"`
}

func (p *Processor) HandleSearchDataStoreRecords(c *gin.Context) {
	if p.dataStoreRepo == nil {
		p.writeProblem(c, newProblemDetails(http.StatusInternalServerError, "SYSTEM_FAILURE", "DataStore repository is unavailable", nil))
		return
	}

	var req dataStoreSearchRequest
	if c.Request.Method == http.MethodPost {
		_ = c.ShouldBindJSON(&req)
	}
	if req.Supi == "" {
		req.Supi = c.Query("supi")
	}

	var startTime, endTime *time.Time
	if req.StartTime == "" {
		req.StartTime = c.Query("startTime")
	}
	if req.EndTime == "" {
		req.EndTime = c.Query("endTime")
	}
	if req.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, req.StartTime); err == nil {
			startTime = &t
		}
	}
	if req.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
			endTime = &t
		}
	}

	if req.Limit <= 0 {
		req.Limit = 100
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	records, total, err := p.dataStoreRepo.SearchRecordsByFilter(ctx, req.Supi, startTime, endTime, req.Limit, req.Offset)
	if err != nil {
		logger.ProcLog.Errorf("SearchRecordsByFilter failed: %v", err)
		p.writeProblem(c, newProblemDetails(http.StatusInternalServerError, "SYSTEM_FAILURE", "Search data store records failed", nil))
		return
	}

	c.JSON(http.StatusOK, dataStoreSearchResponse{
		Total:   total,
		Records: records,
	})
}

func (p *Processor) handleCreateDataStoreRecord(c *gin.Context) {
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

	if problem := validateDataStoreRequestHeaders(c.Request); problem != nil {
		logger.ProcLog.Warnf(
			"Rejected data-store request by header validation: status=%d cause=%s detail=%s",
			problem.Status,
			problem.Cause,
			problem.Detail,
		)
		p.writeProblem(c, problem)
		return
	}

	var req dataStoreRecordPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		problem := mapJSONBindingErrorToProblemDetails(err)
		logger.ProcLog.Warnf(
			"Failed to parse data-store request: status=%d cause=%s detail=%s err=%v",
			problem.Status,
			problem.Cause,
			problem.Detail,
			err,
		)
		p.writeProblem(c, problem)
		return
	}
	logger.ProcLog.Debugf("StorageRequest accepted: dataSubCount=%d", len(req.DataSub))

	if problem := validateDataStoreRecordPayload(&req); problem != nil {
		logger.ProcLog.Warnf(
			"Invalid data-store request payload: status=%d cause=%s detail=%s",
			problem.Status,
			problem.Cause,
			problem.Detail,
		)
		p.writeProblem(c, problem)
		return
	}

	supi, err := extractSupiFromDataSub(req.DataSub)
	if err != nil {
		logger.ProcLog.Warnf("Failed to extract SUPI from dataSub: %v", err)
		p.writeProblem(c, newProblemDetails(
			http.StatusBadRequest,
			"MANDATORY_IE_MISSING",
			"dataSub.smfDataSub.supi is required for ADRF V0 data path",
			[]models.InvalidParam{
				invalidParam("/dataSub/*/smfDataSub/supi", "supi is required and shall be non-empty"),
			},
		))
		return
	}
	logger.ProcLog.Debugf("StorageRequest SUPI resolved: %s", summarizeIdentifier(supi))

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
		p.writeProblem(c, mapDataStoreInsertErrorToProblemDetails(insertErr))
		return
	}

	location := buildDataStoreRecordLocation(c, doc.StoreTransID)
	c.Header("Location", location)

	logger.StoreLog.Debugf(
		"StorageRequest completed: storeTransId=%s supi=%s location=%s",
		summarizeIdentifier(doc.StoreTransID),
		summarizeIdentifier(doc.Supi),
		location,
	)
	c.JSON(http.StatusCreated, req)
}

func validateDataStoreRecordPayload(req *dataStoreRecordPayload) *models.ProblemDetails {
	if req == nil {
		return newProblemDetails(
			http.StatusBadRequest,
			"MANDATORY_IE_MISSING",
			"Request body is required",
			[]models.InvalidParam{
				invalidParam("/", "request body is required"),
			},
		)
	}

	invalidParams := make([]models.InvalidParam, 0, 2)
	if len(req.DataSub) == 0 {
		invalidParams = append(
			invalidParams,
			invalidParam("/dataSub", "dataSub is required and shall contain at least one item"),
		)
	}
	if len(req.DataNotif) == 0 {
		invalidParams = append(
			invalidParams,
			invalidParam("/dataNotif", "dataNotif is required"),
		)
	}
	if len(invalidParams) > 0 {
		return newProblemDetails(
			http.StatusBadRequest,
			"MANDATORY_IE_MISSING",
			"Mandatory IE is missing in NadrfDataStoreRecord payload",
			invalidParams,
		)
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

func validateDataStoreRequestHeaders(req *http.Request) *models.ProblemDetails {
	if req == nil {
		return newProblemDetails(
			http.StatusInternalServerError,
			"SYSTEM_FAILURE",
			"Request context is unavailable",
			nil,
		)
	}

	// TS 29.571 common responses include 411. In ADRF V0 we enforce either a
	// Content-Length header or chunked transfer encoding so malformed clients
	// fail deterministically before body parsing.
	if req.ContentLength < 0 && !hasChunkedTransferEncoding(req) {
		return newProblemDetails(
			http.StatusLengthRequired,
			"LENGTH_REQUIRED",
			"Content-Length header is required when Transfer-Encoding is not chunked",
			[]models.InvalidParam{
				invalidParam("Content-Length", "header is mandatory unless chunked transfer encoding is used"),
			},
		)
	}

	contentType := strings.TrimSpace(req.Header.Get("Content-Type"))
	if contentType == "" {
		return newProblemDetails(
			http.StatusUnsupportedMediaType,
			"UNSUPPORTED_MEDIA_TYPE",
			"Content-Type must be application/json",
			[]models.InvalidParam{
				invalidParam("Content-Type", "application/json is required"),
			},
		)
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return newProblemDetails(
			http.StatusUnsupportedMediaType,
			"UNSUPPORTED_MEDIA_TYPE",
			"Content-Type header is malformed",
			[]models.InvalidParam{
				invalidParam("Content-Type", "unable to parse media type"),
			},
		)
	}
	if !strings.EqualFold(mediaType, adrfMediaTypeJSON) {
		return newProblemDetails(
			http.StatusUnsupportedMediaType,
			"UNSUPPORTED_MEDIA_TYPE",
			fmt.Sprintf("Unsupported media type %q; only application/json is allowed", mediaType),
			[]models.InvalidParam{
				invalidParam("Content-Type", "application/json is required"),
			},
		)
	}

	return nil
}

func mapJSONBindingErrorToProblemDetails(bindErr error) *models.ProblemDetails {
	if bindErr == nil {
		return newProblemDetails(http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload", nil)
	}
	if errors.Is(bindErr, io.EOF) {
		return newProblemDetails(
			http.StatusBadRequest,
			"MANDATORY_IE_MISSING",
			"Request body is required",
			[]models.InvalidParam{
				invalidParam("/", "request body is required"),
			},
		)
	}

	var syntaxErr *json.SyntaxError
	if errors.As(bindErr, &syntaxErr) {
		return newProblemDetails(
			http.StatusBadRequest,
			"INVALID_JSON",
			fmt.Sprintf("Malformed JSON syntax at byte offset %d", syntaxErr.Offset),
			nil,
		)
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(bindErr, &typeErr) {
		path := "/"
		if typeErr.Field != "" {
			path = "/" + strings.ReplaceAll(typeErr.Field, ".", "/")
		}
		return newProblemDetails(
			http.StatusBadRequest,
			"INVALID_JSON",
			"JSON field type mismatch",
			[]models.InvalidParam{
				invalidParam(path, fmt.Sprintf("expects %s", typeErr.Type.String())),
			},
		)
	}

	return newProblemDetails(http.StatusBadRequest, "INVALID_JSON", bindErr.Error(), nil)
}

func hasChunkedTransferEncoding(req *http.Request) bool {
	for _, encoding := range req.TransferEncoding {
		if strings.EqualFold(strings.TrimSpace(encoding), "chunked") {
			return true
		}
	}
	return false
}

func mapDataStoreInsertErrorToProblemDetails(insertErr error) *models.ProblemDetails {
	switch {
	case insertErr == nil:
		return nil
	case errors.Is(insertErr, context.DeadlineExceeded):
		return newProblemDetails(
			http.StatusServiceUnavailable,
			"SYSTEM_FAILURE",
			"Data store operation timed out",
			nil,
		)
	case errors.Is(insertErr, context.Canceled):
		return newProblemDetails(
			http.StatusInternalServerError,
			"SYSTEM_FAILURE",
			"Data store operation was canceled",
			nil,
		)
	case mongo.IsDuplicateKeyError(insertErr):
		return newProblemDetails(
			http.StatusInternalServerError,
			"SYSTEM_FAILURE",
			"Duplicate store transaction identifier generated",
			nil,
		)
	default:
		return newProblemDetails(
			http.StatusInternalServerError,
			"SYSTEM_FAILURE",
			"Failed to persist data-store record",
			nil,
		)
	}
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

func buildDataStoreRecordLocation(c *gin.Context, storeTransID string) string {
	path := fmt.Sprintf(
		"%s%s/%s",
		factory.AdrfDataManagementResUriPrefix,
		factory.AdrfDataStoreRecordsPath,
		storeTransID,
	)
	if c == nil || c.Request == nil || c.Request.Host == "" {
		return path
	}
	return fmt.Sprintf("%s://%s%s", requestScheme(c.Request), c.Request.Host, path)
}

func requestScheme(req *http.Request) string {
	if req == nil {
		return "http"
	}

	if forwardedProto := req.Header.Get("X-Forwarded-Proto"); forwardedProto != "" {
		parts := strings.Split(forwardedProto, ",")
		if len(parts) > 0 {
			proto := strings.TrimSpace(parts[0])
			if proto != "" {
				return proto
			}
		}
	}
	if req.TLS != nil {
		return "https"
	}
	return "http"
}

func newProblemDetails(
	status int,
	cause string,
	detail string,
	invalidParams []models.InvalidParam,
) *models.ProblemDetails {
	problem := &models.ProblemDetails{
		Title:  http.StatusText(status),
		Status: int32(status),
		Cause:  cause,
		Detail: detail,
	}
	if len(invalidParams) > 0 {
		problem.InvalidParams = invalidParams
	}
	return problem
}

func invalidParam(param string, reason string) models.InvalidParam {
	return models.InvalidParam{
		Param:  param,
		Reason: reason,
	}
}

func (p *Processor) writeProblem(c *gin.Context, problem *models.ProblemDetails) {
	if problem == nil {
		problem = newProblemDetails(
			http.StatusInternalServerError,
			"SYSTEM_FAILURE",
			"Unexpected error",
			nil,
		)
	}
	status := int(problem.Status)
	if status == 0 {
		status = http.StatusInternalServerError
		problem.Status = int32(status)
	}
	c.JSON(status, problem)
}
