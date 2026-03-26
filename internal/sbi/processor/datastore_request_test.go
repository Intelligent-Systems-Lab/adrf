package processor

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/adrf/internal/store"
	"github.com/free5gc/adrf/pkg/factory"
	"github.com/free5gc/openapi/models"
)

type stubDataStoreRepo struct {
	insertErr   error
	inserted    []*store.NadrfDataStoreRecordDocument
	snapshotIDs []string
	snapshotErr error
	fetchDoc    *store.NadrfDataStoreRecordDocument
	fetchErr    error
}

func (s *stubDataStoreRepo) InsertDataStoreRecord(
	_ context.Context,
	doc *store.NadrfDataStoreRecordDocument,
) error {
	if s.insertErr != nil {
		return s.insertErr
	}
	s.inserted = append(s.inserted, doc)
	return nil
}

func (s *stubDataStoreRepo) ListStoreTransIDsBySnapshot(
	_ context.Context,
	_ string,
	_ time.Time,
	_ time.Time,
	_ time.Time,
) ([]string, error) {
	if s.snapshotErr != nil {
		return nil, s.snapshotErr
	}
	return append([]string(nil), s.snapshotIDs...), nil
}

func (s *stubDataStoreRepo) GetDataStoreRecordByStoreTransID(
	_ context.Context,
	_ string,
) (*store.NadrfDataStoreRecordDocument, error) {
	if s.fetchErr != nil {
		return nil, s.fetchErr
	}
	return s.fetchDoc, nil
}

func TestCreateDataStoreRecordSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubDataStoreRepo{}
	router := newStoreRequestRouter(NewProcessor(repo))

	response := performStoreRequest(router, validStorePayload(), func(req *http.Request) {
		req.Host = retrievalTestAdrfHost
	})

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d, body=%s", response.Code, response.Body.String())
	}

	location := response.Header().Get("Location")
	expectedLocationPrefix := "http://" + retrievalTestAdrfHost + "/nadrf-datamanagement/v1/data-store-records/"
	if !strings.HasPrefix(location, expectedLocationPrefix) {
		t.Fatalf("unexpected Location header: %s", location)
	}

	if len(repo.inserted) != 1 {
		t.Fatalf("expected 1 inserted record, got %d", len(repo.inserted))
	}
	if repo.inserted[0].StoreTransID == "" {
		t.Fatal("expected non-empty storeTransId")
	}
	if repo.inserted[0].Supi != "imsi-001010000000001" {
		t.Fatalf("unexpected SUPI persisted: %s", repo.inserted[0].Supi)
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

func TestCreateDataStoreRecordRejectUnsupportedMediaType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubDataStoreRepo{}
	router := newStoreRequestRouter(NewProcessor(repo))

	response := performStoreRequest(router, validStorePayload(), func(req *http.Request) {
		req.Header.Set("Content-Type", "text/plain")
	})

	assertProblemDetails(t, response, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE")
}

func TestCreateDataStoreRecordRejectMissingContentLength(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubDataStoreRepo{}
	router := newStoreRequestRouter(NewProcessor(repo))

	response := performStoreRequest(router, validStorePayload(), func(req *http.Request) {
		req.ContentLength = -1
		req.TransferEncoding = nil
	})

	assertProblemDetails(t, response, http.StatusLengthRequired, "LENGTH_REQUIRED")
}

func TestCreateDataStoreRecordRejectInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubDataStoreRepo{}
	router := newStoreRequestRouter(NewProcessor(repo))

	response := performStoreRequest(router, "{", nil)

	assertProblemDetails(t, response, http.StatusBadRequest, "INVALID_JSON")
}

func TestCreateDataStoreRecordRejectMissingMandatoryIE(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubDataStoreRepo{}
	router := newStoreRequestRouter(NewProcessor(repo))

	response := performStoreRequest(router, `{"dataSub":[],"dataNotif":{}}`, nil)

	problem := assertProblemDetails(t, response, http.StatusBadRequest, "MANDATORY_IE_MISSING")
	if len(problem.InvalidParams) == 0 {
		t.Fatal("expected invalidParams for mandatory IE errors")
	}
}

func TestCreateDataStoreRecordMapInsertDeadlineExceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubDataStoreRepo{insertErr: context.DeadlineExceeded}
	router := newStoreRequestRouter(NewProcessor(repo))

	response := performStoreRequest(router, validStorePayload(), nil)

	assertProblemDetails(t, response, http.StatusServiceUnavailable, "SYSTEM_FAILURE")
}

func TestCreateDataStoreRecordRejectWhenRepositoryUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := newStoreRequestRouter(NewProcessor(nil))

	response := performStoreRequest(router, validStorePayload(), nil)

	assertProblemDetails(t, response, http.StatusInternalServerError, "SYSTEM_FAILURE")
}

func TestMapJSONBindingErrorToProblemDetailsEOF(t *testing.T) {
	problem := mapJSONBindingErrorToProblemDetails(io.EOF)
	if problem.Status != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", problem.Status)
	}
	if problem.Cause != "MANDATORY_IE_MISSING" {
		t.Fatalf("unexpected cause: %s", problem.Cause)
	}
}

func newStoreRequestRouter(p *Processor) *gin.Engine {
	router := gin.New()
	router.POST(
		factory.AdrfDataManagementResUriPrefix+factory.AdrfDataStoreRecordsPath,
		p.HandleCreateDataStoreRecord,
	)
	return router
}

func performStoreRequest(
	router *gin.Engine,
	body string,
	mutateReq func(*http.Request),
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(
		http.MethodPost,
		factory.AdrfDataManagementResUriPrefix+factory.AdrfDataStoreRecordsPath,
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

func assertProblemDetails(
	t *testing.T,
	response *httptest.ResponseRecorder,
	expectedStatus int,
	expectedCause string,
) models.ProblemDetails {
	t.Helper()

	if response.Code != expectedStatus {
		t.Fatalf(
			"expected status %d, got %d, body=%s",
			expectedStatus,
			response.Code,
			response.Body.String(),
		)
	}

	var problem models.ProblemDetails
	if err := json.Unmarshal(response.Body.Bytes(), &problem); err != nil {
		t.Fatalf("failed to parse ProblemDetails: %v", err)
	}
	if int(problem.Status) != expectedStatus {
		t.Fatalf("expected problem status %d, got %d", expectedStatus, problem.Status)
	}
	if problem.Cause != expectedCause {
		t.Fatalf("expected problem cause %s, got %s", expectedCause, problem.Cause)
	}

	return problem
}

func validStorePayload() string {
	return `{
  "dataSub": [
    {
      "smfDataSub": {
        "supi": "imsi-001010000000001",
        "notifId": "notif-0001",
        "notifUri": "http://nwdaf.local/nnwdaf-events/v1/notifications",
        "notifMethod": "PERIODIC",
        "eventSubs": [
          {
            "event": "UPF_EVENT"
          }
        ]
      }
    }
  ],
  "dataNotif": {
    "upfEventNotifs": [
      {
        "notifId": "notif-0001",
        "upfEvent": "USER_DATA_USAGE_MEASURES",
        "timeStamp": "2026-03-26T10:00:00Z"
      }
    ]
  }
}`
}
