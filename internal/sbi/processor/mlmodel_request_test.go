package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/free5gc/adrf/internal/store"
	"github.com/free5gc/adrf/pkg/factory"
)

type stubMLModelRepo struct {
	insertErr   error
	inserted    []*store.MLModelStoreRecordDocument
	fetchDoc    *store.MLModelStoreRecordDocument
	fetchErr    error
	listDocs    []*store.MLModelStoreRecordDocument
	listErr     error
	updateErr   error
	deleteErr   error
}

func (s *stubMLModelRepo) InsertMLModelStoreRecord(_ context.Context, doc *store.MLModelStoreRecordDocument) error {
	if s.insertErr != nil {
		return s.insertErr
	}
	s.inserted = append(s.inserted, doc)
	return nil
}

func (s *stubMLModelRepo) GetMLModelStoreRecord(_ context.Context, storeTransId string) (*store.MLModelStoreRecordDocument, error) {
	if s.fetchErr != nil {
		return nil, s.fetchErr
	}
	if s.fetchDoc != nil && s.fetchDoc.StoreTransID == storeTransId {
		return s.fetchDoc, nil
	}
	for _, doc := range s.inserted {
		if doc.StoreTransID == storeTransId {
			return doc, nil
		}
	}
	return nil, nil
}

func (s *stubMLModelRepo) GetMLModelStoreRecords(_ context.Context, modelUniqueIds []string) ([]*store.MLModelStoreRecordDocument, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	if len(modelUniqueIds) == 0 {
		return s.inserted, nil
	}
	var res []*store.MLModelStoreRecordDocument
	for _, doc := range s.inserted {
		for _, uid := range modelUniqueIds {
			if doc.ModelUniqueID == uid {
				res = append(res, doc)
				break
			}
		}
	}
	return res, nil
}

func (s *stubMLModelRepo) UpdateMLModelStoreRecord(_ context.Context, storeTransId string, doc *store.MLModelStoreRecordDocument) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	for idx, d := range s.inserted {
		if d.StoreTransID == storeTransId {
			s.inserted[idx] = doc
			return nil
		}
	}
	return mongo.ErrNoDocuments
}

func (s *stubMLModelRepo) DeleteMLModelStoreRecord(_ context.Context, storeTransId string) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	for idx, d := range s.inserted {
		if d.StoreTransID == storeTransId {
			s.inserted = append(s.inserted[:idx], s.inserted[idx+1:]...)
			return nil
		}
	}
	return mongo.ErrNoDocuments
}

func TestMLModelStoreAndDownloadFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Step 1: Mock remote file server (simulating OpenDaisy task completion download endpoint)
	dummyContent := "mock tarball package content"
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/gzip")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(dummyContent))
	}))
	defer mockServer.Close()

	// Step 2: Initialize Processor with stub db repository
	repo := &stubMLModelRepo{}
	proc := NewProcessor(nil)
	proc.SetMLModelRepo(repo)

	// Set temp directory for model files during tests
	if err := os.MkdirAll("./storage/models", 0755); err != nil {
		t.Fatalf("failed to create storage/models temp directory: %v", err)
	}
	defer os.RemoveAll("./storage")

	// Step 3: Test POST /mlmodel-store-records
	router := gin.New()
	router.POST(
		factory.AdrfMLModelManagementResUriPrefix+factory.AdrfMLModelStoreRecordsPath,
		proc.HandleCreateMLModelStoreRecord,
	)
	router.GET(
		factory.AdrfMLModelManagementResUriPrefix+factory.AdrfMLModelStoreRecordsPath,
		proc.HandleGetMLModelStoreRecords,
	)
	router.GET(
		factory.AdrfMLModelManagementResUriPrefix+factory.AdrfMLModelStoreRecordsPath+"/:storeTransId",
		proc.HandleGetIndividualMLModelStoreRecord,
	)
	router.GET(
		factory.AdrfMLModelManagementResUriPrefix+factory.AdrfMLModelStoreRecordsPath+"/:storeTransId/model",
		proc.HandleDownloadMLModelFile,
	)
	router.PUT(
		factory.AdrfMLModelManagementResUriPrefix+factory.AdrfMLModelStoreRecordsPath+"/:storeTransId",
		proc.HandleUpdateIndividualMLModelStoreRecord,
	)
	router.DELETE(
		factory.AdrfMLModelManagementResUriPrefix+factory.AdrfMLModelStoreRecordsPath+"/:storeTransId",
		proc.HandleDeleteIndividualMLModelStoreRecord,
	)

	postPayload := fmt.Sprintf(`{
		"nfInstanceId": "nwdaf-instance-001",
		"mlModelInfo": [{
			"modelUniqueId": "uuid-1234-5678",
			"mlFileAddr": "%s/download/uuid-1234-5678"
		}]
	}`, mockServer.URL)

	req := httptest.NewRequest(
		http.MethodPost,
		factory.AdrfMLModelManagementResUriPrefix+factory.AdrfMLModelStoreRecordsPath,
		strings.NewReader(postPayload),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d, body=%s", w.Code, w.Body.String())
	}

	location := w.Header().Get("Location")
	if location == "" {
		t.Fatal("expected non-empty Location header in response")
	}

	var res NadrfMLModelStoreRecord
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if res.StoreResult != "ML_MODEL_FILE_STORED_IN_ADRF" {
		t.Fatalf("unexpected StoreResult: %s", res.StoreResult)
	}

	if len(repo.inserted) != 1 {
		t.Fatalf("expected 1 record in repository, got %d", len(repo.inserted))
	}

	doc := repo.inserted[0]
	if doc.StoreResult != "ML_MODEL_FILE_STORED_IN_ADRF" {
		t.Fatalf("unexpected db record StoreResult: %s", doc.StoreResult)
	}
	if doc.NfInstanceID != "nwdaf-instance-001" {
		t.Fatalf("expected NfInstanceID nwdaf-instance-001, got %s", doc.NfInstanceID)
	}

	// Verify local file exists and content matches original mock payload
	filePath := filepath.Join("./storage/models", fmt.Sprintf("%s.tar.gz", doc.StoreTransID))
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read persisted model file: %v", err)
	}
	if string(content) != dummyContent {
		t.Fatalf("persisted model content mismatch: expected %q, got %q", dummyContent, string(content))
	}

	// Step 4: Test GET /mlmodel-store-records
	reqGetList := httptest.NewRequest(
		http.MethodGet,
		factory.AdrfMLModelManagementResUriPrefix+factory.AdrfMLModelStoreRecordsPath+"?model-unique-ids=uuid-1234-5678",
		nil,
	)
	wGetList := httptest.NewRecorder()
	router.ServeHTTP(wGetList, reqGetList)

	if wGetList.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", wGetList.Code)
	}

	var list []NadrfMLModelStoreRecord
	if err := json.Unmarshal(wGetList.Body.Bytes(), &list); err != nil {
		t.Fatalf("failed to parse list response: %v", err)
	}
	if len(list) != 1 || list[0].MlModelInfo[0].ModelUniqueId != "uuid-1234-5678" {
		t.Fatalf("unexpected list size or item: %+v", list)
	}

	// Step 5: Test GET /mlmodel-store-records/:storeTransId
	reqGetInd := httptest.NewRequest(
		http.MethodGet,
		factory.AdrfMLModelManagementResUriPrefix+factory.AdrfMLModelStoreRecordsPath+"/"+doc.StoreTransID,
		nil,
	)
	wGetInd := httptest.NewRecorder()
	router.ServeHTTP(wGetInd, reqGetInd)

	if wGetInd.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", wGetInd.Code)
	}

	// Step 6: Test GET /mlmodel-store-records/:storeTransId/model (Download model file)
	reqDownload := httptest.NewRequest(
		http.MethodGet,
		factory.AdrfMLModelManagementResUriPrefix+factory.AdrfMLModelStoreRecordsPath+"/"+doc.StoreTransID+"/model",
		nil,
	)
	wDownload := httptest.NewRecorder()
	router.ServeHTTP(wDownload, reqDownload)

	if wDownload.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", wDownload.Code)
	}
	if wDownload.Header().Get("Content-Type") != "application/gzip" {
		t.Fatalf("expected Content-Type application/gzip, got %s", wDownload.Header().Get("Content-Type"))
	}
	if wDownload.Body.String() != dummyContent {
		t.Fatalf("download content mismatch: expected %q, got %q", dummyContent, wDownload.Body.String())
	}

	// Step 7: Test PUT /mlmodel-store-records/:storeTransId
	putPayload := fmt.Sprintf(`{
		"nfInstanceId": "nwdaf-instance-001",
		"mlModelInfo": [{
			"modelUniqueId": "uuid-9999-9999",
			"mlFileAddr": "http://updated-addr/model",
			"mlStorageSize": 9999
		}]
	}`)
	reqPut := httptest.NewRequest(
		http.MethodPut,
		factory.AdrfMLModelManagementResUriPrefix+factory.AdrfMLModelStoreRecordsPath+"/"+doc.StoreTransID,
		strings.NewReader(putPayload),
	)
	reqPut.Header.Set("Content-Type", "application/json")
	wPut := httptest.NewRecorder()
	router.ServeHTTP(wPut, reqPut)

	if wPut.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", wPut.Code)
	}

	updatedDoc, _ := repo.GetMLModelStoreRecord(context.Background(), doc.StoreTransID)
	if updatedDoc.ModelUniqueID != "uuid-9999-9999" || updatedDoc.MlFileAddr != "http://updated-addr/model" {
		t.Fatalf("model record update was not persisted: %+v", updatedDoc)
	}

	// Step 8: Test DELETE /mlmodel-store-records/:storeTransId
	reqDel := httptest.NewRequest(
		http.MethodDelete,
		factory.AdrfMLModelManagementResUriPrefix+factory.AdrfMLModelStoreRecordsPath+"/"+doc.StoreTransID,
		nil,
	)
	wDel := httptest.NewRecorder()
	router.ServeHTTP(wDel, reqDel)

	if wDel.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content, got %d", wDel.Code)
	}

	if len(repo.inserted) != 0 {
		t.Fatalf("expected record to be deleted, got %d remaining records", len(repo.inserted))
	}

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("model file still exists on disk after deletion")
	}
}

