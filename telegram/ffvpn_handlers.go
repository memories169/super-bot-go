package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"super-bot/core"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func ffvpnContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 15*time.Second)
}

// HandleFFVPNStatusCommand handles /ffvpn — current server + proxy state.
func HandleFFVPNStatusCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🔄 *Đang tải ffvpn...*")
	msg.ParseMode = "Markdown"
	sent, err := bot.Send(msg)
	if err != nil {
		log.Println("ffvpn: send:", err)
		return
	}

	ctx, cancel := ffvpnContext()
	defer cancel()
	ch := make(chan core.FFVPNInfo, 1)
	core.GetFFVPNInfo(ctx, ch)
	info := <-ch
	if info.Error != "" {
		edit := tgbotapi.NewEditMessageText(update.Message.Chat.ID, sent.MessageID, "❌ ffvpn: "+info.Error)
		bot.Send(edit)
		return
	}

	text := FormatFFVPNMessage(info)
	keyboard := CreateFFVPNKeyboard(info)
	edit := tgbotapi.NewEditMessageText(update.Message.Chat.ID, sent.MessageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	bot.Send(edit)
}

// HandleFFVPNButtonCallback handles "ffvpn|<host>" and "ffvpnrefresh".
func HandleFFVPNButtonCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	callback := update.CallbackQuery
	data := callback.Data

	if data == "ffvpnrefresh" {
		bot.Request(tgbotapi.NewCallback(callback.ID, "🔄 Đang cập nhật..."))
	} else if strings.HasPrefix(data, "ffvpn|") {
		host := strings.TrimPrefix(data, "ffvpn|")
		bot.Request(tgbotapi.NewCallback(callback.ID, "🔄 Đang chuyển sang "+host+"..."))
		if err := core.SwitchFFVPN(host); err != nil {
			bot.Request(tgbotapi.NewCallbackWithAlert(callback.ID, "❌ Lỗi: "+err.Error()))
		} else {
			bot.Request(tgbotapi.NewCallback(callback.ID, "✅ Đã chuyển sang "+host))
		}
	}

	ctx, cancel := ffvpnContext()
	defer cancel()
	ch := make(chan core.FFVPNInfo, 1)
	core.GetFFVPNInfo(ctx, ch)
	info := <-ch
	if info.Error != "" {
		return
	}
	text := FormatFFVPNMessage(info)
	keyboard := CreateFFVPNKeyboard(info)
	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	_, _ = bot.Send(edit) // ignore "message is not modified"
}

// FormatFFVPNMessage renders the ffvpn panel.
func FormatFFVPNMessage(info core.FFVPNInfo) string {
	var sb strings.Builder
	sb.WriteString("🦊 *FFVPN*\n")
	sb.WriteString(fmt.Sprintf("📍 Server: `%s`\n", ffvpnDisplay(info.Current)))
	sb.WriteString(fmt.Sprintf("🔗 Proxy: *%s* (account `%s`)\n",
		onOffT(info.Status.Running), valT(info.Status.Account)))
	if info.Status.EdgeSuspect {
		sb.WriteString("⚠️ _Edge này đang từ chối tunnel — chọn server khác._\n")
	}
	sb.WriteString(fmt.Sprintf("🌐 %d edges khả dụng\n", len(info.Servers)))
	sb.WriteString("Bấm một server để đổi — sing-box không cần đổi gì.")
	return sb.String()
}

// CreateFFVPNKeyboard builds the edge-selection keyboard (2 per row, max
// 24 edges, current highlighted 🟢).
func CreateFFVPNKeyboard(info core.FFVPNInfo) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton

	shown := 0
	for _, e := range info.Servers {
		if shown >= 24 {
			break
		}
		shown++
		icon := "🌐"
		if e.ServerHost == info.Current.ServerHost {
			icon = "🟢"
		}
		label := fmt.Sprintf("%s %s", icon, e.CountryCode)
		btn := tgbotapi.NewInlineKeyboardButtonData(label, "ffvpn|"+e.ServerHost)
		row = append(row, btn)
		if len(row) == 2 {
			rows = append(rows, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh", "ffvpnrefresh"),
	})
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func ffvpnDisplay(e core.FFVPNEdge) string {
	return fmt.Sprintf("%s/%s (%s:%d)", e.CountryName, e.CityName, e.ServerHost, e.ServerPort)
}

func onOffT(b bool) string {
	if b {
		return "running"
	}
	return "stopped"
}

func valT(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}
