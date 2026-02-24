package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGetCurrentTimestamp(t *testing.T) {
	before := time.Now().UnixNano() / int64(time.Millisecond)
	ts := getCurrentTimestamp()
	after := time.Now().UnixNano() / int64(time.Millisecond)

	if ts < before || ts > after {
		t.Errorf("timestamp %d not between %d and %d", ts, before, after)
	}
}

func TestStaticFileServer(t *testing.T) {
	router := gin.New()
	router.GET("/usage", staticFileServer)

	req := httptest.NewRequest(http.MethodGet, "/usage", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("expected content type text/html; charset=utf-8, got %q", contentType)
	}

	body := w.Body.String()
	if len(body) == 0 {
		t.Error("expected non-empty body for usage page")
	}
}

func TestPostApiKeyHandler_InvalidJSON(t *testing.T) {
	router := gin.New()
	router.POST("/api", postApiKeyHandler)

	req := httptest.NewRequest(http.MethodPost, "/api", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDeleteApiKeyHandler_NoKey(t *testing.T) {
	credDir := setupTestCredentialsDir(t)
	_ = credDir

	router := gin.New()
	router.DELETE("/api", deleteApiKeyHandler)

	req := httptest.NewRequest(http.MethodDelete, "/api", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// No API key header means unauthorized
	if w.Code != http.StatusNetworkAuthenticationRequired {
		t.Errorf("expected status %d, got %d", http.StatusNetworkAuthenticationRequired, w.Code)
	}
}

func TestDeleteApiKeyHandler_ValidKey(t *testing.T) {
	credDir := setupTestCredentialsDir(t)

	ae := ApiEntry{
		ApiKey:  "delete-test-key",
		SlackId: "U1234",
		Revoked: false,
	}
	createTestApiEntry(t, credDir, ae)

	router := gin.New()
	router.DELETE("/api", deleteApiKeyHandler)

	req := httptest.NewRequest(http.MethodDelete, "/api", nil)
	req.Header.Set("X-API-Key", "delete-test-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %q", resp["status"])
	}
}

func TestDeleteApiKeyHandler_AlreadyRevoked(t *testing.T) {
	credDir := setupTestCredentialsDir(t)

	ae := ApiEntry{
		ApiKey:  "already-revoked-key",
		SlackId: "U1234",
		Revoked: true,
	}
	createTestApiEntry(t, credDir, ae)

	router := gin.New()
	router.DELETE("/api", deleteApiKeyHandler)

	req := httptest.NewRequest(http.MethodDelete, "/api", nil)
	req.Header.Set("X-API-Key", "already-revoked-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNetworkAuthenticationRequired {
		t.Errorf("expected status %d, got %d", http.StatusNetworkAuthenticationRequired, w.Code)
	}
}

func TestUploadHandler_NoApiKey(t *testing.T) {
	setupTestCredentialsDir(t)

	router := gin.New()
	router.POST("/upload", uploadHandler)

	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestUploadHandler_InvalidApiKey(t *testing.T) {
	setupTestCredentialsDir(t)

	router := gin.New()
	router.POST("/upload", uploadHandler)

	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	req.Header.Set("X-API-Key", "bogus-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestUploadHandler_ValidUpload(t *testing.T) {
	credDir := setupTestCredentialsDir(t)
	tmpDir := t.TempDir()
	uploadsDir := filepath.Join(tmpDir, "uploads")
	os.MkdirAll(uploadsDir, 0o755)
	viper.Set("uploads_dir", uploadsDir)

	ae := ApiEntry{
		ApiKey:  "upload-test-key",
		SlackId: "U9999",
		Revoked: false,
	}
	createTestApiEntry(t, credDir, ae)

	// Build multipart form body
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("image", "test.jpg")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	io.WriteString(part, "fake image data")
	writer.WriteField("caption", "Test caption")
	writer.Close()

	router := gin.New()
	router.POST("/upload", uploadHandler)

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-API-Key", "upload-test-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["message"] != "Upload successful" {
		t.Errorf("expected message 'Upload successful', got %q", resp["message"])
	}

	// Verify the image file and JSON file were created in uploads dir
	entries, _ := os.ReadDir(uploadsDir)
	imageCount := 0
	jsonCount := 0
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".jpg" {
			imageCount++
		}
		if filepath.Ext(e.Name()) == ".json" {
			jsonCount++
		}
	}
	if imageCount != 1 {
		t.Errorf("expected 1 image file in uploads, got %d", imageCount)
	}
	if jsonCount != 1 {
		t.Errorf("expected 1 json file in uploads, got %d", jsonCount)
	}
}

func TestUploadHandler_NoImageField(t *testing.T) {
	credDir := setupTestCredentialsDir(t)

	ae := ApiEntry{
		ApiKey:  "upload-no-image-key",
		SlackId: "U8888",
		Revoked: false,
	}
	createTestApiEntry(t, credDir, ae)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("caption", "no image here")
	writer.Close()

	router := gin.New()
	router.POST("/upload", uploadHandler)

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-API-Key", "upload-no-image-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUploadAuthorization_RevokedKey(t *testing.T) {
	credDir := setupTestCredentialsDir(t)

	ae := ApiEntry{
		ApiKey:  "revoked-upload-key",
		SlackId: "U7777",
		Revoked: true,
	}
	createTestApiEntry(t, credDir, ae)

	router := gin.New()
	router.POST("/test", func(c *gin.Context) {
		err := uploadAuthorization(c)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("X-API-Key", "revoked-upload-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Revoked keys should not pass authorization
	if w.Code == http.StatusOK {
		t.Error("expected revoked key to fail authorization")
	}
}

func TestImageInfoJsonFields(t *testing.T) {
	info := ImageInfo{
		ImagePath: "/path/to/img.png",
		Caption:   "hello world",
		ApiKey:    "key-abc",
	}

	data, _ := json.Marshal(info)
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	expectedKeys := []string{"image_path", "caption", "api_key"}
	for _, key := range expectedKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected JSON key %q to be present", key)
		}
	}
}