func TestMLModelStoreDownloadFailed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Step 1: Mock a server returning 404
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	repo := &stubMLModelRepo{}
	proc := NewProcessor(nil)
	proc.SetMLModelRepo(repo)

	router := gin.New()
	router.POST(
		factory.AdrfMLModelManagementResUriPrefix+factory.AdrfMLModelStoreRecordsPath,
		proc.HandleCreateMLModelStoreRecord,
	)

	postPayload := fmt.Sprintf(`{
		"nfInstanceId": "nwdaf-instance-001",
		"mlModelInfo": [{
			"modelUniqueId": "uuid-fail",
			"mlFileAddr": "%s/download/uuid-fail"
		}]
	}`, mockServer.URL)

	req := httptest.NewRequest(
		http.MethodPost,
		factory.AdrfMLModelManagementResUriPrefix+factory.AdrfMLModelStoreRecordsPath,
		strings.NewReader(postPayload),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d, body=%s", w.Code, w.Body.String())
	}

	var res NadrfMLModelStoreRecord
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res.StoreResult != "ML_MODEL_FILE_DOWNLOAD_FAILED" {
		t.Fatalf("expected failure store result, got: %s", res.StoreResult)
	}

	// Clean up storage directory
	_ = os.RemoveAll("./storage")
}

func TestMLModelStoreValidationRules(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubMLModelRepo{}
	proc := NewProcessor(nil)
	proc.SetMLModelRepo(repo)

	router := gin.New()
	router.POST(
		factory.AdrfMLModelManagementResUriPrefix+factory.AdrfMLModelStoreRecordsPath,
		proc.HandleCreateMLModelStoreRecord,
	)

	// Case 1: Missing both nfInstanceId and nfSetId -> expect 400 Bad Request
	payloadNoNf := `{
		"mlModelInfo": [{
			"modelUniqueId": "uuid-001",
			"mlFileAddr": "http://localhost/model"
		}]
	}`
	req1 := httptest.NewRequest(http.MethodPost, factory.AdrfMLModelManagementResUriPrefix+factory.AdrfMLModelStoreRecordsPath, strings.NewReader(payloadNoNf))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request when missing nfInstanceId/nfSetId, got %d", w1.Code)
	}

	// Case 2: Providing both nfInstanceId and nfSetId -> expect 400 Bad Request
	payloadBothNf := `{
		"nfInstanceId": "nwdaf-001",
		"nfSetId": "nwdaf-set-001",
		"mlModelInfo": [{
			"modelUniqueId": "uuid-001",
			"mlFileAddr": "http://localhost/model"
		}]
	}`
	req2 := httptest.NewRequest(http.MethodPost, factory.AdrfMLModelManagementResUriPrefix+factory.AdrfMLModelStoreRecordsPath, strings.NewReader(payloadBothNf))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request when providing both nfInstanceId and nfSetId, got %d", w2.Code)
	}

	// Case 3: Providing only nfSetId -> expect 201 Created
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("mock content"))
	}))
	defer mockServer.Close()

	payloadNfSet := fmt.Sprintf(`{
		"nfSetId": "nwdaf-set-001",
		"mlModelInfo": [{
			"modelUniqueId": "uuid-set-001",
			"mlFileAddr": "%s/model"
		}]
	}`, mockServer.URL)
	req3 := httptest.NewRequest(http.MethodPost, factory.AdrfMLModelManagementResUriPrefix+factory.AdrfMLModelStoreRecordsPath, strings.NewReader(payloadNfSet))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created with nfSetId, got %d, body=%s", w3.Code, w3.Body.String())
	}
	_ = os.RemoveAll("./storage")
}
