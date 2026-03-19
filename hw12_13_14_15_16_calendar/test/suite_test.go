package test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

var (
	testDB       *sql.DB
	kafkaBrokers []string
	apiBaseURL   string
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		log.Fatalf("Failed to setup test environment: %v", err)
	}

	code := m.Run()

	cleanup()

	os.Exit(code)
}

func setup() error {
	// Получаем параметры из окружения (задаются в docker-compose)
	pgHost := getEnv("POSTGRES_HOST", "localhost")
	pgPort := getEnv("POSTGRES_PORT", "5432")
	pgUser := getEnv("POSTGRES_USER", "test")
	pgPass := getEnv("POSTGRES_PASSWORD", "test")
	pgDB := getEnv("POSTGRES_DB", "testdb")

	kafkaBrokers = []string{getEnv("KAFKA_BROKERS", "localhost:9092")}
	apiHost := getEnv("API_HOST", "localhost")
	apiPort := getEnv("API_PORT", "8888")
	apiBaseURL = fmt.Sprintf("http://%s:%s", apiHost, apiPort)

	// Подключение к PostgreSQL
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=public",
		pgUser, pgPass, pgHost, pgPort, pgDB)

	var err error
	testDB, err = sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}

	if err = testDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping postgres: %w", err)
	}

	// Очищаем таблицы перед тестами
	if err := truncateDB(); err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}

	return nil
}

func cleanup() {
	if testDB != nil {
		_ = truncateDB()
		testDB.Close()
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func truncateDB() error {
	_, err := testDB.Exec(`
        TRUNCATE TABLE events CASCADE;
        TRUNCATE TABLE notifications CASCADE;
    `)
	return err
}

type TestHTTPClient struct {
	client  *http.Client
	baseURL string
}

func NewTestHTTPClient() *TestHTTPClient {
	return &TestHTTPClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: apiBaseURL,
	}
}

func (c *TestHTTPClient) Post(endpoint string, body interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return c.client.Do(req)
}

func (c *TestHTTPClient) Get(endpoint string) (*http.Response, error) {
	fullURL := c.baseURL + endpoint

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", fullURL, err)
	}

	return resp, nil
}

func (c *TestHTTPClient) ParseResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if len(body) == 0 {
		return nil
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w\nBody: %s", err, string(body))
	}

	return nil
}
