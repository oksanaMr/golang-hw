CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    event_time TIMESTAMPTZ NOT NULL,
    duration BIGINT NOT NULL,
    description TEXT,
    user_id UUID NOT NULL,
    notify_before BIGINT
);

-- Индекс для поиска по времени
CREATE INDEX IF NOT EXISTS idx_events_event_time ON events(event_time);

-- Индекс для поиска по пользователю
CREATE INDEX IF NOT EXISTS idx_events_user_id ON events(user_id);

-- Составной индекс для поиска событий пользователя за период
CREATE INDEX IF NOT EXISTS idx_events_user_time ON events(user_id, event_time);