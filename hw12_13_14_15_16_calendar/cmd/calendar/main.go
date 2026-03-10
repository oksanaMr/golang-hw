package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"
	"time"

	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/app"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/logger"
	internalhttp "github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/server/http"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.toml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config := NewConfig()
	config.readConfig(configFile)

	logg, err := logger.New(config.Logger.Level)
	if err != nil {
		panic(err)
	}
	defer logg.Close()

	var storage storage.EventStorage
	if config.Storage.Mode == "in-memory" {
		storage = memorystorage.New()
	} else {
		storage, err := sqlstorage.New(config.Storage.Dsn)
		if err != nil {
			logg.Error(err.Error())
			panic(err)
		}
		defer storage.Close()
	}

	calendar := app.New(storage)

	server := internalhttp.NewServer(logg, calendar)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx, config.Server.Host, config.Server.Port); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		panic(err)
	}
}
