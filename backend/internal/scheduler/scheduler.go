package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/f1-rivals-cup/backend/internal/model"
	"github.com/f1-rivals-cup/backend/internal/repository"
)

// MatchScheduler handles automatic match status updates
type MatchScheduler struct {
	matchRepo *repository.MatchRepository
	interval  time.Duration
	stopCh    chan struct{}
	stopOnce  sync.Once
	location  *time.Location
}

// New creates a new MatchScheduler instance
func New(matchRepo *repository.MatchRepository, interval time.Duration) *MatchScheduler {
	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		slog.Warn("Failed to load Asia/Seoul timezone, using UTC", "error", err)
		loc = time.UTC
	}

	return &MatchScheduler{
		matchRepo: matchRepo,
		interval:  interval,
		stopCh:    make(chan struct{}),
		location:  loc,
	}
}

// Start begins the scheduler loop
func (s *MatchScheduler) Start(ctx context.Context) {
	slog.Info("MatchScheduler started", "interval", s.interval)

	// Run immediately on start
	s.checkAndUpdateMatches(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("MatchScheduler stopping due to context cancellation")
			return
		case <-s.stopCh:
			slog.Info("MatchScheduler stopped")
			return
		case <-ticker.C:
			s.checkAndUpdateMatches(ctx)
		}
	}
}

// Stop signals the scheduler to stop (idempotent)
func (s *MatchScheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
}

// checkAndUpdateMatches checks upcoming matches and updates their status if needed
func (s *MatchScheduler) checkAndUpdateMatches(ctx context.Context) {
	now := time.Now().In(s.location)
	slog.Debug("MatchScheduler checking matches", "time", now)

	matches, err := s.matchRepo.ListUpcomingMatches(ctx)
	if err != nil {
		slog.Error("MatchScheduler: failed to list upcoming matches", "error", err)
		return
	}

	if len(matches) == 0 {
		slog.Debug("MatchScheduler: no upcoming matches found")
		return
	}

	for _, match := range matches {
		matchTime, err := s.parseMatchDateTime(match.MatchDate, match.MatchTime)
		if err != nil {
			slog.Warn("MatchScheduler: failed to parse match datetime",
				"match_id", match.ID,
				"match_date", match.MatchDate,
				"match_time", match.MatchTime,
				"error", err)
			continue
		}

		if now.After(matchTime) || now.Equal(matchTime) {
			if err := s.matchRepo.UpdateStatus(ctx, match.ID, model.MatchStatusInProgress); err != nil {
				slog.Error("MatchScheduler: failed to update match status",
					"match_id", match.ID,
					"error", err)
				continue
			}
			slog.Info("MatchScheduler: match status updated to in_progress",
				"match_id", match.ID,
				"track", match.Track,
				"round", match.Round,
				"match_time", matchTime)
		}
	}
}

// parseMatchDateTime parses date (YYYY-MM-DD) and optional time (HH:MM or HH:MM:SS) into time.Time
func (s *MatchScheduler) parseMatchDateTime(dateStr string, timeStr *string) (time.Time, error) {
	// Default time to 00:00:00 if not provided
	timeVal := "00:00:00"
	if timeStr != nil && *timeStr != "" {
		timeVal = *timeStr
	}

	// Parse combined datetime with multiple format support
	combined := dateStr + " " + timeVal

	// Try HH:MM:SS format first
	t, err := time.ParseInLocation("2006-01-02 15:04:05", combined, s.location)
	if err == nil {
		return t, nil
	}

	// Try HH:MM format (no seconds)
	return time.ParseInLocation("2006-01-02 15:04", combined, s.location)
}
