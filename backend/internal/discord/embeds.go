package discord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/f1-rivals-cup/backend/internal/model"
)

const (
	colorF1Red    = 0xE10600
	colorGold     = 0xFFD700
	colorSchedule = 0x00D2BE
	colorInfo     = 0x3498DB
	colorError    = 0xFF4444
)

func buildStandingsEmbed(league *model.League, standings []model.StandingsEntry) *discordgo.MessageEmbed {
	if len(standings) == 0 {
		return &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("🏆 %s - 드라이버 순위", league.Name),
			Description: "등록된 순위 데이터가 없습니다.",
			Color:       colorF1Red,
		}
	}

	var sb strings.Builder
	sb.WriteString("```\n")
	sb.WriteString(fmt.Sprintf(" %-3s | %-16s | %-12s | %5s | %s | %s\n", "#", "Driver", "Team", "Pts", "W", "P"))
	sb.WriteString(fmt.Sprintf("%-4s|%-18s|%-14s|%6s|%3s|%3s\n", "----", "------------------", "--------------", "------", "---", "---"))

	for _, s := range standings {
		team := "-"
		if s.TeamName != nil {
			team = truncate(*s.TeamName, 12)
		}
		sb.WriteString(fmt.Sprintf(" %-3d | %-16s | %-12s | %5.0f | %d | %d\n",
			s.Rank,
			truncate(s.DriverName, 16),
			team,
			s.TotalPoints,
			s.Wins,
			s.Podiums,
		))
	}
	sb.WriteString("```")

	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🏆 %s - 드라이버 순위", league.Name),
		Description: sb.String(),
		Color:       colorF1Red,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("시즌 %d", league.Season),
		},
	}
}

func buildTeamStandingsEmbed(league *model.League, standings []model.TeamStandingsEntry) *discordgo.MessageEmbed {
	if len(standings) == 0 {
		return &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("🏎️ %s - 팀 순위", league.Name),
			Description: "등록된 팀 순위 데이터가 없습니다.",
			Color:       colorGold,
		}
	}

	var sb strings.Builder
	sb.WriteString("```\n")
	sb.WriteString(fmt.Sprintf(" %-3s | %-16s | %5s | %s | %s | %s\n", "#", "Team", "Pts", "W", "P", "Drivers"))
	sb.WriteString(fmt.Sprintf("%-4s|%-18s|%6s|%3s|%3s|%8s\n", "----", "------------------", "------", "---", "---", "--------"))

	for _, s := range standings {
		sb.WriteString(fmt.Sprintf(" %-3d | %-16s | %5.0f | %d | %d | %d\n",
			s.Rank,
			truncate(s.TeamName, 16),
			s.TotalPoints,
			s.Wins,
			s.Podiums,
			s.DriverCount,
		))
	}
	sb.WriteString("```")

	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🏎️ %s - 팀 순위", league.Name),
		Description: sb.String(),
		Color:       colorGold,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("시즌 %d", league.Season),
		},
	}
}

func buildScheduleEmbed(league *model.League, matches []*model.Match) *discordgo.MessageEmbed {
	if len(matches) == 0 {
		return &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("📅 %s - 레이스 일정", league.Name),
			Description: "등록된 일정이 없습니다.",
			Color:       colorSchedule,
		}
	}

	fields := make([]*discordgo.MessageEmbedField, 0, len(matches))
	for _, m := range matches {
		name := fmt.Sprintf("Round %d - %s", m.Round, m.Track)

		timeStr := m.MatchDate
		if m.MatchTime != nil {
			timeStr += " " + *m.MatchTime
		}

		status := statusLabel(m.Status)
		value := fmt.Sprintf("%s | %s", timeStr, status)

		if m.HasSprint {
			sprintTime := ""
			if m.SprintDate != nil {
				sprintTime = *m.SprintDate
			}
			if m.SprintTime != nil {
				sprintTime += " " + *m.SprintTime
			}
			sprintStatus := statusLabel(m.SprintStatus)
			value += fmt.Sprintf("\n🏃 Sprint: %s | %s", sprintTime, sprintStatus)
		}

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  name,
			Value: value,
		})
	}

	return &discordgo.MessageEmbed{
		Title:  fmt.Sprintf("📅 %s - 레이스 일정", league.Name),
		Color:  colorSchedule,
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("시즌 %d | 총 %d 라운드", league.Season, len(matches)),
		},
	}
}

