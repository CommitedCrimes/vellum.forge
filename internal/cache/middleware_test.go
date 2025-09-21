package cache

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseCapture_HeaderHandling(t *testing.T) {
	// Create a test response writer
	w := httptest.NewRecorder()

	// Create response capture
	capture := NewResponseCapture(w)

	// Test that headers are set on the underlying writer
	capture.Header().Set("Content-Type", "text/html")
	capture.Header().Set("X-Test", "value")

	// Test that WriteHeader captures status code
	capture.WriteHeader(http.StatusCreated)

	// Test that Write captures body
	testBody := []byte("test response body")
	capture.Write(testBody)

	// Get captured data
	body, headers, statusCode := capture.GetCapturedData()

	// Verify captured data
	if string(body) != "test response body" {
		t.Errorf("Expected body 'test response body', got '%s'", string(body))
	}

	if statusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, statusCode)
	}

	if headers.Get("Content-Type") != "text/html" {
		t.Errorf("Expected Content-Type 'text/html', got '%s'", headers.Get("Content-Type"))
	}

	if headers.Get("X-Test") != "value" {
		t.Errorf("Expected X-Test 'value', got '%s'", headers.Get("X-Test"))
	}

	// Test flush
	err := capture.Flush()
	if err != nil {
		t.Errorf("Flush returned error: %v", err)
	}

	// Verify that the underlying writer received the data
	result := w.Result()
	if result.StatusCode != http.StatusCreated {
		t.Errorf("Expected flushed status code %d, got %d", http.StatusCreated, result.StatusCode)
	}

	if result.Header.Get("Content-Type") != "text/html" {
		t.Errorf("Expected flushed Content-Type 'text/html', got '%s'", result.Header.Get("Content-Type"))
	}

	if w.Body.String() != "test response body" {
		t.Errorf("Expected flushed body 'test response body', got '%s'", w.Body.String())
	}
}
