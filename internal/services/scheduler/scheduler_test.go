package scheduler

import (
	"testing"
	"time"
	_ "time/tzdata"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func mustLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatalf("LoadLocation(%q) error = %v", name, err)
	}
	return loc
}

func TestParseSpec(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		wantErr bool
	}{
		{name: "weekday list", spec: "0 9 * * 1,3,5"},
		{name: "four fields rejected", spec: "0 9 * *", wantErr: true},
		{name: "six fields rejected", spec: "0 0 9 * * 1", wantErr: true},
		{name: "descriptor rejected", spec: "@daily", wantErr: true},
		{name: "out of range rejected", spec: "0 25 * * *", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseSpec(tt.spec); (err != nil) != tt.wantErr {
				t.Errorf("ParseSpec(%q) error = %v, wantErr %v", tt.spec, err, tt.wantErr)
			}
		})
	}
}

func TestNextRunIsInGuayaquil(t *testing.T) {
	loc := mustLocation(t, "America/Guayaquil")
	schedule, err := ParseSpec("0 9 * * 1,3,5")
	if err != nil {
		t.Fatalf("ParseSpec() error = %v", err)
	}

	// Sunday 2026-10-11 12:00 in Guayaquil (UTC-5, no DST).
	next := schedule.Next(time.Date(2026, 10, 11, 12, 0, 0, 0, loc))
	if want := time.Date(2026, 10, 12, 14, 0, 0, 0, time.UTC); !next.Equal(want) {
		t.Fatalf("first run = %s, want %s (Monday 09:00 Guayaquil)", next.Format(time.RFC3339), want.Format(time.RFC3339))
	}

	for i := 0; i < 6; i++ {
		next = schedule.Next(next)
		local := next.In(loc)
		if local.Hour() != 9 || local.Minute() != 0 {
			t.Errorf("run %d at %s, want 09:00 local", i+1, local.Format(time.RFC3339))
		}
		if wd := local.Weekday(); wd != time.Wednesday && wd != time.Friday && wd != time.Monday {
			t.Errorf("run %d on %s, want Mon/Wed/Fri only", i+1, wd)
		}
	}
}

func TestSchedulerReportsNextRunInConfiguredZone(t *testing.T) {
	loc := mustLocation(t, "America/Guayaquil")
	s := New(loc, zap.NewNop())
	if err := s.Add("run_check_updates", "0 9 * * 1,3,5", func() {}); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	s.Start()
	defer s.Stop()

	entries := s.cron.Entries()
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	next := entries[0].Next.In(loc)
	if next.Hour() != 9 || next.Minute() != 0 {
		t.Errorf("next run = %s, want 09:00 Guayaquil", next.Format(time.RFC3339))
	}
	if wd := next.Weekday(); wd != time.Monday && wd != time.Wednesday && wd != time.Friday {
		t.Errorf("next run on %s, want Mon/Wed/Fri", wd)
	}
}

func TestAddRejectsInvalidSpec(t *testing.T) {
	s := New(time.UTC, zap.NewNop())
	if err := s.Add("run_cleanup", "0 9 * *", func() {}); err == nil {
		t.Fatal("Add() error = nil, want error for invalid spec")
	}
}

func TestSchedulerRunsUntilStopped(t *testing.T) {
	// @every needs a parser with the Descriptor option; production uses the
	// strict 5-field parser, so the test injects one that accepts @every.
	s := newWithParser(time.UTC, zap.NewNop(), cron.NewParser(cron.Descriptor))
	runs := make(chan struct{}, 100)
	if err := s.Add("tick", "@every 1s", func() { runs <- struct{}{} }); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	s.Start()
	select {
	case <-runs:
	case <-time.After(5 * time.Second):
		t.Fatal("scheduled function did not run")
	}

	s.Stop()
	before := len(runs)
	time.Sleep(2500 * time.Millisecond)
	if after := len(runs); after != before {
		t.Errorf("function ran %d time(s) after Stop()", after-before)
	}
}
