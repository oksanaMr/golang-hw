package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	EventsCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "calendar_events_created_total",
		Help: "Total number of events created",
	})

	EventsUpdated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "calendar_events_updated_total",
		Help: "Total number of events updated",
	})

	EventsDeleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "calendar_events_deleted_total",
		Help: "Total number of events deleted",
	})

	// Количество событий за сегодня
	EventsToday = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "calendar_events_today",
		Help: "Number of events today",
	})

	// Общее количество отправленных уведомлений
	NotificationSent = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "calendar_notifications_sent_total",
		Help: "Total number of notifications sent",
	})

	// Количество отправленных уведомлений за сегодня
	NotificationsToday = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "calendar_notifications_today",
		Help: "Number of notifications today",
	})

	// Счетчик запросов по методу, пути и статусу
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// Гистограмма времени выполнения запросов
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets, // стандартные бакеты: 0.005, 0.01, 0.025, 0.05, ...
		},
		[]string{"method", "path"},
	)

	// Счетчик ошибок (для быстрого мониторинга)
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total number of HTTP errors (4xx, 5xx)",
		},
		[]string{"method", "path", "status"},
	)
)
