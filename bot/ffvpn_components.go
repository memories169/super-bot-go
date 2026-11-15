package bot

import (
	"fmt"
	"super-bot/core"

	"github.com/bwmarrin/discordgo"
)

// CreateFFVPNButtons builds the edge-selection components (5 per row, max
// 25 buttons — Discord limits), current edge highlighted green.
func CreateFFVPNButtons(info core.FFVPNInfo) []discordgo.MessageComponent {
	var components []discordgo.MessageComponent
	var row []discordgo.MessageComponent

	shown := 0
	for _, e := range info.Servers {
		if shown >= 25 {
			break
		}
		shown++
		style := discordgo.SecondaryButton
		emoji := "🌐"
		if e.ServerHost == info.Current.ServerHost {
			style = discordgo.SuccessButton
			emoji = "🟢"
		}
		btn := discordgo.Button{
			Label:    fmt.Sprintf("%s (%s)", e.CountryCode, e.CityName),
			Style:    style,
			CustomID: "ffvpn_" + e.ServerHost,
			Emoji:    &discordgo.ComponentEmoji{Name: emoji},
		}
		row = append(row, btn)
		if len(row) == 5 {
			components = append(components, discordgo.ActionsRow{Components: row})
			row = []discordgo.MessageComponent{}
		}
	}
	if len(row) > 0 {
		components = append(components, discordgo.ActionsRow{Components: row})
	}
	components = append(components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "🔄 Refresh",
				Style:    discordgo.PrimaryButton,
				CustomID: "ffvpn_refresh",
				Emoji:    &discordgo.ComponentEmoji{Name: "🔄"},
			},
		},
	})
	return components
}

// CreateFFVPNEmbed renders the ffvpn panel as an embed.
func CreateFFVPNEmbed(info core.FFVPNInfo) *discordgo.MessageEmbed {
	state := "stopped"
	if info.Status.Running {
		state = "running"
	}
	color := 0x888888
	if info.Status.Running {
		color = 0x2ecc71
	}
	description := fmt.Sprintf(
		"**Server:** %s\n**Proxy:** %s\n**Account:** %s\n**Edges:** %d",
		ffvpnDisplayD(info.Current), state, valD(info.Status.Account), len(info.Servers))
	if info.Status.EdgeSuspect {
		description += "\n⚠️ This edge is rejecting tunnels server-side — pick another server."
	}
	return &discordgo.MessageEmbed{
		Title:       "🦊 FFVPN",
		Description: description,
		Color:       color,
	}
}

func ffvpnDisplayD(e core.FFVPNEdge) string {
	return fmt.Sprintf("%s/%s (%s:%d)", e.CountryName, e.CityName, e.ServerHost, e.ServerPort)
}

func valD(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}
