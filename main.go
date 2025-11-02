package main

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	//if cfg.KafkaBrokers == "" {
	//	log.Fatal("KAFKA_BROKERS is required")
	//}
	//if cfg.KafkaTopic == "" {
	//	log.Fatal("KAFKA_TOPIC is required")
	//}
	//if cfg.KafkaGroupID == "" {
	//	cfg.KafkaGroupID = "notifier-group"
	//}

	userClient := client.NewUserServiceClient(cfg.UserServiceURL)

	// create bot
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}
	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// start updates handler (bind username <-> chat_id)
	go function.StartTelegramListener(ctx, bot, userClient)

	// start kafka consumer
	//go function.StartKafkaConsumer(ctx, cfg, bot, userClient)

	// wait for signal
	<-sigs
	log.Println("shutting down notifier...")
	cancel()
	time.Sleep(500 * time.Millisecond)
}
