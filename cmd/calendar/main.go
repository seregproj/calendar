package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/seregproj/calendar/api/proto"
	calendarapp "github.com/seregproj/calendar/internal/app/calendar"
	internallogger "github.com/seregproj/calendar/internal/logger"
	internalgrpc "github.com/seregproj/calendar/internal/server/grpc"
	internalstorage "github.com/seregproj/calendar/internal/storage"
	memorystorage "github.com/seregproj/calendar/internal/storage/memory"
	sqlstorage "github.com/seregproj/calendar/internal/storage/sql"
	"google.golang.org/grpc"
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

	var storage calendarapp.Storage
	switch config.Storage.Type {
	case "memory":
		storage = memorystorage.New()
	case "pgsql":
		ss := sqlstorage.New()
		err = ss.Connect(ctx, config.Storage.PGSQL.DSN)
		if err != nil {
			fmt.Println(fmt.Errorf("cant prepare sql conn: %w", err))
			os.Exit(1) //nolint:gocritic
		}

		defer ss.Close(ctx)

		storage = ss
	default:
		fmt.Println(fmt.Errorf("invalid storage type: %s", config.Storage.Type))
		os.Exit(1)
	}

	calendar := calendarapp.New(logger, storage, internalstorage.NewUUIDGen())
	eventsService := internalgrpc.NewEventServer(calendar)

	// HTTP
	mux := runtime.NewServeMux()
	if err = proto.RegisterEventServiceHandlerServer(ctx, mux, eventsService); err != nil {
		logger.Error(fmt.Sprintf("cant register service handler: %v", err))

		return
	}

	httpServer := http.Server{
		Addr:    net.JoinHostPort(config.Server.HTTP.Host, config.Server.HTTP.Port),
		Handler: mux,
	}

	go func() {
		defer cancel()

		if err = httpServer.ListenAndServe(); err != nil {
			logger.Error(fmt.Sprintf("cant start HTTP server: %v", err))
		}
	}()

	// GRPC
	grpcServer := grpc.NewServer()

	go func() {
		defer cancel()

		proto.RegisterEventServiceServer(grpcServer, eventsService)
		l, err := net.Listen("tcp", net.JoinHostPort(config.Server.GRPC.Host, config.Server.GRPC.Port))
		if err != nil {
			logger.Error(fmt.Sprintf("cant get grpc listener: %v", err))

			return
		}

		err = grpcServer.Serve(l)
		if err != nil {
			logger.Error(fmt.Sprintf("cant start GRPC server: %v", err))
		}
	}()

	<-ctx.Done()
	fmt.Println("Graceful shutdown...")
	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err = httpServer.Shutdown(ctx); err != nil {
		logger.Error(fmt.Sprintf("cant shutdown http server: %v", err))
	}
	grpcServer.Stop()
}
