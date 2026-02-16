package discord

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/f1-rivals-cup/backend/internal/repository"
	"github.com/google/uuid"
)

// CommandHandler dispatches slash commands to individual handlers.
type CommandHandler struct {
	leagueRepo      *repository.LeagueRepository
	matchRepo       *repository.MatchRepository
	matchResultRepo *repository.MatchResultRepository
}

// NewCommandHandler creates a new CommandHandler.
func NewCommandHandler(leagueRepo *repository.LeagueRepository, matchRepo *repository.MatchRepository, matchResultRepo *repository.MatchResultRepository) *CommandHandler {
	return &CommandHandler{
		leagueRepo:      leagueRepo,
		matchRepo:       matchRepo,
		matchResultRepo: matchResultRepo,
	}
}

// Handle is the top-level InteractionCreate handler.
func (h *CommandHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Discord handler panic recovered", "recover", r)
			respondError(s, i, "내부 오류가 발생했습니다.")
		}
	}()

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		h.handleCommand(s, i)
	case discordgo.InteractionApplicationCommandAutocomplete:
		h.handleAutocomplete(s, i)
	}
}

func (h *CommandHandler) handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "standings":
		h.handleStandings(s, i)
	case "team-standings":
		h.handleTeamStandings(s, i)
	case "schedule":
		h.handleSchedule(s, i)
	case "results":
		h.handleResults(s, i)
	case "leagues":
		h.handleLeagues(s, i)
	case "league-info":
		h.handleLeagueInfo(s, i)
	}
}

func (h *CommandHandler) handleAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	for _, opt := range data.Options {
		if !opt.Focused {
			continue
		}
		switch opt.Name {
		case "league":
			handleLeagueAutocomplete(s, i, h.leagueRepo, opt)
		case "match":
			handleMatchAutocomplete(s, i, h.leagueRepo, h.matchRepo, opt)
		}
		return
	}
}

func (h *CommandHandler) handleStandings(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()
	leagueID, err := parseLeagueOption(i)
	if err != nil {
		respondError(s, i, "잘못된 리그 ID입니다.")
		return
	}

	league, err := h.leagueRepo.GetByID(ctx, leagueID)
	if err != nil {
		respondError(s, i, "리그를 찾을 수 없습니다.")
		return
	}

	standings, err := h.matchResultRepo.GetLeagueStandings(ctx, leagueID)
	if err != nil {
		respondError(s, i, "순위 데이터를 불러올 수 없습니다.")
		return
	}

	respondEmbed(s, i, buildStandingsEmbed(league, standings))
}

func (h *CommandHandler) handleTeamStandings(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()
	leagueID, err := parseLeagueOption(i)
	if err != nil {
		respondError(s, i, "잘못된 리그 ID입니다.")
		return
	}

	league, err := h.leagueRepo.GetByID(ctx, leagueID)
	if err != nil {
		respondError(s, i, "리그를 찾을 수 없습니다.")
		return
	}

	standings, err := h.matchResultRepo.GetTeamStandings(ctx, leagueID)
	if err != nil {
		respondError(s, i, "팀 순위 데이터를 불러올 수 없습니다.")
		return
	}

	respondEmbed(s, i, buildTeamStandingsEmbed(league, standings))
}

func (h *CommandHandler) handleSchedule(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()
	leagueID, err := parseLeagueOption(i)
	if err != nil {
		respondError(s, i, "잘못된 리그 ID입니다.")
		return
	}

	league, err := h.leagueRepo.GetByID(ctx, leagueID)
	if err != nil {
		respondError(s, i, "리그를 찾을 수 없습니다.")
		return
	}

	matches, err := h.matchRepo.ListByLeague(ctx, leagueID)
	if err != nil {
		respondError(s, i, "일정 데이터를 불러올 수 없습니다.")
		return
	}

	respondEmbed(s, i, buildScheduleEmbed(league, matches))
}

func (h *CommandHandler) handleResults(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()
	opts := i.ApplicationCommandData().Options
	if len(opts) == 0 {
		respondError(s, i, "매치를 선택해주세요.")
		return
	}

	matchID, err := uuid.Parse(opts[0].StringValue())
	if err != nil {
		respondError(s, i, "잘못된 매치 ID입니다.")
		return
	}

	match, err := h.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		respondError(s, i, "매치를 찾을 수 없습니다.")
		return
	}

	results, err := h.matchResultRepo.ListByMatch(ctx, matchID)
	if err != nil {
		respondError(s, i, "결과 데이터를 불러올 수 없습니다.")
		return
	}

	league, _ := h.leagueRepo.GetByID(ctx, match.LeagueID)

	respondEmbed(s, i, buildResultsEmbed(match, league, results))
}

func (h *CommandHandler) handleLeagues(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()
	leagues, _, err := h.leagueRepo.List(ctx, 1, 25, "")
	if err != nil {
		respondError(s, i, "리그 목록을 불러올 수 없습니다.")
		return
	}

	respondEmbed(s, i, buildLeaguesEmbed(leagues))
}

func (h *CommandHandler) handleLeagueInfo(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()
	leagueID, err := parseLeagueOption(i)
	if err != nil {
		respondError(s, i, "잘못된 리그 ID입니다.")
		return
	}

	league, err := h.leagueRepo.GetByID(ctx, leagueID)
	if err != nil {
		respondError(s, i, "리그를 찾을 수 없습니다.")
		return
	}

	respondEmbed(s, i, buildLeagueInfoEmbed(league))
}

// parseLeagueOption extracts the league UUID from the first command option.
func parseLeagueOption(i *discordgo.InteractionCreate) (uuid.UUID, error) {
	opts := i.ApplicationCommandData().Options
	if len(opts) == 0 {
		return uuid.Nil, fmt.Errorf("no league option")
	}
	return uuid.Parse(opts[0].StringValue())
}

func respondEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
	if err != nil {
		slog.Error("Failed to respond to interaction", "error", err)
	}
}

func respondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{buildErrorEmbed(message)},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		slog.Error("Failed to respond with error", "error", err)
	}
}
