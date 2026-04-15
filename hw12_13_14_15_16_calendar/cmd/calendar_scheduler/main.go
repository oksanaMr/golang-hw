package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/pkg/kafka"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/scheduler"
	"github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/oksanaMr/golang-hw/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/cmd/config/calendar_scheduler.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	config := NewConfig()
	err := config.readConfig(configFile)
	if err != nil {
		panic(err)
	}

	// Ждем Kafka
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := kafka.WaitForKafka(ctx, config.Kafka.Brokers, 10); err != nil {
		panic(err)
	}

	// Создаем продюсера
	producer, err := kafka.NewProducer(config.Kafka.Brokers, config.Kafka.Topics.Notifications, config.Kafka.Timeout)
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	// Подключаемся к БД
	logg, err := logger.New(config.Logger.Level, config.Logger.Filename)
	if err != nil {
		panic(err)
	}
	defer logg.Close()

	var storage storage.EventStorage
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

	service := scheduler.NewService(storage, producer, logg)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	logg.Info("Планировщик запущен")

	for {
		select {
		case <-ticker.C:
			// Обрабатываем уведомления
			service.ProcessNotifications(context.Background())

			// Раз в день чистим старые события
			if time.Now().Hour() == 0 && time.Now().Minute() == 0 {
				service.CleanupOldEvents(context.Background())
			}

		case <-sigChan:
			logg.Info("Планировщик остановлен")
			return
		}
	}
}
