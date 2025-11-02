package function

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
	"telegramMessenger/client"
	"time"
)

// Telegram updates processing: /start -> prompt, otherwise treat text as username to bind
func StartTelegramListener(ctx context.Context, bot *tgbotapi.BotAPI, userClient *client.UserServiceClient) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			log.Println("telegram listener stopped")
			return
		case update, ok := <-updates:
			if !ok {
				log.Println("updates channel closed")
				return
			}
			if update.Message == nil {
				continue
			}

			chatID := update.Message.Chat.ID
			tgnickname := update.Message.From.UserName

			// commands
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "start":
					msg := tgbotapi.NewMessage(chatID, "Привет! Чтобы привязать аккаунт, пришли свой логин (username) из приложения.")
					if _, err := bot.Send(msg); err != nil {
						log.Println("bot send error:", err)
					}
				case "help":
					msg := tgbotapi.NewMessage(chatID, "Пришли логин из приложения — и я привяжу уведомления к этому чату.")
					if _, err := bot.Send(msg); err != nil {
						log.Println("bot send error:", err)
					}
				default:
					msg := tgbotapi.NewMessage(chatID, "Неизвестная команда")
					bot.Send(msg)
				}
				continue
			}

			// treat plain text as username for binding
			username := strings.TrimSpace(update.Message.Text)
			if username == "" {
				bot.Send(tgbotapi.NewMessage(chatID, "Пустой логин — пришлите ваш логин из приложения"))
				continue
			}

			// call user-service to link
			go func(username string, chatID int64, tgnickname string) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				err := userClient.LinkTelegramAccount(ctx, username, chatID, tgnickname)
				if err != nil {
					log.Printf("link telegram failed for %s: %v", username, err)
					bot.Send(tgbotapi.NewMessage(chatID, "Ошибка: не удалось привязать аккаунт. Проверьте логин и попробуйте снова."))
					return
				}
				bot.Send(tgbotapi.NewMessage(chatID, "Аккаунт успешно привязан"))
			}(username, chatID, tgnickname)
		}
	}
}
