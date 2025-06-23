package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Save original env vars
	originalAPIKey := os.Getenv("ARK_API_KEY")
	originalModel := os.Getenv("MODEL")
	originalBaseURL := os.Getenv("BASE_URL")
	originalStream := os.Getenv("STREAM")

	// Clean up after test
	defer func() {
		os.Setenv("ARK_API_KEY", originalAPIKey)
		os.Setenv("MODEL", originalModel)
		os.Setenv("BASE_URL", originalBaseURL)
		os.Setenv("STREAM", originalStream)
	}()

	// Test with empty environment
	os.Unsetenv("ARK_API_KEY")
	os.Unsetenv("MODEL")
	os.Unsetenv("BASE_URL")
	os.Unsetenv("STREAM")

	config := LoadConfig()

	if config.APIKey != "" {
		t.Errorf("Expected empty APIKey, got %s", config.APIKey)
	}
	if config.Model != "" {
		t.Errorf("Expected empty Model, got %s", config.Model)
	}
	if config.BaseURL != "" {
		t.Errorf("Expected empty BaseURL, got %s", config.BaseURL)
	}
	if config.Stream {
		t.Errorf("Expected Stream to be false, got true")
	}

	// Test with set environment variables
	os.Setenv("ARK_API_KEY", "test-key")
	os.Setenv("MODEL", "test-model")
	os.Setenv("BASE_URL", "https://test.com")
	os.Setenv("STREAM", "true")

	config = LoadConfig()

	if config.APIKey != "test-key" {
		t.Errorf("Expected APIKey 'test-key', got %s", config.APIKey)
	}
	if config.Model != "test-model" {
		t.Errorf("Expected Model 'test-model', got %s", config.Model)
	}
	if config.BaseURL != "https://test.com" {
		t.Errorf("Expected BaseURL 'https://test.com', got %s", config.BaseURL)
	}
	if !config.Stream {
		t.Errorf("Expected Stream to be true, got false")
	}
}

func TestConfigStruct(t *testing.T) {
	config := Config{
		APIKey:  "test-api-key",
		Model:   "test-model",
		BaseURL: "https://test.example.com",
		Stream:  true,
	}

	if config.APIKey != "test-api-key" {
		t.Errorf("Expected APIKey 'test-api-key', got %s", config.APIKey)
	}
	if config.Model != "test-model" {
		t.Errorf("Expected Model 'test-model', got %s", config.Model)
	}
	if config.BaseURL != "https://test.example.com" {
		t.Errorf("Expected BaseURL 'https://test.example.com', got %s", config.BaseURL)
	}
	if !config.Stream {
		t.Errorf("Expected Stream to be true, got false")
	}
}
