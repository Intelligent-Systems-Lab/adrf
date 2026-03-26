package processor

import (
	"context"
	"time"
)

// retrievalSubscriptionState stores the server-side runtime state of one
// retrieval subscription.
//
// The request body alone is not enough for retrieval execution because ADRF must
// freeze a deterministic snapshot at subscription creation time and keep the
// resulting fetch correlation IDs for later callback delivery.
type retrievalSubscriptionState struct {
	SubscriptionID  string
	NotifCorrID     string
	NotificationURI string
	Supi            string

	TimePeriodStart time.Time
	TimePeriodStop  time.Time
	SnapshotAt      time.Time
	CreatedAt       time.Time

	ConsTrigNotif bool
	FetchCorrIDs  []string

	DispatchCtx    context.Context
	DispatchCancel context.CancelFunc
}

func (p *Processor) storeRetrievalSubscriptionState(state *retrievalSubscriptionState) {
	if state == nil {
		return
	}

	p.retrievalSubMu.Lock()
	defer p.retrievalSubMu.Unlock()
	p.retrievalSubs[state.SubscriptionID] = state
}

func (p *Processor) getRetrievalSubscriptionState(subscriptionID string) (*retrievalSubscriptionState, bool) {
	p.retrievalSubMu.RLock()
	defer p.retrievalSubMu.RUnlock()

	state, ok := p.retrievalSubs[subscriptionID]
	return state, ok
}

func (p *Processor) deleteRetrievalSubscriptionState(subscriptionID string) (*retrievalSubscriptionState, bool) {
	p.retrievalSubMu.Lock()
	defer p.retrievalSubMu.Unlock()

	state, ok := p.retrievalSubs[subscriptionID]
	if !ok {
		return nil, false
	}
	delete(p.retrievalSubs, subscriptionID)
	return state, true
}
