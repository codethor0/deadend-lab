package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"deadend-lab/pkg/dee"
)

func TestScenarioIgnoresSeedParam(t *testing.T) {
	body := []byte(`{"seed": 42}`)
	req := httptest.NewRequest(http.MethodPost, "/scenario/safe", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handleScenario(w, req, dee.Safe)
	if w.Code != http.StatusOK {
		t.Errorf("scenario with seed param: got status %d", w.Code)
	}
	var res ScenarioResult
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatalf("scenario JSON: %v", err)
	}
	if res.ReasonCode != "ok" && res.ReasonCode != "error" {
		t.Errorf("reason_code must be 'ok' or 'error', got %q", res.ReasonCode)
	}
}

func TestHealthSchema(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handleHealth(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("health: got status %d", w.Code)
	}
	var m map[string]string
	if err := json.NewDecoder(w.Body).Decode(&m); err != nil {
		t.Fatalf("health JSON: %v", err)
	}
	if m["status"] != "ok" {
		t.Errorf("health status: got %q", m["status"])
	}
}

func TestScenarioSchemaAndUniformFailure(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/scenario/safe", nil)
	w := httptest.NewRecorder()
	handleScenario(w, req, dee.Safe)
	if w.Code != http.StatusOK {
		t.Errorf("scenario: got status %d", w.Code)
	}
	var res ScenarioResult
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatalf("scenario JSON: %v", err)
	}
	if res.ReasonCode != "ok" && res.ReasonCode != "error" {
		t.Errorf("reason_code must be 'ok' or 'error', got %q", res.ReasonCode)
	}
	if res.SessionIDTrunc != "" && len(res.SessionIDTrunc) != 16 {
		t.Errorf("session_id_trunc must be 8 bytes hex (16 chars) or empty, got len %d", len(res.SessionIDTrunc))
	}
}
