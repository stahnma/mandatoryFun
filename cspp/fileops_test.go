package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetupDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("creates new directory", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "newdir")
		setupDirectory(dir)
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("directory was not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("expected a directory")
		}
	})

	t.Run("creates nested directories", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "a", "b", "c")
		setupDirectory(dir)
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("nested directory was not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("expected a directory")
		}
	})

	t.Run("no error on existing directory", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "existing")
		os.MkdirAll(dir, 0o755)
		// Should not panic or error
		setupDirectory(dir)
	})
}

func TestMoveToDir(t *testing.T) {
	tmpDir := t.TempDir()
	destDir := filepath.Join(tmpDir, "dest")
	os.MkdirAll(destDir, 0o755)

	t.Run("moves file to destination", func(t *testing.T) {
		srcFile := filepath.Join(tmpDir, "testfile.txt")
		os.WriteFile(srcFile, []byte("hello"), 0o644)

		moveToDir(srcFile, destDir)

		// Source should no longer exist
		if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
			t.Error("source file should have been moved")
		}

		// Dest should exist
		destFile := filepath.Join(destDir, "testfile.txt")
		data, err := os.ReadFile(destFile)
		if err != nil {
			t.Fatalf("destination file not found: %v", err)
		}
		if string(data) != "hello" {
			t.Errorf("file content = %q, want %q", string(data), "hello")
		}
	})

	t.Run("handles nonexistent source gracefully", func(t *testing.T) {
		// Should log error but not panic
		moveToDir("/nonexistent/file.txt", destDir)
	})
}

func TestPrettyPrintJSON(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		err := prettyPrintJSON(data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("nil data", func(t *testing.T) {
		err := prettyPrintJSON(nil)
		if err != nil {
			t.Errorf("unexpected error for nil: %v", err)
		}
	})

	t.Run("complex nested data", func(t *testing.T) {
		data := map[string]interface{}{
			"name":   "test",
			"values": []int{1, 2, 3},
		}
		err := prettyPrintJSON(data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
