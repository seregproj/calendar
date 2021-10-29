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
	schedulerapp "github.com/seregproj/calendar/internal/app/scheduler"
	internallogger "github.com/seregproj/calendar/internal/logger"
	"github.com/seregproj/calendar/internal/messagebroker/rbmq/notifications"
	memorystorage "github.com/seregproj/calendar/internal/storage/memory"
	sqlstorage "github.com/seregproj/calendar/internal/storage/sql"
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

	var storage schedulerapp.Storage
	switch config.Storage.Type {
	case "memory":
		storage = memorystorage.New()
	case "pgsql":
		ss := sqlstorage.New()
		err = ss.Connect(ctx, config.Storage.PGSQL.DSN)
		if err != nil {
			fmt.Println(fmt.Errorf("cant prepare sql conn: %w", err))

			return
		}

		defer ss.Close(ctx)

		storage = ss
	default:
		fmt.Println(fmt.Errorf("invalid storage type: %s", config.Storage.Type))

		return
	}

	producerRbmq := notifications.NewProducer(config.MessageBroker.QueueName)
	if err := producerRbmq.Connect(ctx, config.MessageBroker.DSN); err != nil {
		fmt.Println("cant connect to rbmq: ", err)

		return
	}

	scheduler := schedulerapp.New(logger, storage, producerRbmq)

	go func() {
		defer cancel()

		if err := scheduler.ProcessActualEvents(ctx, config.App.Notifications.Limit); err != nil {
			fmt.Println("cant process actual events: ", err)

			return
		}

		fmt.Println("successfully processed")
	}()

	<-ctx.Done()
	fmt.Println("Graceful shutdown...")
	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
}
