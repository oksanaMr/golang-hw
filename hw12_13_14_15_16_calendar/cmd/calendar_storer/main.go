package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/interfaces"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/pkg/kafka"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/storage/sql"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/storer"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/cmd/calendar_storer/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	config := NewConfig()
	err := config.readConfig(configFile)
	if err != nil {
		panic(err)
	}

	// Создаем консьюмера
	consumer, err := kafka.NewConsumer(config.Kafka.Brokers, config.Kafka.GroupID,
		config.Kafka.Topics.Notifications, config.Kafka.Timeout,
	)
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	// Подключаемся к БД
	logg, err := logger.New(config.Logger.Level, config.Logger.Filename)
	if err != nil {
		panic(err)
	}
	defer logg.Close()

	var storage storage.NotificationStorage
	if config.Storage.Mode == "in-memory" {
		storage = memorystorage.New()
	} else {
		var err error
		storage, err = sqlstorage.New(config.Storage.Dsn)
		if err != nil {
			logg.Error(err.Error())
			panic(err)
		}
	}
	defer storage.Close()

	service := storer.NewService(storage, logg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logg.Info("Хранитель остановлен")
		cancel()
	}()

	logg.Info("Хранитель запущен, ожидание сообщений...")

	// Подписываемся на топик
	err = consumer.Subscribe(ctx,
		func(ctx context.Context, msg interfaces.Message) {
			service.HandleMessage(ctx, msg)
		})

	if err != nil && err != context.Canceled {
		panic(err)
	}
}
