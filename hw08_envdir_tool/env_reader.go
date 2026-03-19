package main

import (
	"os"
	"path/filepath"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	result := make(map[string]EnvValue)

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, fileInfo := range files {
		if fileInfo.IsDir() || strings.Contains(fileInfo.Name(), "=") {
			continue
		}

		// Читаем файл
		content, err := os.ReadFile(filepath.Join(dir, fileInfo.Name()))
		if err != nil {
			return nil, err
		}

		if len(content) == 0 {
			result[fileInfo.Name()] = EnvValue{NeedRemove: true}
			continue
		}

		// Берем первую строку
		lines := strings.SplitN(string(content), "\n", 2)
		firstLine := lines[0]

		// Удаляем пробелы и табуляцию в конце и заменяем терминальные нули
		firstLine = strings.TrimRight(firstLine, " \t")
		firstLine = strings.ReplaceAll(firstLine, "\x00", "\n")

		result[fileInfo.Name()] = EnvValue{
			Value:      firstLine,
			NeedRemove: false,
		}
	}

	return result, nil
}
