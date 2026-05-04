package handlers

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHealthCheckHandler(t *testing.T) {
	c, w := newCtx("GET", "/health")
	HealthCheckHandler(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", body["status"])
	}
	if body["service"] != "library-service" {
		t.Errorf("expected service=library-service, got %q", body["service"])
	}
}
