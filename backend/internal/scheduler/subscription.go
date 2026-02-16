package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/f1-rivals-cup/backend/internal/repository"
)

// SubscriptionScheduler handles automatic expiration of subscriptions
type SubscriptionScheduler struct {
	subscriptionRepo *repository.SubscriptionRepository
	interval         time.Duration
	stopCh           chan struct{}
	stopOnce         sync.Once
}

// NewSubscriptionScheduler creates a new SubscriptionScheduler instance
func NewSubscriptionScheduler(subscriptionRepo *repository.SubscriptionRepository, interval time.Duration) *SubscriptionScheduler {
	return &SubscriptionScheduler{
		subscriptionRepo: subscriptionRepo,
		interval:         interval,
		stopCh:           make(chan struct{}),
	}
}

// Start begins the scheduler loop
func (s *SubscriptionScheduler) Start(ctx context.Context) {
	slog.Info("SubscriptionScheduler started", "interval", s.interval)

	// Run immediately on start
	s.expireSubscriptions(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("SubscriptionScheduler stopping due to context cancellation")
			return
		case <-s.stopCh:
			slog.Info("SubscriptionScheduler stopped")
			return
		case <-ticker.C:
			s.expireSubscriptions(ctx)
		}
	}
}

// Stop signals the scheduler to stop (idempotent)
func (s *SubscriptionScheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
}

func (s *SubscriptionScheduler) expireSubscriptions(ctx context.Context) {
	count, err := s.subscriptionRepo.ExpireSubscriptions(ctx)
	if err != nil {
		slog.Error("SubscriptionScheduler: failed to expire subscriptions", "error", err)
		return
	}
	if count > 0 {
		slog.Info("SubscriptionScheduler: expired subscriptions", "count", count)
	}
}
