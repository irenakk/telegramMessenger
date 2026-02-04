package main

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/segmentio/kafka-go"
	"log"
	"os"
	"os/signal"
	"syscall"
	"telegramMessenger/client"
	"telegramMessenger/config"
	"telegramMessenger/function"
	"time"
)

func main() {
	cfg := config.LoadConfig()

	if cfg.TelegramToken == "" {
		log.Fatal("TELEGRAM_TOKEN is required")
	}
	if cfg.UserServiceURL == "" {
		log.Fatal("USER_SERVICE_URL is required")
	}

	userClient := client.NewUserServiceClient(cfg.UserServiceURL)

	// create bot
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}
	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Initialize Kafka consumer
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "test-topic",
		GroupID:  "tg-group",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	// graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// start updates handler (bind username <-> chat_id)
	go function.StartTelegramListener(ctx, bot, userClient, reader)

	// wait for signal
	<-sigs
	log.Println("shutting down notifier...")
	cancel()
	time.Sleep(500 * time.Millisecond)
}
