package config

import (
	"strings"
	"testing"
)

const serverOnly = "server:\n  address: \":8080\"\n"

const reportSection = "dependency_report:\n  webhook_token: \"tok\"\n  telegram_chat_id: \"456\"\n"

// loadZone loads content and checks the zone, or the error it must contain.
func loadZone(t *testing.T, content, wantZone, wantErr string) {
	t.Helper()
	cfg, err := loadYAML(t, content)
	if wantErr != "" {
		if err == nil || !strings.Contains(err.Error(), wantErr) {
			t.Fatalf("Load() error = %v, want it to contain %q", err, wantErr)
		}
		return
	}
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Location.String() != wantZone {
		t.Errorf("Location = %s, want %s", cfg.Location, wantZone)
	}
	if cfg.Catalog == nil {
		t.Error("Catalog is nil")
	}
}

func TestLoadDefaultTimezone(t *testing.T) {
	loadZone(t, serverOnly, "America/Guayaquil", "")
}

func TestLoadEuropeMadrid(t *testing.T) {
	loadZone(t, serverOnly+"timezone: Europe/Madrid\n", "Europe/Madrid", "")
}

func TestLoadInvalidTimezoneFails(t *testing.T) {
	loadZone(t, serverOnly+"timezone: Mars/Olympus\n", "", "timezone")
}

func TestDependencyReportDefaults(t *testing.T) {
	cfg, err := loadYAML(t, reportSection)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	r := cfg.DependencyReport
	if r == nil {
		t.Fatal("DependencyReport is nil")
	}
	if r.Path != "/webhook/dependency-report" || r.TokenHeader != "X-Report-Token" || r.MaxBodyBytes != 10485760 ||
		r.DocumentName != "dependency-report-{date}.html" || r.DateLayout != "2006-01-02" {
		t.Errorf("defaults = %+v", r)
	}
}

func TestDependencyReportAbsentStaysNil(t *testing.T) {
	cfg, err := loadYAML(t, serverOnly)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DependencyReport != nil {
		t.Errorf("DependencyReport = %+v, want nil without the section", cfg.DependencyReport)
	}
}

// expectSettingError loads the report section with one setting and checks the error it must name.
func expectSettingError(t *testing.T, setting, wantErr string) {
	t.Helper()
	_, err := loadYAML(t, reportSection+"  "+setting+"\n")
	if err == nil || !strings.Contains(err.Error(), wantErr) {
		t.Errorf("Load() error = %v, want it to contain %q", err, wantErr)
	}
}

func TestDependencyReportInvalidPath(t *testing.T) {
	expectSettingError(t, `path: "webhook/deps"`, "dependency_report.path")
}

func TestDependencyReportDocumentNameNeedsDate(t *testing.T) {
	expectSettingError(t, `document_name: "report.html"`, "dependency_report.document_name")
	expectSettingError(t, `document_name: "a/{date}.html"`, "dependency_report.document_name")
}

func TestDependencyReportZeroBodyLimit(t *testing.T) {
	expectSettingError(t, "max_body_bytes: 0", "dependency_report.max_body_bytes")
}

func TestDependencyReportTokenHeaderAndLayout(t *testing.T) {
	expectSettingError(t, `token_header: "X Deps"`, "dependency_report.token_header")
	expectSettingError(t, `token_header: ""`, "dependency_report.token_header")
	expectSettingError(t, `date_layout: ""`, "dependency_report.date_layout")
}

func TestDependencyReportCustomSettings(t *testing.T) {
	cfg, err := loadYAML(t, reportSection+"  path: \"/hooks/deps\"\n  token_header: \"X-Deps-Token\"\n")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DependencyReport.Path != "/hooks/deps" || cfg.DependencyReport.TokenHeader != "X-Deps-Token" {
		t.Errorf("settings = %+v", cfg.DependencyReport)
	}
}

func TestLoadMessagesUnknownKeyFails(t *testing.T) {
	_, err := loadYAML(t, serverOnly+"messages:\n  caption:\n    headr: \"x\"\n")
	if err == nil || !strings.Contains(err.Error(), "caption.headr") {
		t.Errorf("Load() error = %v, want it to name caption.headr", err)
	}
}

func TestLoadMessagesHeaderOverride(t *testing.T) {
	cfg, err := loadYAML(t, serverOnly+"messages:\n  caption:\n    header: \"🔔 {date}\"\n")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	got, err := cfg.Catalog.Render("caption.header", map[string]string{"date": "2026-10-12"})
	if err != nil || got != "🔔 2026-10-12" {
		t.Errorf("Render() = %q, %v; want the overridden header", got, err)
	}
}
