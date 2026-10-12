package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/handlers"
)

type noopSender struct{}

func (noopSender) SendHTML(string, string) error { return nil }

func (noopSender) SendDocument(string, string, []byte) error { return nil }

func TestRegisterRoutesDependencyReport(t *testing.T) {
	tests := []struct {
		name       string
		report     *config.DependencyReportConfig
		wantStatus int
	}{
		{name: "section absent returns 404", wantStatus: http.StatusNotFound},
		{name: "section present is registered and authenticated", report: &config.DependencyReportConfig{WebhookToken: "secret", TelegramChatID: "1"}, wantStatus: http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{DependencyReport: tt.report}
			router := mux.NewRouter()
			registerRoutes(router, cfg, handlers.NewWebhookHandler(nil, cfg, zap.NewNop()), noopSender{}, zap.NewNop())

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/webhook/dependency-report", nil))
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
