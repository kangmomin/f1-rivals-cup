package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/f1-rivals-cup/backend/internal/repository"
)

func handleLeagueAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate, leagueRepo *repository.LeagueRepository, focused *discordgo.ApplicationCommandInteractionDataOption) {
	ctx := context.Background()
	input := strings.ToLower(focused.StringValue())

	leagues, _, err := leagueRepo.List(ctx, 1, 25, "")
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionApplicationCommandAutocompleteResult,
			Data: &discordgo.InteractionResponseData{},
		})
		return
	}

	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(leagues))
	for _, l := range leagues {
		if input != "" && !strings.Contains(strings.ToLower(l.Name), input) {
			continue
		}
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  fmt.Sprintf("%s (시즌 %d)", l.Name, l.Season),
			Value: l.ID.String(),
		})
		if len(choices) >= 25 {
			break
		}
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
}

func handleMatchAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate, leagueRepo *repository.LeagueRepository, matchRepo *repository.MatchRepository, focused *discordgo.ApplicationCommandInteractionDataOption) {
	ctx := context.Background()
	input := strings.ToLower(focused.StringValue())

	leagues, _, err := leagueRepo.List(ctx, 1, 25, "")
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionApplicationCommandAutocompleteResult,
			Data: &discordgo.InteractionResponseData{},
		})
		return
	}

	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0)
	for _, l := range leagues {
		matches, err := matchRepo.ListByLeague(ctx, l.ID)
		if err != nil {
			continue
		}
		for _, m := range matches {
			label := fmt.Sprintf("Round %d - %s (%s)", m.Round, m.Track, l.Name)
			if input != "" && !strings.Contains(strings.ToLower(label), input) {
				continue
			}
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  truncate(label, 100),
				Value: m.ID.String(),
			})
			if len(choices) >= 25 {
				break
			}
		}
		if len(choices) >= 25 {
			break
		}
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
}
