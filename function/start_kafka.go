package function

import (
	"context"
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/segmentio/kafka-go"
	"log"
	"strings"
	"telegramMessenger/client"
	"telegramMessenger/config"
	"time"
)

// Kafka consumer loop.
// Message key must be username (bytes) and value — string text to send.
func StartKafkaConsumer(ctx context.Context, cfg config.Config, bot *tgbotapi.BotAPI, userClient *client.UserServiceClient) {
	brokers := strings.Split(cfg.KafkaBrokers, ",")
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  cfg.KafkaGroupID,
		Topic:    cfg.KafkaTopic,
		MinBytes: 1,    // 1B
		MaxBytes: 10e6, // 10MB
	})
	defer func() {
		if err := r.Close(); err != nil {
			log.Println("kafka reader close error:", err)
		}
	}()

	log.Println("kafka consumer started, listening topic:", cfg.KafkaTopic)

	for {
		// allow context cancellation
		m, err := r.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				log.Println("kafka reader exiting:", err)
				return
			}
			log.Println("kafka read error:", err)
			// small backoff
			time.Sleep(1 * time.Second)
			continue
		}

		username := string(m.Key)
		text := string(m.Value)
		if username == "" {
			log.Printf("kafka message without key: %s", text)
			continue
		}

		// get chat id from user-service
		chatID, err := userClient.GetChatIDByUsername(context.Background(), username)
		if err != nil {
			log.Printf("failed to get chat id for %s: %v", username, err)
			continue
		}
		if chatID == 0 {
			log.Printf("no chat id for user %s; skipping", username)
			continue
		}

		// send tg message
		msg := tgbotapi.NewMessage(chatID, text)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("telegram send error to %d (user %s): %v", chatID, username, err)
			// nothing more to do
		} else {
			log.Printf("sent notification to %s (chat %d): %s", username, chatID, text)
		}
	}
}
