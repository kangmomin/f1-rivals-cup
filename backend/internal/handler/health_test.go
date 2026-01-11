package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestHealthHandler_Check(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewHealthHandler()

	// Test
	if err := h.Check(c); err != nil {
		t.Fatalf("HealthHandler.Check() error = %v", err)
	}

	// Assertions
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	var response HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response.Status)
	}
}

func TestHealthHandler_Check_ResponseFormat(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewHealthHandler()

	// Test
	_ = h.Check(c)

	// Verify JSON format
	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json; charset=UTF-8" {
		t.Errorf("Expected Content-Type 'application/json; charset=UTF-8', got '%s'", contentType)
	}

	// Verify exact JSON structure
	expected := `{"status":"ok"}`
	actual := rec.Body.String()

	// Parse and re-marshal to compare
	var expectedJSON, actualJSON map[string]interface{}
	json.Unmarshal([]byte(expected), &expectedJSON)
	json.Unmarshal([]byte(actual), &actualJSON)

	if expectedJSON["status"] != actualJSON["status"] {
		t.Errorf("Expected response %s, got %s", expected, actual)
	}
}
