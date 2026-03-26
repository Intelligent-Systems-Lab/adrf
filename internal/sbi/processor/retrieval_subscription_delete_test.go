package processor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/adrf/pkg/factory"
)

func TestDeleteDataRetrievalSubscriptionSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	processor := NewProcessor(&stubDataStoreRepo{})
	cancelCalled := false
	processor.storeRetrievalSubscriptionState(&retrievalSubscriptionState{
		SubscriptionID: "sub-001",
		DispatchCancel: func() { cancelCalled = true },
	})
	router := newRetrievalSubscriptionDeleteRouter(processor)

	response := performRetrievalSubscriptionDeleteRequest(router, "sub-001")
	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", response.Code)
	}
	if !cancelCalled {
		t.Fatal("expected dispatch cancel to be called")
	}

	if _, exists := processor.getRetrievalSubscriptionState("sub-001"); exists {
		t.Fatal("expected subscription state to be removed after delete")
	}
}

func TestDeleteDataRetrievalSubscriptionNoopWhenNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	processor := NewProcessor(&stubDataStoreRepo{})
	router := newRetrievalSubscriptionDeleteRouter(processor)

	response := performRetrievalSubscriptionDeleteRequest(router, "missing-subscription")
	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", response.Code)
	}
}

func TestDeleteDataRetrievalSubscriptionRejectEmptyPathID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	processor := NewProcessor(&stubDataStoreRepo{})
	router := newRetrievalSubscriptionDeleteRouter(processor)

	req := httptest.NewRequest(
		http.MethodDelete,
		factory.AdrfDataManagementResUriPrefix+factory.AdrfDataRetrievalSubscriptionsPath+"/%20%20",
		nil,
	)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	assertProblemDetails(t, response, http.StatusBadRequest, "MANDATORY_IE_MISSING")
}

func newRetrievalSubscriptionDeleteRouter(p *Processor) *gin.Engine {
	router := gin.New()
	router.DELETE(
		factory.AdrfDataManagementResUriPrefix+factory.AdrfDataRetrievalSubscriptionsPath+"/:subscriptionId",
		p.HandleDeleteDataRetrievalSubscription,
	)
	return router
}

func performRetrievalSubscriptionDeleteRequest(
	router *gin.Engine,
	subscriptionID string,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(
		http.MethodDelete,
		factory.AdrfDataManagementResUriPrefix+factory.AdrfDataRetrievalSubscriptionsPath+"/"+subscriptionID,
		nil,
	)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}
