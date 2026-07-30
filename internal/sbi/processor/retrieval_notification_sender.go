package processor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/adrf/internal/logger"
	"github.com/free5gc/adrf/pkg/factory"
)

const (
	defaultCorrIDBatchSize         = 100
	defaultCallbackRequestTimeout  = 5 * time.Second
	retrievalNotifyResponseMaxRead = 2048
)

// retrievalNotificationSender sends retrieval notifications to NF consumers.
//
// This abstraction makes callback behavior testable without external network
// dependencies and keeps request handler logic focused on validation/storage.
type retrievalNotificationSender interface {
	SendFetchInstructions(ctx context.Context, req retrievalNotifyRequest) error
}

// retrievalNotifyRequest captures callback payload context for one subscription.
type retrievalNotifyRequest struct {
	NotificationURI string
	NotifCorrID     string
	FetchURI        string
	FetchCorrIDs    []string
	CorrIDBatchSize int
}

type httpRetrievalNotificationSender struct {
	httpClient *http.Client
}

func newHTTPRetrievalNotificationSender() *httpRetrievalNotificationSender {
	return &httpRetrievalNotificationSender{
		httpClient: &http.Client{Timeout: defaultCallbackRequestTimeout},
	}
}

func (s *httpRetrievalNotificationSender) SendFetchInstructions(
	ctx context.Context,
	req retrievalNotifyRequest,
) error {
	if s == nil || s.httpClient == nil {
		return fmt.Errorf("retrieval notification sender is not initialized")
	}
	if strings.TrimSpace(req.NotificationURI) == "" {
		return fmt.Errorf("notification URI is empty")
	}
	if strings.TrimSpace(req.NotifCorrID) == "" {
		return fmt.Errorf("notification correlation ID is empty")
	}
	if strings.TrimSpace(req.FetchURI) == "" {
		return fmt.Errorf("fetch URI is empty")
	}

	batches := chunkFetchCorrIDs(req.FetchCorrIDs, req.CorrIDBatchSize)
	for batchIdx, batch := range batches {
		payload := retrievalFetchNotificationPayload{
			NotifCorrID: req.NotifCorrID,
			TimeStamp:   time.Now().UTC(),
			FetchInstruct: retrievalFetchInstruction{
				FetchURI:     req.FetchURI,
				FetchCorrIDs: batch.FetchCorrIDs,
			},
			TerminationReq: batch.TerminationReq,
		}

		bodyBytes, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal retrieval notification payload: %w", marshalErr)
		}

		httpReq, reqErr := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			req.NotificationURI,
			bytes.NewReader(bodyBytes),
		)
		if reqErr != nil {
			return fmt.Errorf("failed to build retrieval notification request: %w", reqErr)
		}
		httpReq.Header.Set("Content-Type", "application/json")

		httpResp, sendErr := s.httpClient.Do(httpReq)
		if sendErr != nil {
			return fmt.Errorf("failed to send retrieval notification callback: %w", sendErr)
		}

		respBody, readErr := io.ReadAll(io.LimitReader(httpResp.Body, retrievalNotifyResponseMaxRead))
		closeErr := httpResp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("failed to read callback response body: %w", readErr)
		}
		if closeErr != nil {
			return fmt.Errorf("failed to close callback response body: %w", closeErr)
		}

		if httpResp.StatusCode < http.StatusOK || httpResp.StatusCode >= http.StatusMultipleChoices {
			return fmt.Errorf(
				"retrieval notification callback returned status=%d body=%q",
				httpResp.StatusCode,
				string(respBody),
			)
		}

		logger.ProcLog.Debugf(
			"RetrievalNotify callback sent: notifCorrId=%s batch=%d/%d corrIds=%d terminationReq=%t status=%d",
			summarizeIdentifier(req.NotifCorrID),
			batchIdx+1,
			len(batches),
			len(batch.FetchCorrIDs),
			batch.TerminationReq,
			httpResp.StatusCode,
		)
	}

	return nil
}

type retrievalFetchInstruction struct {
	FetchURI     string   `json:"fetchUri"`
	FetchCorrIDs []string `json:"fetchCorrIds"`
}

type retrievalFetchNotificationPayload struct {
	NotifCorrID    string                    `json:"notifCorrId"`
	FetchInstruct  retrievalFetchInstruction `json:"fetchInstruct"`
	TerminationReq bool                      `json:"terminationReq,omitempty"`
	TimeStamp      time.Time                 `json:"timeStamp"`
}

