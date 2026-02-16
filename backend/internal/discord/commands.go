package discord

import "github.com/bwmarrin/discordgo"

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "standings",
		Description: "드라이버 순위 조회",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:         discordgo.ApplicationCommandOptionString,
				Name:         "league",
				Description:  "리그 선택",
				Required:     true,
				Autocomplete: true,
			},
		},
	},
	{
		Name:        "team-standings",
		Description: "팀 순위 조회",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:         discordgo.ApplicationCommandOptionString,
				Name:         "league",
				Description:  "리그 선택",
				Required:     true,
				Autocomplete: true,
			},
		},
	},
	{
		Name:        "schedule",
		Description: "레이스 일정 조회",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:         discordgo.ApplicationCommandOptionString,
				Name:         "league",
				Description:  "리그 선택",
				Required:     true,
				Autocomplete: true,
			},
		},
	},
	{
		Name:        "results",
		Description: "레이스 결과 조회",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:         discordgo.ApplicationCommandOptionString,
				Name:         "match",
				Description:  "매치 선택",
				Required:     true,
				Autocomplete: true,
			},
		},
	},
	{
		Name:        "leagues",
		Description: "활성 리그 목록 조회",
	},
	{
		Name:        "league-info",
		Description: "리그 상세 정보 조회",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:         discordgo.ApplicationCommandOptionString,
				Name:         "league",
				Description:  "리그 선택",
				Required:     true,
				Autocomplete: true,
			},
		},
	},
}
