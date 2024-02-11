package main

import (
	"testing"

	"github.com/spf13/viper"
)

// MockLogger is a mock logger for testing purposes
type MockLogger struct{}

func (l *MockLogger) Debugln(args ...interface{}) {}
func (l *MockLogger) Errorln(args ...interface{}) {}

func TestValidatePortVsBaseURL(t *testing.T) {
	// Mock configuration
	viper.Set("base_url", "http://example.com:8080")
	viper.Set("port", "8081")

	validatePortVsBaseURL()

	if port := viper.GetString("port"); port != "8080" {
		t.Errorf("Expected port to be set to 8080, got %s", port)
	}
}

func TestValidatePortVsBaseURL_NoBaseURL(t *testing.T) {
	viper.Set("base_url", "")
	viper.Set("port", "8081")

	validatePortVsBaseURL()

	if port := viper.GetString("port"); port != "8081" {
		t.Errorf("Expected port to remain unchanged, got %s", port)
	}
}

func TestValidatePortVsBaseURL_InvalidBaseURL(t *testing.T) {
	viper.Set("base_url", "invalid-url")
	viper.Set("port", "8081")

	validatePortVsBaseURL()
	// FIXME catch the error log
}

func TestValidatePortVsBaseURL_BaseURLWithoutPort(t *testing.T) {
	// Mock configuration
	viper.Set("base_url", "http://example.com")
	viper.Set("port", "8081")

	validatePortVsBaseURL()

	if port := viper.GetString("port"); port != "8081" {
		t.Errorf("Expected port to remain unchanged, got %s", port)
	}
}

func TestValidatePortVsBaseURL_Port8080(t *testing.T) {
	viper.Set("base_url", "http://example.com:8080")
	viper.Set("port", "8080")

	validatePortVsBaseURL()
	// Expect no message to be logged
}

func TestValidatePortVsBaseURL_CustomPort(t *testing.T) {
	viper.Set("base_url", "http://example.com:9000")
	viper.Set("port", "8081")

	validatePortVsBaseURL()
	// FIXME  overridden message to be logged
}
