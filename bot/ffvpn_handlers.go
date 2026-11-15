package bot

import (
	"context"
	"fmt"
	"strings"
	"time"
	"super-bot/core"

	"github.com/bwmarrin/discordgo"
)

func ffvpnContextD() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 15*time.Second)
}

// HandleFFVPNStatusCommand handles the /ffvpn slash command.
func HandleFFVPNStatusCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	ctx, cancel := ffvpnContextD()
	defer cancel()
	ch := make(chan core.FFVPNInfo, 1)
	core.GetFFVPNInfo(ctx, ch)
	info := <-ch
	if info.Error != "" {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: stringPtr("❌ ffvpn: " + info.Error),
		})
		return
	}

	embed := CreateFFVPNEmbed(info)
	components := CreateFFVPNButtons(info)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
}

// HandleFFVPNButtonClick handles "ffvpn_<host>" and "ffvpn_refresh".
func HandleFFVPNButtonClick(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})

	if customID != "ffvpn_refresh" && strings.HasPrefix(customID, "ffvpn_") {
		host := strings.TrimPrefix(customID, "ffvpn_")
		if err := core.SwitchFFVPN(host); err != nil {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "❌ " + err.Error(),
				Flags:   discordgo.MessageFlagsEphemeral,
			})
			return
		}
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("✅ Đã chuyển sang %s", host),
			Flags:   discordgo.MessageFlagsEphemeral,
		})
	}

	ctx, cancel := ffvpnContextD()
	defer cancel()
	ch := make(chan core.FFVPNInfo, 1)
	core.GetFFVPNInfo(ctx, ch)
	info := <-ch
	if info.Error != "" {
		return
	}
	embed := CreateFFVPNEmbed(info)
	components := CreateFFVPNButtons(info)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
}
