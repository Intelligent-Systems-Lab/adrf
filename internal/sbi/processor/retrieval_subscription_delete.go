package processor

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/adrf/internal/logger"
)

func (p *Processor) handleDeleteDataRetrievalSubscription(c *gin.Context) {
	subscriptionID := strings.TrimSpace(c.Param("subscriptionId"))
	if subscriptionID == "" {
		p.writeProblem(c, newProblemDetails(
			http.StatusBadRequest,
			"MANDATORY_IE_MISSING",
			"subscriptionId is required in path",
			nil,
		))
		return
	}

	state, removed := p.deleteRetrievalSubscriptionState(subscriptionID)
	if removed && state != nil && state.DispatchCancel != nil {
		state.DispatchCancel()
	}

	if removed {
		logger.ProcLog.Infof(
			"RetrievalUnsubscribe completed: subscriptionId=%s",
			summarizeIdentifier(subscriptionID),
		)
	} else {
		// ADRF V0 treats already-removed subscription as a successful cleanup.
		logger.ProcLog.Debugf(
			"RetrievalUnsubscribe no-op: subscriptionId=%s already removed",
			summarizeIdentifier(subscriptionID),
		)
	}

	c.Status(http.StatusNoContent)
}
