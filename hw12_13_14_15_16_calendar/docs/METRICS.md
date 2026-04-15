## Документация по метрикам календаря

### 1. Бизнес-метрики

#### `calendar_events_created_total` (Counter)
- **Что измеряет**: Общее количество созданных событий за всё время работы сервиса
- **Почему важно**: Позволяет отслеживать общую активность пользователей и рост системы. В сочетании с `events_updated_total` и `events_deleted_total` дает полную картину жизненного цикла событий
- **Как использовать**: 
  - Рассчитывать rate (созданий в секунду/минуту) для определения пиковых нагрузок
  - Сравнивать с историческими данными для выявления трендов
  - Выявлять аномальные скачки (например, DDoS-атаки или баги, создающие дубликаты)

#### `calendar_events_updated_total` (Counter)
- **Что измеряет**: Общее количество обновлений событий
- **Почему важно**: Высокое количество обновлений может указывать на:
  - Частые изменения расписания (нормальная активность)
  - Проблемы на клиентской стороне (например, баги, вызывающие множественные обновления)
  - Проблемы с синхронизацией данных

#### `calendar_events_deleted_total` (Counter)
- **Что измеряет**: Общее количество удаленных событий
- **Почему важно**: Помогает отслеживать:
  - Естественный оборот данных (удаление старых событий)
  - Аномальное удаление (возможные проблемы с безопасностью или баги)
  - Соотношение created/deleted для оценки роста базы данных

#### `calendar_events_today` (Gauge)
- **Что измеряет**: Текущее количество событий, созданных сегодня
- **Почему важно**: 
  - Позволяет оценить дневную активность пользователей в реальном времени
  - Помогает планировать capacity (емкость) системы
  - Быстрое обнаружение аномалий (падение или резкий рост активности)

**Анализ производительности на основе бизнес-метрик:**
```promql
# Выявление пиковых нагрузок
rate(calendar_events_created_total[5m])

# Отношение создания к удалению
rate(calendar_events_created_total[1h]) / rate(calendar_events_deleted_total[1h])

# Прогнозирование роста базы данных
predict_linear(calendar_events_created_total[1d], 3600)
```

---

### 2. Метрики уведомлений

#### `calendar_notifications_sent_total` (Gauge)
- **Что измеряет**: Общее количество отправленных уведомлений
- **Почему важно**: 
  - Ключевой показатель работы системы уведомлений
  - Позволяет оценить нагрузку на планировщик и Kafka
  - При падении этого счетчика при наличии событий с `NotifyBefore` - сигнал о проблемах в pipeline уведомлений

#### `calendar_notifications_today` (Gauge)
- **Что измеряет**: Количество уведомлений, отправленных сегодня
- **Почему важно**:
  - Помогает предсказывать пиковые нагрузки в разное время суток
  - Позволяет настроить autoscaling для обработчиков уведомлений
  - Быстрое обнаружение проблем (например, если ожидается 1000 уведомлений, а их нет)

**Анализ узких мест в уведомлениях:**
```promql
# Обнаружение проблем с доставкой
calendar_notifications_today / events_with_notify_today < 0.9  # менее 90% доставки

# Анализ нагрузки по часам
increase(calendar_notifications_sent_total[1h])  # уведомления в час
```

---

### 3. HTTP-метрики

#### `http_requests_total` (Counter с labels)
- **Что измеряет**: Количество HTTP запросов с разбивкой по:
  - `method`: GET, POST, PUT, DELETE
  - `path`: эндпоинты API
  - `status`: коды ответов (2xx, 3xx, 4xx, 5xx)
- **Почему важно**: 
  - Полная картина нагрузки на API
  - Позволяет определить самые популярные эндпоинты для оптимизации
  - Выявление проблемных эндпоинтов по статусам ошибок

#### `http_request_duration_seconds` (Histogram)
- **Что измеряет**: Время выполнения запросов (гистограмма)
- **Почему важно**: 
  - Основной показатель производительности API
  - Позволяет вычислить процентили (p50, p95, p99)
  - Выявление медленных эндпоинтов

#### `http_errors_total` (Counter с labels)
- **Что измеряет**: Количество ошибок HTTP (4xx и 5xx)
- **Почему важно**:
  - Быстрое обнаружение проблем API
  - Выявление конкретных эндпоинтов с высоким процентом ошибок
  - Мониторинг SLA (допустимый уровень ошибок)

**Анализ производительности API:**
```promql
# P95 время ответа
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Процент ошибок на эндпоинт
sum(rate(http_errors_total[5m])) by (path) / sum(rate(http_requests_total[5m])) by (path)

# Медленные эндпоинты (p99 > 1s)
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m])) > 1

# Самые нагруженные эндпоинты
topk(5, sum(rate(http_requests_total[5m])) by (path))
```

---

## Как использовать метрики для выявления узких мест

### 1. **Проблемы с базой данных**
Если время ответа растет для GET эндпоинтов:
```promql
http_request_duration_seconds{path="/events", method="GET"}
```
**Причина**: Возможно, требуется оптимизация запросов, добавление индексов или кэширование

### 2. **Высокая нагрузка на API**
Если `http_requests_total` резко возрастает:
```promql
rate(http_requests_total[1m])
```
**Действия**: 
- Проверить логи на предмет DDoS
- Рассмотреть горизонтальное масштабирование
- Добавить rate limiting

### 3. **Проблемы с уведомлениями**
Если `calendar_notifications_sent_total` не растет, а события с `NotifyBefore` создаются:
```promql
calendar_notifications_today / (calendar_events_created_total with notify_before=true)
```
**Проблемы**:
- Планировщик не запускается
- Проблемы с подключением к Kafka
- Ошибки в сторере при сохранении уведомлений

### 4. **Деградация производительности**
Если p95 запросов начинает расти:
```promql
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, path))
```
**Поиск причины**:
- Проверить `http_errors_total` на наличие таймаутов
- Посмотреть системные метрики (CPU, память)
- Проверить соединение с БД

### 5. **Аномалии в бизнес-логике**
Если `events_created_total` растет, а `events_deleted_total` падает:
```promql
rate(events_created_total[1h]) - rate(events_deleted_total[1h])
```
**Возможные причины**:
- Баг в логике удаления событий
- Пользователи не могут удалить события
- Проблемы с UI

---

## Рекомендуемые алерты (для Prometheus)

```yaml
# Высокий процент ошибок
- alert: HighErrorRate
  expr: sum(rate(http_errors_total[5m])) / sum(rate(http_requests_total[5m])) > 0.05
  annotations:
    summary: "More than 5% of requests are failing"

# Медленные ответы
- alert: SlowResponses
  expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
  annotations:
    summary: "95th percentile response time > 1s"

# Нет уведомлений
- alert: NoNotifications
  expr: calendar_notifications_today == 0 and day_of_week != 7
  annotations:
    summary: "No notifications sent today despite having events"

# Аномальное количество созданий
- alert: HighEventCreation
  expr: rate(calendar_events_created_total[5m]) > 100
  annotations:
    summary: "High rate of event creation detected"
```