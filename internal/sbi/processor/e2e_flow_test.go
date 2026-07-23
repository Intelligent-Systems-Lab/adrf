package processor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/free5gc/adrf/internal/store"
	"github.com/free5gc/adrf/pkg/factory"
)

const (
	e2eNotifCorrID = "e2e-retrain-job-001"
	e2eSupi        = "imsi-001010000000001"
)

func TestE2EStoreSubscribeNotifyFetchUnsubscribe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalConfig := factory.AdrfConfig
	factory.AdrfConfig = &factory.Config{
		Configuration: &factory.Configuration{
			Retrieval: &factory.Retrieval{
				CorrIDBatchSize: 2,
			},
		},
	}
	t.Cleanup(func() {
		factory.AdrfConfig = originalConfig
	})

	repo := newInMemoryDataStoreRepo()
	processor := NewProcessor(repo)
	router := newE2ERouter(processor)

	fakeNWDAF := newFakeNWDAFCallback(router)
	processor.retrievalNotifier = &httpRetrievalNotificationSender{
		httpClient: &http.Client{Transport: fakeNWDAF},
	}

	startTimes := []string{
		"2026-03-26T09:05:00Z",
		"2026-03-26T09:10:00Z",
		"2026-03-26T09:15:00Z",
	}
	for _, startTime := range startTimes {
		resp := performE2ERequest(
			router,
			http.MethodPost,
			factory.AdrfDataManagementResUriPrefix+factory.AdrfDataStoreRecordsPath,
			buildE2EStorePayload(e2eSupi, startTime),
			func(req *http.Request) {
				req.Header.Set("Content-Type", "application/json")
				req.Host = "adrf.local"
			},
		)
		if resp.Code != http.StatusCreated {
			t.Fatalf("expected store status 201, got %d, body=%s", resp.Code, resp.Body.String())
		}
	}

	subscribeResp := performE2ERequest(
		router,
		http.MethodPost,
		factory.AdrfDataManagementResUriPrefix+factory.AdrfDataRetrievalSubscriptionsPath,
		buildE2ERetrievalSubscribePayload(e2eNotifCorrID, e2eSupi),
		func(req *http.Request) {
			req.Header.Set("Content-Type", "application/json")
			req.Host = "adrf.local"
		},
	)
	if subscribeResp.Code != http.StatusCreated {
		t.Fatalf(
			"expected retrieval-subscribe status 201, got %d, body=%s",
			subscribeResp.Code,
			subscribeResp.Body.String(),
		)
	}

	subscriptionID := extractResourceIDFromLocationHeader(
		subscribeResp.Header().Get("Location"),
		"/data-retrieval-subscriptions/",
	)
	if subscriptionID == "" {
		t.Fatal("expected non-empty subscription ID from Location header")
	}

	select {
	case <-fakeNWDAF.done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for retrieval notify completion")
	}

	notifies := fakeNWDAF.Notifications()
	if len(notifies) != 2 {
		t.Fatalf("expected 2 notify callbacks by corrIdBatchSize=2, got %d", len(notifies))
	}
	if notifies[0].TerminationReq {
		t.Fatal("expected first callback terminationReq=false")
	}
	if len(notifies[0].FetchInstruct.FetchCorrIDs) != 2 {
		t.Fatalf("expected first callback corr ids=2, got %d", len(notifies[0].FetchInstruct.FetchCorrIDs))
	}
	if !notifies[1].TerminationReq {
		t.Fatal("expected last callback terminationReq=true")
	}
	if len(notifies[1].FetchInstruct.FetchCorrIDs) != 1 {
		t.Fatalf("expected last callback corr ids=1, got %d", len(notifies[1].FetchInstruct.FetchCorrIDs))
	}

	fetchedRecords := fakeNWDAF.FetchedRecords()
	if len(fetchedRecords) != 3 {
		t.Fatalf("expected 3 fetched records, got %d", len(fetchedRecords))
	}
	for _, record := range fetchedRecords {
		supi, err := extractSupiFromDataSub(record.DataSub)
		if err != nil {
			t.Fatalf("failed to extract fetched SUPI: %v", err)
		}
		if supi != e2eSupi {
			t.Fatalf("unexpected fetched SUPI: %s", supi)
		}
	}

	deleteResp := performE2ERequest(
		router,
		http.MethodDelete,
		factory.AdrfDataManagementResUriPrefix+
			factory.AdrfDataRetrievalSubscriptionsPath+
			"/"+subscriptionID,
		"",
		nil,
	)
	if deleteResp.Code != http.StatusNoContent {
		t.Fatalf("expected unsubscribe status 204, got %d", deleteResp.Code)
	}
	if _, exists := processor.getRetrievalSubscriptionState(subscriptionID); exists {
		t.Fatal("expected retrieval subscription state removed after unsubscribe")
	}
}