func buildResultsEmbed(match *model.Match, league *model.League, results []*model.MatchResult) *discordgo.MessageEmbed {
	title := fmt.Sprintf("🏁 Round %d - %s", match.Round, match.Track)
	if league != nil {
		title += fmt.Sprintf(" (%s)", league.Name)
	}

	if len(results) == 0 {
		return &discordgo.MessageEmbed{
			Title:       title,
			Description: "등록된 결과가 없습니다.",
			Color:       colorF1Red,
		}
	}

	var sb strings.Builder
	sb.WriteString("```\n")
	sb.WriteString(fmt.Sprintf(" %-3s | %-16s | %-12s | %5s | %s\n", "Pos", "Driver", "Team", "Pts", "FL"))
	sb.WriteString(fmt.Sprintf("%-4s|%-18s|%-14s|%6s|%3s\n", "----", "------------------", "--------------", "------", "---"))

	for _, r := range results {
		pos := "-"
		if r.Position != nil {
			pos = fmt.Sprintf("%d", *r.Position)
		}
		if r.DNF {
			pos = "DNF"
		}

		driver := "-"
		if r.ParticipantName != nil {
			driver = truncate(*r.ParticipantName, 16)
		}

		team := "-"
		if r.TeamName != nil {
			team = truncate(*r.TeamName, 12)
		}

		fl := " "
		if r.FastestLap {
			fl = "⚡"
		}

		totalPts := r.Points + r.SprintPoints
		sb.WriteString(fmt.Sprintf(" %-3s | %-16s | %-12s | %5.0f | %s\n",
			pos, driver, team, totalPts, fl,
		))
	}
	sb.WriteString("```")

	return &discordgo.MessageEmbed{
		Title:       title,
		Description: sb.String(),
		Color:       colorF1Red,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("%s | %s", match.MatchDate, statusLabel(match.Status)),
		},
	}
}

func buildLeaguesEmbed(leagues []*model.League) *discordgo.MessageEmbed {
	if len(leagues) == 0 {
		return &discordgo.MessageEmbed{
			Title:       "📋 리그 목록",
			Description: "등록된 리그가 없습니다.",
			Color:       colorInfo,
		}
	}

	fields := make([]*discordgo.MessageEmbedField, 0, len(leagues))
	for _, l := range leagues {
		value := fmt.Sprintf("시즌 %d | %s", l.Season, statusLabelKorean(l.Status))
		if l.Description != nil && *l.Description != "" {
			value += "\n" + truncate(*l.Description, 100)
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  l.Name,
			Value: value,
		})
	}

	return &discordgo.MessageEmbed{
		Title:  "📋 리그 목록",
		Color:  colorInfo,
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("총 %d개 리그", len(leagues)),
		},
	}
}

func buildLeagueInfoEmbed(league *model.League) *discordgo.MessageEmbed {
	fields := []*discordgo.MessageEmbedField{
		{Name: "시즌", Value: fmt.Sprintf("%d", league.Season), Inline: true},
		{Name: "상태", Value: statusLabelKorean(league.Status), Inline: true},
	}

	if league.StartDate != nil {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name: "시작일", Value: league.StartDate.Format("2006-01-02"), Inline: true,
		})
	}
	if league.EndDate != nil {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name: "종료일", Value: league.EndDate.Format("2006-01-02"), Inline: true,
		})
	}
	if league.MatchTime != nil {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name: "경기 시간", Value: *league.MatchTime, Inline: true,
		})
	}
	if league.ContactInfo != nil {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name: "연락처", Value: *league.ContactInfo,
		})
	}

	desc := ""
	if league.Description != nil {
		desc = *league.Description
	}

	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("ℹ️ %s", league.Name),
		Description: desc,
		Color:       colorInfo,
		Fields:      fields,
	}
}

func buildErrorEmbed(message string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "❌ 오류",
		Description: message,
		Color:       colorError,
	}
}

func statusLabel(status model.MatchStatus) string {
	switch status {
	case model.MatchStatusUpcoming:
		return "[UPCOMING]"
	case model.MatchStatusInProgress:
		return "[LIVE]"
	case model.MatchStatusCompleted:
		return "[COMPLETED]"
	case model.MatchStatusCancelled:
		return "[CANCELLED]"
	default:
		return string(status)
	}
}

func statusLabelKorean(status model.LeagueStatus) string {
	switch status {
	case model.LeagueStatusDraft:
		return "준비중"
	case model.LeagueStatusOpen:
		return "모집중"
	case model.LeagueStatusInProgress:
		return "진행중"
	case model.LeagueStatusCompleted:
		return "완료"
	case model.LeagueStatusCancelled:
		return "취소"
	default:
		return string(status)
	}
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}
