package processor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/adrf/pkg/factory"
)

func TestCreateDataRetrievalSubscriptionSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubDataStoreRepo{
		snapshotIDs: []string{"st-001", "st-002", "st-003"},
	}
	processor := NewProcessor(repo)
	router := newRetrievalSubscriptionRouter(processor)

	response := performRetrievalSubscriptionRequest(router, validRetrievalSubscriptionPayload(), func(req *http.Request) {
		req.Host = "adrf.local"
	})

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d, body=%s", response.Code, response.Body.String())
	}

	location := response.Header().Get("Location")
	prefix := "http://adrf.local/nadrf-datamanagement/v1/data-retrieval-subscriptions/"
	if !strings.HasPrefix(location, prefix) {
		t.Fatalf("unexpected Location header: %s", location)
	}

	subscriptionID := strings.TrimPrefix(location, prefix)
	if subscriptionID == "" {
		t.Fatal("expected non-empty subscription ID in Location header")
	}

	state, ok := processor.getRetrievalSubscriptionState(subscriptionID)
	if !ok {
		t.Fatalf("subscription state not found for id=%s", subscriptionID)
	}
	if state.Supi != "imsi-001010000000001" {
		t.Fatalf("unexpected SUPI in state: %s", state.Supi)
	}
	if len(state.FetchCorrIDs) != 3 {
		t.Fatalf("expected 3 fetch corr ids, got %d", len(state.FetchCorrIDs))
	}

	var body retrievalSubscriptionPayload
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body.NotifCorrID != "retrain-job-001" {
		t.Fatalf("unexpected notifCorrId: %s", body.NotifCorrID)
	}
}

func TestCreateDataRetrievalSubscriptionRejectWhenConsTrigNotifFalse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubDataStoreRepo{}
	router := newRetrievalSubscriptionRouter(NewProcessor(repo))

	response := performRetrievalSubscriptionRequest(router, retrievalPayloadWithConsTrigFalse(), nil)

	problem := assertProblemDetails(t, response, http.StatusBadRequest, "MANDATORY_IE_MISSING")
	if len(problem.InvalidParams) == 0 {
		t.Fatal("expected invalidParams in bad request response")
	}
}

func TestCreateDataRetrievalSubscriptionRejectInvalidWindow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubDataStoreRepo{}
	router := newRetrievalSubscriptionRouter(NewProcessor(repo))

	response := performRetrievalSubscriptionRequest(router, retrievalPayloadWithInvalidWindow(), nil)

	assertProblemDetails(t, response, http.StatusBadRequest, "MANDATORY_IE_MISSING")
}

func TestCreateDataRetrievalSubscriptionMapSnapshotTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubDataStoreRepo{snapshotErr: context.DeadlineExceeded}
	router := newRetrievalSubscriptionRouter(NewProcessor(repo))

	response := performRetrievalSubscriptionRequest(router, validRetrievalSubscriptionPayload(), nil)

	assertProblemDetails(t, response, http.StatusServiceUnavailable, "SYSTEM_FAILURE")
}

func TestCreateDataRetrievalSubscriptionRejectInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubDataStoreRepo{}
	router := newRetrievalSubscriptionRouter(NewProcessor(repo))

	response := performRetrievalSubscriptionRequest(router, "{", nil)

	assertProblemDetails(t, response, http.StatusBadRequest, "INVALID_JSON")
}

func newRetrievalSubscriptionRouter(p *Processor) *gin.Engine {
	router := gin.New()
	router.POST(
		factory.AdrfDataManagementResUriPrefix+factory.AdrfDataRetrievalSubscriptionsPath,
		p.HandleCreateDataRetrievalSubscription,
	)
	return router
}

func performRetrievalSubscriptionRequest(
	router *gin.Engine,
	body string,
	mutateReq func(*http.Request),
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(
		http.MethodPost,
		factory.AdrfDataManagementResUriPrefix+factory.AdrfDataRetrievalSubscriptionsPath,
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	if mutateReq != nil {
		mutateReq(req)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}

func validRetrievalSubscriptionPayload() string {
	return `{
  "notifCorrId": "retrain-job-001",
  "notificationURI": "http://nwdaf.local/adrf/retrieval-notify",
  "timePeriod": {
    "startTime": "2026-03-26T09:00:00Z",
    "stopTime": "2026-03-26T09:30:00Z"
  },
  "dataSub": {
    "smfDataSub": {
      "supi": "imsi-001010000000001"
    }
  },
  "consTrigNotif": true
}`
}

func retrievalPayloadWithConsTrigFalse() string {
	return `{
  "notifCorrId": "retrain-job-002",
  "notificationURI": "http://nwdaf.local/adrf/retrieval-notify",
  "timePeriod": {
    "startTime": "2026-03-26T09:00:00Z",
    "stopTime": "2026-03-26T09:30:00Z"
  },
  "dataSub": {
    "smfDataSub": {
      "supi": "imsi-001010000000001"
    }
  },
  "consTrigNotif": false
}`
}

func retrievalPayloadWithInvalidWindow() string {
	return `{
  "notifCorrId": "retrain-job-003",
  "notificationURI": "http://nwdaf.local/adrf/retrieval-notify",
  "timePeriod": {
    "startTime": "2026-03-26T10:00:00Z",
    "stopTime": "2026-03-26T09:30:00Z"
  },
  "dataSub": {
    "smfDataSub": {
      "supi": "imsi-001010000000001"
    }
  },
  "consTrigNotif": true
}`
}