type inMemoryDataStoreRepo struct {
	mu    sync.RWMutex
	byID  map[string]*store.NadrfDataStoreRecordDocument
	order []string
}

func newInMemoryDataStoreRepo() *inMemoryDataStoreRepo {
	return &inMemoryDataStoreRepo{
		byID: make(map[string]*store.NadrfDataStoreRecordDocument),
	}
}

func (r *inMemoryDataStoreRepo) InsertDataStoreRecord(
	_ context.Context,
	doc *store.NadrfDataStoreRecordDocument,
) error {
	if doc == nil {
		return fmt.Errorf("nil data store document")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *doc
	r.byID[copied.StoreTransID] = &copied
	r.order = append(r.order, copied.StoreTransID)
	return nil
}

func (r *inMemoryDataStoreRepo) GetDataStoreRecordByStoreTransID(
	_ context.Context,
	storeTransID string,
) (*store.NadrfDataStoreRecordDocument, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	doc, ok := r.byID[storeTransID]
	if !ok {
		return nil, mongo.ErrNoDocuments
	}
	copied := *doc
	return &copied, nil
}

func (r *inMemoryDataStoreRepo) ListStoreTransIDsBySnapshot(
	_ context.Context,
	supi string,
	snapshotCutoff time.Time,
	windowStart time.Time,
	windowStop time.Time,
) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]string, 0, len(r.order))
	for _, storeTransID := range r.order {
		doc := r.byID[storeTransID]
		if doc == nil {
			continue
		}
		if strings.TrimSpace(doc.Supi) != strings.TrimSpace(supi) {
			continue
		}
		if doc.IngestedAt.After(snapshotCutoff) {
			continue
		}
		if !dataNotifHasStartTimeMatch(doc.DataNotif, windowStart, windowStop) {
			continue
		}
		result = append(result, storeTransID)
	}
	return result, nil
}

func (r *inMemoryDataStoreRepo) SearchRecordsByFilter(
	_ context.Context,
	_ string,
	_, _ *time.Time,
	_, _ int64,
) ([]*store.NadrfDataStoreRecordDocument, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var docs []*store.NadrfDataStoreRecordDocument
	for _, id := range r.order {
		if doc, ok := r.byID[id]; ok {
			docs = append(docs, doc)
		}
	}
	return docs, int64(len(docs)), nil
}

func dataNotifHasStartTimeMatch(dataNotif bson.M, windowStart time.Time, windowStop time.Time) bool {
	rawNotifs, ok := mapValueAny(dataNotif, "upfEventNotifs")
	if !ok {
		return false
	}

	for _, notifAny := range toAnySliceE2E(rawNotifs) {
		notifMap, notifOK := toMapAny(notifAny)
		if !notifOK {
			continue
		}
		rawItems, itemsOK := mapValueAny(notifMap, "notificationItems")
		if !itemsOK {
			continue
		}
		for _, itemAny := range toAnySliceE2E(rawItems) {
			itemMap, itemOK := toMapAny(itemAny)
			if !itemOK {
				continue
			}
			rawStartTime, startTimeOK := mapValueAny(itemMap, "startTime")
			if !startTimeOK {
				continue
			}
			startTime, parsed := parseTimeAny(rawStartTime)
			if !parsed {
				continue
			}
			startTime = startTime.UTC()
			if startTime.Before(windowStart) || startTime.After(windowStop) {
				continue
			}
			return true
		}
	}

	return false
}

func mapValueAny(m map[string]any, key string) (any, bool) {
	value, ok := m[key]
	return value, ok
}

func toAnySliceE2E(raw any) []any {
	switch value := raw.(type) {
	case []any:
		return value
	default:
		return nil
	}
}

func toMapAny(raw any) (map[string]any, bool) {
	switch value := raw.(type) {
	case map[string]any:
		return value, true
	case bson.M:
		return map[string]any(value), true
	default:
		return nil, false
	}
}

func parseTimeAny(raw any) (time.Time, bool) {
	switch value := raw.(type) {
	case time.Time:
		return value, true
	case string:
		parsed, err := time.Parse(time.RFC3339Nano, value)
		if err != nil {
			parsed, err = time.Parse(time.RFC3339, value)
		}
		if err != nil {
			return time.Time{}, false
		}
		return parsed, true
	default:
		return time.Time{}, false
	}
}

type fakeNWDAFCallback struct {
	router *gin.Engine

	mu            sync.Mutex
	notifications []retrievalFetchNotificationPayload
	fetched       []dataStoreRecordPayload

	done     chan struct{}
	doneOnce sync.Once
}

