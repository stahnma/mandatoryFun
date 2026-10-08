package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/slack-go/slack"
	"github.com/spf13/viper"
)

func TestIsImage(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     bool
	}{
		{"jpg lowercase", "photo.jpg", true},
		{"jpeg lowercase", "photo.jpeg", true},
		{"png lowercase", "photo.png", true},
		{"gif lowercase", "photo.gif", true},
		{"JPG uppercase", "photo.JPG", true},
		{"JPEG uppercase", "photo.JPEG", true},
		{"PNG uppercase", "photo.PNG", true},
		{"GIF uppercase", "photo.GIF", true},
		{"mixed case", "photo.JpG", true},
		{"txt file", "readme.txt", false},
		{"json file", "data.json", false},
		{"no extension", "photo", false},
		{"empty string", "", false},
		{"dot only", ".", false},
		{"webp not supported", "photo.webp", false},
		{"bmp not supported", "photo.bmp", false},
		{"svg not supported", "photo.svg", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isImage(tt.fileName); got != tt.want {
				t.Errorf("isImage(%q) = %v, want %v", tt.fileName, got, tt.want)
			}
		})
	}
}

func TestHasJsonExtension(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     bool
	}{
		{"json lowercase", "data.json", true},
		{"JSON uppercase", "data.JSON", true},
		{"Json mixed", "data.Json", true},
		{"txt file", "data.txt", false},
		{"no extension", "data", false},
		{"empty string", "", false},
		{"json in name", "json.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasJsonExtension(tt.fileName); got != tt.want {
				t.Errorf("hasJsonExtension(%q) = %v, want %v", tt.fileName, got, tt.want)
			}
		})
	}
}

func TestIsJson(t *testing.T) {
	tmpDir := t.TempDir()
	discardDir := filepath.Join(tmpDir, "discard")
	os.MkdirAll(discardDir, 0o755)
	viper.Set("discard_dir", discardDir)

	t.Run("valid json file", func(t *testing.T) {
		f := filepath.Join(tmpDir, "valid.json")
		os.WriteFile(f, []byte(`{"key":"value"}`), 0o644)
		if !isJson(f) {
			t.Error("expected isJson to return true for valid json file")
		}
	})

	t.Run("invalid json content", func(t *testing.T) {
		f := filepath.Join(tmpDir, "invalid.json")
		os.WriteFile(f, []byte(`not json`), 0o644)
		if isJson(f) {
			t.Error("expected isJson to return false for invalid json content")
		}
	})

	t.Run("non-json extension", func(t *testing.T) {
		f := filepath.Join(tmpDir, "data.txt")
		os.WriteFile(f, []byte(`{"key":"value"}`), 0o644)
		if isJson(f) {
			t.Error("expected isJson to return false for non-json extension")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		f := filepath.Join(tmpDir, "nope.json")
		if isJson(f) {
			t.Error("expected isJson to return false for nonexistent file")
		}
	})
}

func TestHandleNewFile_IgnoresImageFiles(t *testing.T) {
	// handleNewFile should return early for image files (no json processing)
	tmpDir := t.TempDir()
	imgFile := filepath.Join(tmpDir, "test.jpg")
	os.WriteFile(imgFile, []byte("fake image"), 0o644)

	// This should not panic or error - it just returns early
	handleNewFile(imgFile)
}

func TestHandleNewFile_InvalidJsonContent(t *testing.T) {
	tmpDir := t.TempDir()
	discardDir := filepath.Join(tmpDir, "discard")
	processedDir := filepath.Join(tmpDir, "processed")
	os.MkdirAll(discardDir, 0o755)
	os.MkdirAll(processedDir, 0o755)

	viper.Set("discard_dir", discardDir)
	viper.Set("processed_dir", processedDir)

	f := filepath.Join(tmpDir, "bad.json")
	os.WriteFile(f, []byte(`not json at all`), 0o644)

	// isJson will move it to discard, handleNewFile should not panic
	handleNewFile(f)
}

func TestImageInfoJsonRoundTrip(t *testing.T) {
	info := ImageInfo{
		ImagePath: "/tmp/test.jpg",
		Caption:   "Test caption",
		ApiKey:    "abc-123",
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("failed to marshal ImageInfo: %v", err)
	}

	var decoded ImageInfo
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("failed to unmarshal ImageInfo: %v", err)
	}

	if decoded.ImagePath != info.ImagePath {
		t.Errorf("ImagePath = %q, want %q", decoded.ImagePath, info.ImagePath)
	}
	if decoded.Caption != info.Caption {
		t.Errorf("Caption = %q, want %q", decoded.Caption, info.Caption)
	}
	if decoded.ApiKey != info.ApiKey {
		t.Errorf("ApiKey = %q, want %q", decoded.ApiKey, info.ApiKey)
	}
}

type fakeLister struct {
	pages [][]slack.Channel
	calls int
	err   error
}

func (f *fakeLister) GetConversations(params *slack.GetConversationsParameters) ([]slack.Channel, string, error) {
	if f.err != nil {
		return nil, "", f.err
	}
	page := f.pages[f.calls]
	f.calls++
	next := ""
	if f.calls < len(f.pages) {
		next = "cursor"
	}
	return page, next, nil
}

func channel(id, name string) slack.Channel {
	var c slack.Channel
	c.ID = id
	c.Name = name
	return c
}

func TestResolveChannelID(t *testing.T) {
	pages := [][]slack.Channel{
		{channel("C11111111", "general")},
		{channel("C22222222", "cspp"), channel("G33333333", "secret")},
	}

	tests := []struct {
		name      string
		input     string
		want      string
		wantCalls int
		wantErr   bool
	}{
		{"public ID passes through", "C0123ABCDE", "C0123ABCDE", 0, false},
		{"private ID passes through", "G0123ABCDE", "G0123ABCDE", 0, false},
		{"name with hash", "#cspp", "C22222222", 2, false},
		{"name without hash", "cspp", "C22222222", 2, false},
		{"private channel name", "secret", "G33333333", 2, false},
		{"unknown name", "#nope", "", 2, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channelIDCache = map[string]string{}
			f := &fakeLister{pages: pages}
			got, err := resolveChannelID(f, tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveChannelID(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("resolveChannelID(%q) = %q, want %q", tt.input, got, tt.want)
			}
			if f.calls != tt.wantCalls {
				t.Errorf("GetConversations called %d times, want %d", f.calls, tt.wantCalls)
			}
		})
	}
}

func TestResolveChannelIDCaches(t *testing.T) {
	channelIDCache = map[string]string{}
	f := &fakeLister{pages: [][]slack.Channel{{channel("C22222222", "cspp")}}}
	for i := 0; i < 3; i++ {
		if got, err := resolveChannelID(f, "#cspp"); err != nil || got != "C22222222" {
			t.Fatalf("resolveChannelID = %q, %v", got, err)
		}
	}
	if f.calls != 1 {
		t.Errorf("GetConversations called %d times, want 1", f.calls)
	}
}

func TestResolveChannelIDAPIError(t *testing.T) {
	channelIDCache = map[string]string{}
	f := &fakeLister{err: errors.New("missing_scope")}
	if _, err := resolveChannelID(f, "#cspp"); err == nil {
		t.Fatal("expected error")
	}
}