type fetchCorrIDBatch struct {
	FetchCorrIDs   []string
	TerminationReq bool
}

// chunkFetchCorrIDs splits correlation IDs into callback batches and marks the
// last batch with terminationReq=true.
//
// When there is no data to fetch, ADRF still sends one terminating callback so
// the consumer can close the retrieval workflow deterministically.
func chunkFetchCorrIDs(fetchCorrIDs []string, batchSize int) []fetchCorrIDBatch {
	if batchSize <= 0 {
		batchSize = defaultCorrIDBatchSize
	}

	if len(fetchCorrIDs) == 0 {
		return []fetchCorrIDBatch{
			{FetchCorrIDs: []string{}, TerminationReq: true},
		}
	}

	batches := make([]fetchCorrIDBatch, 0, (len(fetchCorrIDs)+batchSize-1)/batchSize)
	for start := 0; start < len(fetchCorrIDs); start += batchSize {
		end := start + batchSize
		if end > len(fetchCorrIDs) {
			end = len(fetchCorrIDs)
		}
		batchIDs := append([]string(nil), fetchCorrIDs[start:end]...)
		batches = append(batches, fetchCorrIDBatch{FetchCorrIDs: batchIDs})
	}

	batches[len(batches)-1].TerminationReq = true
	return batches
}

func resolveCorrIDBatchSize() int {
	if factory.AdrfConfig != nil &&
		factory.AdrfConfig.Configuration != nil &&
		factory.AdrfConfig.Configuration.Retrieval != nil &&
		factory.AdrfConfig.Configuration.Retrieval.CorrIDBatchSize > 0 {
		return factory.AdrfConfig.Configuration.Retrieval.CorrIDBatchSize
	}
	return defaultCorrIDBatchSize
}

func buildDataStoreRecordsFetchURI(c *gin.Context) string {
	path := fmt.Sprintf("%s%s", factory.AdrfDataManagementResUriPrefix, factory.AdrfDataStoreRecordsPath)
	if c == nil || c.Request == nil || c.Request.Host == "" {
		return path
	}
	return fmt.Sprintf("%s://%s%s", requestScheme(c.Request), c.Request.Host, path)
}

func (p *Processor) dispatchRetrievalFetchNotifications(
	dispatchCtx context.Context,
	state *retrievalSubscriptionState,
	fetchURI string,
) {
	if p == nil || state == nil {
		return
	}
	if p.retrievalNotifier == nil {
		logger.ProcLog.Warnf(
			"RetrievalNotify dispatch skipped: notifier is nil (subscriptionId=%s)",
			state.SubscriptionID,
		)
		return
	}

	effectiveFetchURI := fetchURI
	if state.DatasetURL != "" {
		effectiveFetchURI = state.DatasetURL
	}

	notifyReq := retrievalNotifyRequest{
		NotificationURI: state.NotificationURI,
		NotifCorrID:     state.NotifCorrID,
		FetchURI:        effectiveFetchURI,
		FetchCorrIDs:    append([]string(nil), state.FetchCorrIDs...),
		CorrIDBatchSize: resolveCorrIDBatchSize(),
	}

	logger.ProcLog.Infof(
		"RetrievalNotify dispatch started: subscriptionId=%s notifCorrId=%s batchesBy=%d corrIds=%d",
		summarizeIdentifier(state.SubscriptionID),
		summarizeIdentifier(state.NotifCorrID),
		notifyReq.CorrIDBatchSize,
		len(notifyReq.FetchCorrIDs),
	)

	if err := p.retrievalNotifier.SendFetchInstructions(dispatchCtx, notifyReq); err != nil {
		if errors.Is(err, context.Canceled) {
			logger.ProcLog.Infof(
				"RetrievalNotify dispatch canceled: subscriptionId=%s notifCorrId=%s",
				summarizeIdentifier(state.SubscriptionID),
				summarizeIdentifier(state.NotifCorrID),
			)
			return
		}

		logger.ProcLog.Errorf(
			"RetrievalNotify dispatch failed: subscriptionId=%s notifCorrId=%s err=%v",
			summarizeIdentifier(state.SubscriptionID),
			summarizeIdentifier(state.NotifCorrID),
			err,
		)
		return
	}

	logger.ProcLog.Infof(
		"RetrievalNotify dispatch completed: subscriptionId=%s notifCorrId=%s",
		summarizeIdentifier(state.SubscriptionID),
		summarizeIdentifier(state.NotifCorrID),
	)
}
