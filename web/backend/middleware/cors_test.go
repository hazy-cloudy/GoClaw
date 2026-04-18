package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS_AllowsLocalOriginWithoutCredentials(t *testing.T) {
	h := CORS(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	req.Header.Set("Origin", "http://127.0.0.1:3000")
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://127.0.0.1:3000" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("Access-Control-Allow-Credentials = %q, want empty", got)
	}
}

func TestCORS_PreflightOnlyForAllowedOrigins(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	h := CORS(next)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/config", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("allowed preflight status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodOptions, "/api/config", nil)
	req2.Header.Set("Origin", "https://example.com")
	req2.Header.Set("Access-Control-Request-Method", "POST")
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusTeapot {
		t.Fatalf("disallowed preflight status = %d, want %d", rec2.Code, http.StatusTeapot)
	}
}
