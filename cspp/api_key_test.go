package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func setupTestCredentialsDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	credDir := filepath.Join(tmpDir, "credentials")
	os.MkdirAll(credDir, 0o755)
	viper.Set("credentials_dir", credDir)
	return credDir
}

func createTestApiEntry(t *testing.T, credDir string, ae ApiEntry) {
	t.Helper()
	data, err := json.Marshal(ae)
	if err != nil {
		t.Fatalf("failed to marshal ApiEntry: %v", err)
	}
	filename := filepath.Join(credDir, ae.ApiKey+".json")
	err = os.WriteFile(filename, data, 0o644)
	if err != nil {
		t.Fatalf("failed to write test api entry: %v", err)
	}
}

func TestApiEntry_IsRevoked(t *testing.T) {
	tests := []struct {
		name    string
		revoked bool
		want    bool
	}{
		{"not revoked", false, false},
		{"revoked", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ae := ApiEntry{Revoked: tt.revoked}
			if got := ae.isRevoked(); got != tt.want {
				t.Errorf("isRevoked() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApiEntry_Save(t *testing.T) {
	credDir := setupTestCredentialsDir(t)

	ae := ApiEntry{
		ApiKey:    "test-key-123",
		IssueDate: "2024-01-01T00:00:00Z",
		SlackId:   "U12345",
		Revoked:   false,
	}

	ae.save()

	filename := filepath.Join(credDir, "test-key-123.json")
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	var loaded ApiEntry
	err = json.Unmarshal(data, &loaded)
	if err != nil {
		t.Fatalf("failed to unmarshal saved data: %v", err)
	}

	if loaded.ApiKey != ae.ApiKey {
		t.Errorf("ApiKey = %q, want %q", loaded.ApiKey, ae.ApiKey)
	}
	if loaded.SlackId != ae.SlackId {
		t.Errorf("SlackId = %q, want %q", loaded.SlackId, ae.SlackId)
	}
	if loaded.Revoked != ae.Revoked {
		t.Errorf("Revoked = %v, want %v", loaded.Revoked, ae.Revoked)
	}
}

func TestLoadApiEntryFromFile(t *testing.T) {
	credDir := setupTestCredentialsDir(t)

	ae := ApiEntry{
		ApiKey:    "load-test-key",
		IssueDate: "2024-01-01T00:00:00Z",
		SlackId:   "U99999",
		Revoked:   false,
	}
	createTestApiEntry(t, credDir, ae)

	loaded, err := loadApiEntryFromFile(filepath.Join(credDir, "load-test-key.json"))
	if err != nil {
		t.Fatalf("loadApiEntryFromFile failed: %v", err)
	}

	if loaded.ApiKey != ae.ApiKey {
		t.Errorf("ApiKey = %q, want %q", loaded.ApiKey, ae.ApiKey)
	}
	if loaded.SlackId != ae.SlackId {
		t.Errorf("SlackId = %q, want %q", loaded.SlackId, ae.SlackId)
	}
}

func TestLoadApiEntryFromFile_NonexistentFile(t *testing.T) {
	_, err := loadApiEntryFromFile("/nonexistent/path.json")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestLoadApiEntryFromFile_InvalidJson(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "bad.json")
	os.WriteFile(f, []byte("not json"), 0o644)

	_, err := loadApiEntryFromFile(f)
	if err == nil {
		t.Error("expected error for invalid json, got nil")
	}
}

func TestSearchAPIKeyInFile(t *testing.T) {
	credDir := setupTestCredentialsDir(t)

	ae := ApiEntry{
		ApiKey:  "search-key-abc",
		SlackId: "U11111",
	}
	createTestApiEntry(t, credDir, ae)

	filePath := filepath.Join(credDir, "search-key-abc.json")

	t.Run("matching key", func(t *testing.T) {
		match, err := searchAPIKeyInFile(filePath, "search-key-abc")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !match {
			t.Error("expected match for correct key")
		}
	})

	t.Run("non-matching key", func(t *testing.T) {
		match, err := searchAPIKeyInFile(filePath, "wrong-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if match {
			t.Error("expected no match for wrong key")
		}
	})
}

func TestSearchAPIKeyInDirectory(t *testing.T) {
	credDir := setupTestCredentialsDir(t)

	entries := []ApiEntry{
		{ApiKey: "dir-key-1", SlackId: "U1"},
		{ApiKey: "dir-key-2", SlackId: "U2"},
		{ApiKey: "dir-key-3", SlackId: "U3"},
	}
	for _, ae := range entries {
		createTestApiEntry(t, credDir, ae)
	}

	t.Run("finds existing key", func(t *testing.T) {
		matches, err := SearchAPIKeyInDirectory("dir-key-2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(matches) != 1 {
			t.Fatalf("expected 1 match, got %d", len(matches))
		}
	})

	t.Run("no matches for missing key", func(t *testing.T) {
		matches, err := SearchAPIKeyInDirectory("nonexistent-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(matches) != 0 {
			t.Errorf("expected 0 matches, got %d", len(matches))
		}
	})
}

func TestValidateApiKey(t *testing.T) {
	credDir := setupTestCredentialsDir(t)

	ae := ApiEntry{
		ApiKey:  "valid-key-123",
		SlackId: "U5555",
		Revoked: false,
	}
	createTestApiEntry(t, credDir, ae)

	t.Run("valid non-revoked key", func(t *testing.T) {
		valid, err := validateApiKey("valid-key-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !valid {
			t.Error("expected key to be valid")
		}
	})

	t.Run("nonexistent key", func(t *testing.T) {
		valid, _ := validateApiKey("no-such-key")
		if valid {
			t.Error("expected nonexistent key to be invalid")
		}
	})
}

func TestValidateApiKey_Revoked(t *testing.T) {
	credDir := setupTestCredentialsDir(t)

	ae := ApiEntry{
		ApiKey:  "revoked-key-456",
		SlackId: "U6666",
		Revoked: true,
	}
	createTestApiEntry(t, credDir, ae)

	valid, err := validateApiKey("revoked-key-456")
	if valid {
		t.Error("expected revoked key to be invalid")
	}
	if err == nil {
		t.Error("expected error for revoked key")
	}
}

func TestRevokeApiKey(t *testing.T) {
	credDir := setupTestCredentialsDir(t)

	ae := ApiEntry{
		ApiKey:  "to-revoke-key",
		SlackId: "U7777",
		Revoked: false,
	}
	createTestApiEntry(t, credDir, ae)

	result := revokeApiKey("to-revoke-key")
	if !result {
		t.Fatal("expected revokeApiKey to return true")
	}

	// Verify the key is now revoked on disk
	loaded, err := loadApiEntryFromFile(filepath.Join(credDir, "to-revoke-key.json"))
	if err != nil {
		t.Fatalf("failed to load revoked entry: %v", err)
	}
	if !loaded.Revoked {
		t.Error("expected key to be revoked after revokeApiKey")
	}
}

func TestRevokeApiKey_NonexistentKey(t *testing.T) {
	setupTestCredentialsDir(t)

	result := revokeApiKey("nonexistent-key")
	if result {
		t.Error("expected revokeApiKey to return false for nonexistent key")
	}
}

func TestIsRevoked(t *testing.T) {
	credDir := setupTestCredentialsDir(t)

	t.Run("not revoked", func(t *testing.T) {
		ae := ApiEntry{ApiKey: "not-revoked", Revoked: false}
		createTestApiEntry(t, credDir, ae)

		revoked, err := isRevoked(filepath.Join(credDir, "not-revoked.json"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if revoked {
			t.Error("expected key to not be revoked")
		}
	})

	t.Run("revoked", func(t *testing.T) {
		ae := ApiEntry{ApiKey: "is-revoked", Revoked: true}
		createTestApiEntry(t, credDir, ae)

		revoked, err := isRevoked(filepath.Join(credDir, "is-revoked.json"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !revoked {
			t.Error("expected key to be revoked")
		}
	})

	t.Run("nonexistent file fails safe", func(t *testing.T) {
		revoked, err := isRevoked("/nonexistent/file.json")
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
		if !revoked {
			t.Error("expected fail-safe: revoked should be true on error")
		}
	})
}

func TestGenerateApiKey(t *testing.T) {
	key := generateApiKey()

	if key == "" {
		t.Fatal("expected non-empty key")
	}

	// UUID format: 8-4-4-4-12
	parts := strings.Split(key, "-")
	if len(parts) != 5 {
		t.Errorf("expected UUID format (5 parts separated by dashes), got %d parts: %q", len(parts), key)
	}

	// Should generate unique keys
	key2 := generateApiKey()
	if key == key2 {
		t.Error("expected two calls to generate different keys")
	}
}

func TestApiEntryJsonRoundTrip(t *testing.T) {
	ae := ApiEntry{
		ApiKey:    "roundtrip-key",
		IssueDate: "2024-02-01T12:00:00Z",
		LastUsed:  "2024-02-10T08:00:00Z",
		Revoked:   false,
		SlackId:   "U88888",
	}

	data, err := json.Marshal(ae)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded ApiEntry
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded != ae {
		t.Errorf("round trip mismatch: got %+v, want %+v", decoded, ae)
	}
}
