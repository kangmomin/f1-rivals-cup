package discord

import (
	"context"
	"log/slog"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/f1-rivals-cup/backend/internal/repository"
)

// Bot manages the Discord bot session lifecycle.
type Bot struct {
	session    *discordgo.Session
	guildID    string
	handler    *CommandHandler
	commandIDs []string
	stopOnce   sync.Once
	stopCh     chan struct{}
}

// NewBot creates a new Discord bot. Call Start to connect.
func NewBot(token, guildID string, leagueRepo *repository.LeagueRepository, matchRepo *repository.MatchRepository, matchResultRepo *repository.MatchResultRepository) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	handler := NewCommandHandler(leagueRepo, matchRepo, matchResultRepo)

	bot := &Bot{
		session: session,
		guildID: guildID,
		handler: handler,
		stopCh:  make(chan struct{}),
	}

	session.AddHandler(handler.Handle)

	return bot, nil
}

// Start opens the session, registers commands, and blocks until ctx is done or Stop is called.
func (b *Bot) Start(ctx context.Context) {
	if err := b.session.Open(); err != nil {
		slog.Error("Failed to open Discord session", "error", err)
		return
	}

	slog.Info("Discord bot connected", "user", b.session.State.User.Username)

	registered, err := b.session.ApplicationCommandBulkOverwrite(b.session.State.User.ID, b.guildID, commands)
	if err != nil {
		slog.Error("Failed to register Discord commands", "error", err)
		b.session.Close()
		return
	}

	b.commandIDs = make([]string, len(registered))
	for i, cmd := range registered {
		b.commandIDs[i] = cmd.ID
	}
	slog.Info("Discord commands registered", "count", len(registered))

	select {
	case <-ctx.Done():
	case <-b.stopCh:
	}

	b.cleanup()
}

// Stop signals the bot to shut down (idempotent).
func (b *Bot) Stop() {
	b.stopOnce.Do(func() {
		slog.Info("Discord bot shutting down")
		close(b.stopCh)
	})
}

func (b *Bot) cleanup() {
	if b.guildID != "" {
		for _, id := range b.commandIDs {
			if err := b.session.ApplicationCommandDelete(b.session.State.User.ID, b.guildID, id); err != nil {
				slog.Warn("Failed to delete Discord command", "id", id, "error", err)
			}
		}
	}
	b.session.Close()
	slog.Info("Discord bot stopped")
}
