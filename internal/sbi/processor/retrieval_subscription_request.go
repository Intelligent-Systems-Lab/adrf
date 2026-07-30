package processor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/free5gc/adrf/internal/logger"
	"github.com/free5gc/adrf/pkg/factory"
	"github.com/free5gc/openapi/models"
)

const (
	retrievalSnapshotQueryTimeout = 10 * time.Second
)

// retrievalSubscriptionPayload models the request shape of
// POST /data-retrieval-subscriptions for ADRF V0.
//
// ADRF V0 supports the fetch-based data path (`consTrigNotif=true`) with
// `dataSub.smfDataSub.supi` as the retrieval scope key.
type retrievalSubscriptionPayload struct {
	NotifCorrID     string             `json:"notifCorrId"`
	NotificationURI string             `json:"notificationURI"`
	TimePeriod      *timeWindowPayload `json:"timePeriod"`
	DataSub         map[string]any     `json:"dataSub,omitempty"`
	ConsTrigNotif   *bool              `json:"consTrigNotif,omitempty"`
	DataSetID       string             `json:"dataSetId,omitempty"`
}

// timeWindowPayload mirrors TS 29.122 TimeWindow fields used by retrieval
// subscription filtering.
type timeWindowPayload struct {
	StartTime *time.Time `json:"startTime"`
	StopTime  *time.Time `json:"stopTime"`
}

func (p *Processor) handleCreateDataRetrievalSubscription(c *gin.Context) {
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
			"Rejected retrieval-subscription request by header validation: status=%d cause=%s detail=%s",
			problem.Status,
			problem.Cause,
			problem.Detail,
		)
		p.writeProblem(c, problem)
		return
	}

	var req retrievalSubscriptionPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		problem := mapJSONBindingErrorToProblemDetails(err)
		logger.ProcLog.Warnf(
			"Failed to parse retrieval-subscription request: status=%d cause=%s detail=%s err=%v",
			problem.Status,
			problem.Cause,
			problem.Detail,
			err,
		)
		p.writeProblem(c, problem)
		return
	}

	if problem := validateDataRetrievalSubscriptionPayload(&req); problem != nil {
		logger.ProcLog.Warnf(
			"Invalid retrieval-subscription request payload: status=%d cause=%s detail=%s",
			problem.Status,
			problem.Cause,
			problem.Detail,
		)
		p.writeProblem(c, problem)
		return
	}

	supi, err := extractSupiFromDataSubObject(req.DataSub)
	if err != nil {
		logger.ProcLog.Warnf("Failed to extract SUPI from retrieval dataSub: %v", err)
		p.writeProblem(c, newProblemDetails(
			http.StatusBadRequest,
			"MANDATORY_IE_MISSING",
			"dataSub.smfDataSub.supi is required for retrieval snapshot filtering",
			[]models.InvalidParam{
				invalidParam("/dataSub/smfDataSub/supi", "supi is required and shall be non-empty"),
			},
		))
		return
	}

	snapshotAt := time.Now().UTC()
	windowStart := req.TimePeriod.StartTime.UTC()
	windowStop := req.TimePeriod.StopTime.UTC()

	snapshotCtx, snapshotCancel := context.WithTimeout(c.Request.Context(), retrievalSnapshotQueryTimeout)
	defer snapshotCancel()

	fetchCorrIDs, snapshotErr := p.dataStoreRepo.ListStoreTransIDsBySnapshot(
		snapshotCtx,
		supi,
		snapshotAt,
		windowStart,
		windowStop,
	)
	if snapshotErr != nil {
		logger.ProcLog.Errorf("Failed to build retrieval snapshot: %v", snapshotErr)
		p.writeProblem(c, mapSnapshotBuildErrorToProblemDetails(snapshotErr))
		return
	}

	subscriptionID := uuid.NewString()

	var snapshotRecords []any
	for _, id := range fetchCorrIDs {
		rec, err := p.dataStoreRepo.GetDataStoreRecordByStoreTransID(snapshotCtx, id)
		if err == nil && rec != nil {
			snapshotRecords = append(snapshotRecords, rec)
		}
	}
	snapshotDir := "./storage/snapshots"
	_ = os.MkdirAll(snapshotDir, 0755)
	snapshotFilePath := fmt.Sprintf("%s/%s.json", snapshotDir, subscriptionID)
	snapshotData, _ := json.Marshal(snapshotRecords)
	_ = os.WriteFile(snapshotFilePath, snapshotData, 0644)

	scheme := "http"
	if c != nil && c.Request != nil && c.Request.TLS != nil {
		scheme = "https"
	}
	host := "192.168.107.5:9888"
	if c != nil && c.Request != nil && c.Request.Host != "" {
		host = c.Request.Host
	}
	datasetURL := fmt.Sprintf("%s://%s/nadrf-datamanagement/v1/data-snapshots/%s/download", scheme, host, subscriptionID)

	dispatchCtx, dispatchCancel := context.WithCancel(context.Background())
	state := &retrievalSubscriptionState{
		SubscriptionID:  subscriptionID,
		NotifCorrID:     req.NotifCorrID,
		NotificationURI: req.NotificationURI,
		Supi:            supi,
		DatasetURL:      datasetURL,
		TimePeriodStart: windowStart,
		TimePeriodStop:  windowStop,
		SnapshotAt:      snapshotAt,
		CreatedAt:       snapshotAt,
		ConsTrigNotif:   req.ConsTrigNotif != nil && *req.ConsTrigNotif,
		FetchCorrIDs:    append([]string(nil), fetchCorrIDs...),
		DispatchCtx:     dispatchCtx,
		DispatchCancel:  dispatchCancel,
	}
	p.storeRetrievalSubscriptionState(state)

	location := buildDataRetrievalSubscriptionLocation(c, subscriptionID)
	c.Header("Location", location)

	logger.ProcLog.Infof(
		"RetrievalSubscribe created: subscriptionId=%s notifCorrId=%s supi=%s snapshotAt=%s fetchCorrIdCount=%d",
		summarizeIdentifier(subscriptionID),
		summarizeIdentifier(req.NotifCorrID),
		summarizeIdentifier(supi),
		snapshotAt.Format(time.RFC3339Nano),
		len(fetchCorrIDs),
	)
	c.JSON(http.StatusCreated, req)

	// Callback delivery is decoupled from API response latency. The subscription
	// is already created and stored, and callback retries/compensation can be
	// handled independently from this synchronous request lifecycle.
	fetchURI := buildDataStoreRecordsFetchURI(c)
	go p.dispatchRetrievalFetchNotifications(dispatchCtx, state, fetchURI)
}

