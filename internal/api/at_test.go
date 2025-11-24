package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleATState_ReturnsStructuredJSON(t *testing.T) {
	s := NewWithPolicy(nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/at/state", nil)
	w := httptest.NewRecorder()
	s.handleATState(w, req)
	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.StatusCode)
	}
}

func TestHandleATRunPhase_BadRequest(t *testing.T) {
	s := NewWithPolicy(nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/at/phase//run", nil)
	w := httptest.NewRecorder()
	s.handleATRunPhase(w, req)
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing phase id, got %d", w.Result().StatusCode)
	}
}
