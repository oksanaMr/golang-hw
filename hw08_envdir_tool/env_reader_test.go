package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDir(t *testing.T) {
	// Создаем временную директорию для тестов
	tmpDir := t.TempDir()

	testCases := []struct {
		name     string
		content  string
		expected EnvValue
	}{
		{
			name:    "FOO",
			content: "123\n",
			expected: EnvValue{
				Value:      "123",
				NeedRemove: false,
			},
		},
		{
			name:    "BAR",
			content: "value with spaces  \t\n",
			expected: EnvValue{
				Value:      "value with spaces",
				NeedRemove: false,
			},
		},
		{
			name:    "EMPTY",
			content: "",
			expected: EnvValue{
				NeedRemove: true,
			},
		},
		{
			name:    "WITH_NULL",
			content: "hello\x00world\n",
			expected: EnvValue{
				Value:      "hello\nworld",
				NeedRemove: false,
			},
		},
	}

	// Создаем файлы
	for _, tc := range testCases {
		err := os.WriteFile(filepath.Join(tmpDir, tc.name), []byte(tc.content), 0o644)
		if err != nil {
			t.Fatal(err)
		}
	}

	err := os.Mkdir(filepath.Join(tmpDir, "SUBDIR"), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "BAD=NAME"), []byte("ignored"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Читаем директорию
	env, err := ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	// Проверяем результаты
	for _, tc := range testCases {
		ev, ok := env[tc.name]
		if !ok {
			t.Errorf("Expected key %s not found", tc.name)
			continue
		}

		if ev.NeedRemove != tc.expected.NeedRemove {
			t.Errorf("%s: NeedRemove = %v, want %v", tc.name, ev.NeedRemove, tc.expected.NeedRemove)
		}

		if ev.Value != tc.expected.Value {
			t.Errorf("%s: Value = %q, want %q", tc.name, ev.Value, tc.expected.Value)
		}
	}

	// Проверяем, что поддиректория и файл с "=" не попали в результат
	if _, ok := env["SUBDIR"]; ok {
		t.Error("SUBDIR should not be in result")
	}
	if _, ok := env["BAD=NAME"]; ok {
		t.Error("BAD=NAME should not be in result")
	}
}
