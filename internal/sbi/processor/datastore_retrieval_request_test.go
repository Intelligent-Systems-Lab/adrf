package processor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/free5gc/adrf/internal/store"
	"github.com/free5gc/adrf/pkg/factory"
)

func TestGetDataStoreRecordsByFetchCorrelationIDSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubDataStoreRepo{
		fetchDoc: &store.NadrfDataStoreRecordDocument{
			StoreTransID: "st-001",
			Supi:         "imsi-001010000000001",
			IngestedAt:   time.Date(2026, 3, 26, 9, 5, 0, 0, time.UTC),
			DataSub: []bson.M{
				{
					"smfDataSub": bson.M{
						"supi": "imsi-001010000000001",
					},
				},
			},
			DataNotif: bson.M{
				"upfEventNotifs": []any{
					bson.M{
						"notifId": "notif-0001",
					},
				},
			},
		},
	}
	router := newDataStoreFetchRouter(NewProcessor(repo))

	response := performDataStoreFetchRequest(router, "st-001")
	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", response.Code, response.Body.String())
	}

	var body dataStoreRecordPayload
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if len(body.DataSub) != 1 {
		t.Fatalf("expected one dataSub item, got %d", len(body.DataSub))
	}
	if len(body.DataNotif) == 0 {
		t.Fatal("expected non-empty dataNotif")
	}
}

func TestGetDataStoreRecordsByFetchCorrelationIDNoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubDataStoreRepo{
		fetchErr: mongo.ErrNoDocuments,
	}
	router := newDataStoreFetchRouter(NewProcessor(repo))

	response := performDataStoreFetchRequest(router, "st-not-found")
	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", response.Code)
	}
	if strings.TrimSpace(response.Body.String()) != "" {
		t.Fatalf("expected empty body for 204, got %q", response.Body.String())
	}
}

func TestGetDataStoreRecordsRejectMissingFetchCorrelationIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubDataStoreRepo{}
	router := newDataStoreFetchRouter(NewProcessor(repo))

	req := httptest.NewRequest(
		http.MethodGet,
		factory.AdrfDataManagementResUriPrefix+factory.AdrfDataStoreRecordsPath,
		nil,
	)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	assertProblemDetails(t, response, http.StatusBadRequest, "MANDATORY_IE_MISSING")
}

func TestGetDataStoreRecordsRejectMultipleFetchCorrelationIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubDataStoreRepo{}
	router := newDataStoreFetchRouter(NewProcessor(repo))

	req := httptest.NewRequest(
		http.MethodGet,
		factory.AdrfDataManagementResUriPrefix+
			factory.AdrfDataStoreRecordsPath+
			"?fetch-correlation-ids=st-001,st-002",
		nil,
	)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	assertProblemDetails(t, response, http.StatusBadRequest, "MANDATORY_IE_MISSING")
}

func TestGetDataStoreRecordsMapFetchDeadlineExceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubDataStoreRepo{
		fetchErr: context.DeadlineExceeded,
	}
	router := newDataStoreFetchRouter(NewProcessor(repo))

	response := performDataStoreFetchRequest(router, "st-timeout")
	assertProblemDetails(t, response, http.StatusServiceUnavailable, "SYSTEM_FAILURE")
}

func TestGetDataStoreRecordsRejectWhenRepositoryUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := newDataStoreFetchRouter(NewProcessor(nil))
	response := performDataStoreFetchRequest(router, "st-001")

	assertProblemDetails(t, response, http.StatusInternalServerError, "SYSTEM_FAILURE")
}

func newDataStoreFetchRouter(p *Processor) *gin.Engine {
	router := gin.New()
	router.GET(
		factory.AdrfDataManagementResUriPrefix+factory.AdrfDataStoreRecordsPath,
		p.HandleGetDataStoreRecords,
	)
	return router
}

func performDataStoreFetchRequest(router *gin.Engine, fetchCorrID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(
		http.MethodGet,
		factory.AdrfDataManagementResUriPrefix+
			factory.AdrfDataStoreRecordsPath+
			"?fetch-correlation-ids="+fetchCorrID,
		nil,
	)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}
