package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	senderapp "github.com/seregproj/calendar/internal/app/sender"
	internallogger "github.com/seregproj/calendar/internal/logger"
	"github.com/seregproj/calendar/internal/messagebroker/rbmq/notifications"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "/etc/calendar/config.yml", "Path to configuration file")
	flag.Parse()

	config := NewConfig()
	err := cleanenv.ReadConfig(configFile, &config)
	if err != nil {
		fmt.Println(fmt.Errorf("cant read config: %w", err))
		os.Exit(1)
	}

	lvl, err := internallogger.NewLevel(config.Logger.Level)
	if err != nil {
		fmt.Println(fmt.Errorf("cant create log level: %w", err))
		os.Exit(1)
	}

	logger, err := internallogger.New(config.Logger.File, lvl)
	if err != nil {
		fmt.Println(fmt.Errorf("cant create log: %w", err))
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	consumerRbmq := notifications.NewConsumer(config.MessageBroker.QueueName)
	if err := consumerRbmq.Connect(ctx, config.MessageBroker.DSN); err != nil {
		fmt.Println("cant connect to rbmq: ", err)

		return
	}

	sender := senderapp.New(logger, consumerRbmq)

	go func() {
		defer cancel()

		if err := sender.SendNotifications(ctx); err != nil {
			fmt.Println("cant send notifications: ", err)

			return
		}

		fmt.Println("successfully processed")
	}()

	<-ctx.Done()
	fmt.Println("Graceful shutdown...")
	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
}
