package utils

import (
	"os"
	"path/filepath"
)

func WriteTempFile(dir string, name string, content string) string {
	path := filepath.Join(dir, name)
	_ = os.WriteFile(path, []byte(content), 0o644)
	return path
}

func ReadTempFile(path string) string {
	data, _ := os.ReadFile(path)
	return string(data)
}

func DeleteTempFile(path string) {
	_ = os.Remove(path)
}
