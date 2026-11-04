package leads //nolint:testpackage // Exercise private middleware with a controlled panic handler.

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestLogging(t *testing.T) {
	for _, path := range []string{"/api/leads?email=private@example.com", "/private@example.com"} {
		var output bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&output, nil))
		router := gin.New()
		router.Use(requestLogger(logger), safeRecovery(logger))
		router.GET("/api/leads", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })
		request := httptest.NewRequest(http.MethodGet, path, strings.NewReader("secret body"))
		request.Header.Set("Authorization", "secret token")
		router.ServeHTTP(httptest.NewRecorder(), request)
		var event map[string]any
		if err := json.Unmarshal(output.Bytes(), &event); err != nil {
			t.Fatal(err)
		}
		expectedRoute := "/api/leads"
		if strings.HasPrefix(path, "/private") {
			expectedRoute = "unmatched"
		}
		if event["route"] != expectedRoute || event["method"] != "GET" ||
			event["status"] == nil || event["duration_ms"] == nil {
			t.Fatalf("missing request metadata: %v", event)
		}
		if strings.Contains(output.String(), "private") || strings.Contains(output.String(), "secret") {
			t.Fatal("request data leaked into logs")
		}
	}
}

func TestRecoveryLogging(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	router := gin.New()
	router.Use(requestLogger(logger), safeRecovery(logger))
	router.GET("/panic", func(_ *gin.Context) { panic("private@example.com") })
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/panic", nil))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatal("panic must return 500")
	}
	if !strings.Contains(output.String(), "panic recovered") || !strings.Contains(output.String(), `"status":500`) {
		t.Fatal("panic and failed request must be logged")
	}
	if strings.Contains(output.String(), "private@example.com") {
		t.Fatal("panic details leaked into logs")
	}
}
