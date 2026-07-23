package service

import (
	"context"
	"time"

	"github.com/free5gc/adrf/internal/logger"
	"github.com/free5gc/adrf/internal/store"
)

type TTLWorker struct {
	dataRepo    *store.DataStoreRepository
	mlModelRepo *store.MLModelRepository
	interval    time.Duration
}

func NewTTLWorker(dataRepo *store.DataStoreRepository, mlModelRepo *store.MLModelRepository, interval time.Duration) *TTLWorker {
	if interval <= 0 {
		interval = 60 * time.Second
	}
	return &TTLWorker{
		dataRepo:    dataRepo,
		mlModelRepo: mlModelRepo,
		interval:    interval,
	}
}

func (w *TTLWorker) Start(ctx context.Context) {
	logger.InitLog.Infof("Starting ADRF TTL background worker (interval: %v)", w.interval)

	go func() {
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.InitLog.Info("ADRF TTL background worker stopped")
				return
			case <-ticker.C:
				w.cleanupExpired(ctx)
			}
		}
	}()
}

func (w *TTLWorker) cleanupExpired(ctx context.Context) {
	now := time.Now().UTC()

	if w.dataRepo != nil {
		deletedData, err := w.dataRepo.DeleteExpiredRecords(ctx, now)
		if err != nil {
			logger.StoreLog.Errorf("TTLWorker cleanupExpired data records error: %v", err)
		} else if deletedData > 0 {
			logger.StoreLog.Infof("TTLWorker purged %d expired data records", deletedData)
		}
	}

	if w.mlModelRepo != nil {
		deletedModels, err := w.mlModelRepo.DeleteExpiredModels(ctx, now)
		if err != nil {
			logger.StoreLog.Errorf("TTLWorker cleanupExpired ML models error: %v", err)
		} else if deletedModels > 0 {
			logger.StoreLog.Infof("TTLWorker purged %d expired ML models", deletedModels)
		}
	}
}
