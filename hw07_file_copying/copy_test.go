package main

import (
	"bytes"
	"os"
	"testing"
)

func TestCopy(t *testing.T) {
	t.Run("file", func(t *testing.T) {
		srcFile := "testdata/input.txt"
		dstFile := "out.txt"

		expectedData, err := os.ReadFile(srcFile)
		if err != nil {
			t.Fatalf("Failed to read source file: %v", err)
		}

		defer os.Remove(dstFile)

		err = Copy(srcFile, dstFile, 0, 0)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			return
		}

		// Проверяем, что файл создан
		if _, err := os.Stat(dstFile); os.IsNotExist(err) {
			t.Errorf("Destination file was not created")
			return
		}

		// Читаем и проверяем содержимое
		actualData, err := os.ReadFile(dstFile)
		if err != nil {
			t.Errorf("Failed to read destination file: %v", err)
			return
		}

		if !bytes.Equal(actualData, expectedData) {
			t.Errorf("Unexpected content\nexpected length: %d\nactual length:   %d",
				len(expectedData), len(actualData))
		}
	})
}