func newFakeNWDAFCallback(router *gin.Engine) *fakeNWDAFCallback {
	return &fakeNWDAFCallback{
		router: router,
		done:   make(chan struct{}),
	}
}

func (f *fakeNWDAFCallback) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	if err = req.Body.Close(); err != nil {
		return nil, err
	}

	var notify retrievalFetchNotificationPayload
	if err = json.Unmarshal(body, &notify); err != nil {
		return nil, err
	}

	for _, corrID := range notify.FetchInstruct.FetchCorrIDs {
		fetchPath, pathErr := buildFetchPath(notify.FetchInstruct.FetchURI, corrID)
		if pathErr != nil {
			return nil, pathErr
		}
		fetchResp := performE2ERequest(f.router, http.MethodGet, fetchPath, "", nil)

		switch fetchResp.Code {
		case http.StatusOK:
			var record dataStoreRecordPayload
			if err = json.Unmarshal(fetchResp.Body.Bytes(), &record); err != nil {
				return nil, err
			}
			f.mu.Lock()
			f.fetched = append(f.fetched, record)
			f.mu.Unlock()
		case http.StatusNoContent:
			// Valid no-data response by spec; keep going.
		default:
			return nil, fmt.Errorf("unexpected fetch status from ADRF: %d", fetchResp.Code)
		}
	}

	f.mu.Lock()
	f.notifications = append(f.notifications, notify)
	f.mu.Unlock()

	if notify.TerminationReq {
		f.doneOnce.Do(func() {
			close(f.done)
		})
	}

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBuffer(nil)),
	}, nil
}

func (f *fakeNWDAFCallback) Notifications() []retrievalFetchNotificationPayload {
	f.mu.Lock()
	defer f.mu.Unlock()

	result := make([]retrievalFetchNotificationPayload, 0, len(f.notifications))
	result = append(result, f.notifications...)
	return result
}

func (f *fakeNWDAFCallback) FetchedRecords() []dataStoreRecordPayload {
	f.mu.Lock()
	defer f.mu.Unlock()

	result := make([]dataStoreRecordPayload, 0, len(f.fetched))
	result = append(result, f.fetched...)
	return result
}

func buildFetchPath(fetchURI string, corrID string) (string, error) {
	parsedURI, err := url.Parse(fetchURI)
	if err != nil {
		return "", err
	}

	path := parsedURI.Path
	if path == "" {
		path = fetchURI
	}

	return fmt.Sprintf("%s?fetch-correlation-ids=%s", path, url.QueryEscape(corrID)), nil
}

func newE2ERouter(p *Processor) *gin.Engine {
	router := gin.New()
	basePath := factory.AdrfDataManagementResUriPrefix

	router.POST(basePath+factory.AdrfDataStoreRecordsPath, p.HandleCreateDataStoreRecord)
	router.GET(basePath+factory.AdrfDataStoreRecordsPath, p.HandleGetDataStoreRecords)
	router.POST(basePath+factory.AdrfDataRetrievalSubscriptionsPath, p.HandleCreateDataRetrievalSubscription)
	router.DELETE(
		basePath+factory.AdrfDataRetrievalSubscriptionsPath+"/:subscriptionId",
		p.HandleDeleteDataRetrievalSubscription,
	)
	return router
}

func performE2ERequest(
	router *gin.Engine,
	method string,
	path string,
	body string,
	mutateReq func(*http.Request),
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if mutateReq != nil {
		mutateReq(req)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}

func buildE2EStorePayload(supi string, startTime string) string {
	return fmt.Sprintf(`{
  "dataSub": [
    {
      "smfDataSub": {
        "supi": "%s",
        "notifId": "notif-e2e-001"
      }
    }
  ],
  "dataNotif": {
    "upfEventNotifs": [
      {
        "notifId": "notif-e2e-001",
        "notificationItems": [
          {
            "startTime": "%s"
          }
        ]
      }
    ]
  }
}`, supi, startTime)
}

func buildE2ERetrievalSubscribePayload(notifCorrID string, supi string) string {
	return fmt.Sprintf(`{
  "notifCorrId": "%s",
  "notificationURI": "http://nwdaf.local/adrf/retrieval-notify",
  "timePeriod": {
    "startTime": "2026-03-26T09:00:00Z",
    "stopTime": "2026-03-26T09:30:00Z"
  },
  "dataSub": {
    "smfDataSub": {
      "supi": "%s"
    }
  },
  "consTrigNotif": true
}`, notifCorrID, supi)
}

func extractResourceIDFromLocationHeader(location string, resourcePath string) string {
	if location == "" {
		return ""
	}
	idx := strings.LastIndex(location, resourcePath)
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(location[idx+len(resourcePath):])
}
