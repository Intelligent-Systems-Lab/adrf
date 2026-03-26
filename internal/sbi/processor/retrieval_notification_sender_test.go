package processor

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
)

func TestChunkFetchCorrIDs(t *testing.T) {
	batches := chunkFetchCorrIDs([]string{"id-1", "id-2", "id-3", "id-4", "id-5"}, 2)
	if len(batches) != 3 {
		t.Fatalf("expected 3 batches, got %d", len(batches))
	}
	if len(batches[0].FetchCorrIDs) != 2 || batches[0].TerminationReq {
		t.Fatal("unexpected first batch")
	}
	if len(batches[1].FetchCorrIDs) != 2 || batches[1].TerminationReq {
		t.Fatal("unexpected second batch")
	}
	if len(batches[2].FetchCorrIDs) != 1 || !batches[2].TerminationReq {
		t.Fatal("unexpected last batch")
	}
}

func TestChunkFetchCorrIDsEmptyInput(t *testing.T) {
	batches := chunkFetchCorrIDs(nil, 10)
	if len(batches) != 1 {
		t.Fatalf("expected one termination batch, got %d", len(batches))
	}
	if len(batches[0].FetchCorrIDs) != 0 {
		t.Fatal("expected empty fetch IDs in termination batch")
	}
	if !batches[0].TerminationReq {
		t.Fatal("expected terminationReq=true for empty input")
	}
}

func TestSendFetchInstructionsBatchedAndTerminated(t *testing.T) {
	roundTripper := &captureRoundTripper{}
	sender := &httpRetrievalNotificationSender{
		httpClient: &http.Client{Transport: roundTripper},
	}

	err := sender.SendFetchInstructions(context.Background(), retrievalNotifyRequest{
		NotificationURI: "http://nwdaf.local/adrf/retrieval-notify",
		NotifCorrID:     "retrain-job-001",
		FetchURI:        "http://adrf.local/nadrf-datamanagement/v1/data-store-records",
		FetchCorrIDs:    []string{"st-1", "st-2", "st-3", "st-4", "st-5"},
		CorrIDBatchSize: 2,
	})
	if err != nil {
		t.Fatalf("unexpected send error: %v", err)
	}

	roundTripper.mu.Lock()
	defer roundTripper.mu.Unlock()
	if len(roundTripper.received) != 3 {
		t.Fatalf("expected 3 callback requests, got %d", len(roundTripper.received))
	}

	if roundTripper.received[0].NotifCorrID != "retrain-job-001" || roundTripper.received[0].TerminationReq {
		t.Fatal("unexpected first callback metadata")
	}
	if len(roundTripper.received[0].FetchInstruct.FetchCorrIDs) != 2 {
		t.Fatal("expected first callback to carry 2 IDs")
	}
	if roundTripper.received[1].TerminationReq {
		t.Fatal("expected second callback terminationReq=false")
	}
	if len(roundTripper.received[1].FetchInstruct.FetchCorrIDs) != 2 {
		t.Fatal("expected second callback to carry 2 IDs")
	}
	if !roundTripper.received[2].TerminationReq {
		t.Fatal("expected last callback terminationReq=true")
	}
	if len(roundTripper.received[2].FetchInstruct.FetchCorrIDs) != 1 {
		t.Fatal("expected last callback to carry 1 ID")
	}
}

type captureRoundTripper struct {
	mu       sync.Mutex
	received []retrievalFetchNotificationPayload
}

func (c *captureRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method != http.MethodPost {
		return nil, io.ErrUnexpectedEOF
	}
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		return nil, io.ErrUnexpectedEOF
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	if err = req.Body.Close(); err != nil {
		return nil, err
	}

	var payload retrievalFetchNotificationPayload
	unmarshalErr := json.Unmarshal(body, &payload)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	c.mu.Lock()
	c.received = append(c.received, payload)
	c.mu.Unlock()

	return &http.Response{
		StatusCode: http.StatusNoContent,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
	}, nil
}