func validateDataRetrievalSubscriptionPayload(req *retrievalSubscriptionPayload) *models.ProblemDetails {
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

	invalidParams := make([]models.InvalidParam, 0, 6)
	if strings.TrimSpace(req.NotifCorrID) == "" {
		invalidParams = append(
			invalidParams,
			invalidParam("/notifCorrId", "notifCorrId is required"),
		)
	}

	if strings.TrimSpace(req.NotificationURI) == "" {
		invalidParams = append(
			invalidParams,
			invalidParam("/notificationURI", "notificationURI is required"),
		)
	} else if _, err := url.ParseRequestURI(req.NotificationURI); err != nil {
		invalidParams = append(
			invalidParams,
			invalidParam("/notificationURI", "notificationURI shall be a valid absolute URI"),
		)
	}

	if req.TimePeriod == nil {
		invalidParams = append(
			invalidParams,
			invalidParam("/timePeriod", "timePeriod is required"),
		)
	} else {
		if req.TimePeriod.StartTime == nil {
			invalidParams = append(
				invalidParams,
				invalidParam("/timePeriod/startTime", "startTime is required"),
			)
		}
		if req.TimePeriod.StopTime == nil {
			invalidParams = append(
				invalidParams,
				invalidParam("/timePeriod/stopTime", "stopTime is required"),
			)
		}
		if req.TimePeriod.StartTime != nil &&
			req.TimePeriod.StopTime != nil &&
			req.TimePeriod.StartTime.After(*req.TimePeriod.StopTime) {
			invalidParams = append(
				invalidParams,
				invalidParam(
					"/timePeriod",
					"startTime shall be earlier than or equal to stopTime",
				),
			)
		}
	}

	if len(req.DataSub) == 0 {
		invalidParams = append(
			invalidParams,
			invalidParam("/dataSub", "dataSub is required in ADRF V0"),
		)
	}

	if req.ConsTrigNotif == nil || !*req.ConsTrigNotif {
		invalidParams = append(
			invalidParams,
			invalidParam(
				"/consTrigNotif",
				"consTrigNotif shall be true in ADRF V0 fetch mode",
			),
		)
	}

	if len(invalidParams) > 0 {
		return newProblemDetails(
			http.StatusBadRequest,
			"MANDATORY_IE_MISSING",
			"Mandatory IE is missing or invalid in NadrfDataRetrievalSubscription payload",
			invalidParams,
		)
	}

	return nil
}

func extractSupiFromDataSubObject(dataSub map[string]any) (string, error) {
	if len(dataSub) == 0 {
		return "", errors.New("dataSub is empty")
	}

	smfDataSubRaw, ok := dataSub["smfDataSub"]
	if !ok {
		return "", errors.New("smfDataSub is missing")
	}

	smfDataSub, ok := smfDataSubRaw.(map[string]any)
	if !ok {
		return "", errors.New("smfDataSub shall be an object")
	}

	supiRaw, ok := smfDataSub["supi"]
	if !ok {
		return "", errors.New("supi is missing")
	}

	supi, ok := supiRaw.(string)
	if !ok {
		return "", errors.New("supi shall be a string")
	}

	supi = strings.TrimSpace(supi)
	if supi == "" {
		return "", errors.New("supi is empty")
	}

	return supi, nil
}

func mapSnapshotBuildErrorToProblemDetails(snapshotErr error) *models.ProblemDetails {
	switch {
	case snapshotErr == nil:
		return nil
	case errors.Is(snapshotErr, context.DeadlineExceeded):
		return newProblemDetails(
			http.StatusServiceUnavailable,
			"SYSTEM_FAILURE",
			"Snapshot query timed out",
			nil,
		)
	case errors.Is(snapshotErr, context.Canceled):
		return newProblemDetails(
			http.StatusInternalServerError,
			"SYSTEM_FAILURE",
			"Snapshot query was canceled",
			nil,
		)
	default:
		return newProblemDetails(
			http.StatusInternalServerError,
			"SYSTEM_FAILURE",
			"Failed to build retrieval snapshot",
			nil,
		)
	}
}

func buildDataRetrievalSubscriptionLocation(c *gin.Context, subscriptionID string) string {
	path := fmt.Sprintf(
		"%s%s/%s",
		factory.AdrfDataManagementResUriPrefix,
		factory.AdrfDataRetrievalSubscriptionsPath,
		subscriptionID,
	)
	if c == nil || c.Request == nil || c.Request.Host == "" {
		return path
	}
	return fmt.Sprintf("%s://%s%s", requestScheme(c.Request), c.Request.Host, path)
}
