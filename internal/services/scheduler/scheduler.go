// Package scheduler runs registered functions on cron schedules in one fixed
// timezone. It does not know about workflows; callers decide what to run.
package scheduler

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// parser accepts exactly five fields: minute, hour, day of month, month, day of week.
var parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

// ParseSpec validates a 5-field cron expression. Config validation uses it too,
// so the config check and the scheduler can never disagree.
func ParseSpec(spec string) (cron.Schedule, error) {
	return parser.Parse(spec)
}

// Scheduler wraps robfig/cron with named entries and a fixed location.
type Scheduler struct {
	cron   *cron.Cron
	loc    *time.Location
	logger *zap.Logger
	names  map[cron.EntryID]string
}

func New(loc *time.Location, logger *zap.Logger) *Scheduler {
	return newWithParser(loc, logger, parser)
}

func newWithParser(loc *time.Location, logger *zap.Logger, p cron.ScheduleParser) *Scheduler {
	return &Scheduler{
		cron:   cron.New(cron.WithLocation(loc), cron.WithParser(p)),
		loc:    loc,
		logger: logger,
		names:  make(map[cron.EntryID]string),
	}
}

// Add registers fn under name. It returns an error if spec is not valid.
// Call it before Start.
func (s *Scheduler) Add(name, spec string, fn func()) error {
	id, err := s.cron.AddFunc(spec, fn)
	if err != nil {
		return fmt.Errorf("schedule %q: %w", name, err)
	}
	s.names[id] = name
	return nil
}

// Start begins running entries and logs the timezone and the next run of each.
func (s *Scheduler) Start() {
	s.cron.Start()
	s.logger.Info("Scheduler started", zap.String("timezone", s.loc.String()))
	for _, entry := range s.cron.Entries() {
		s.logger.Info("Next scheduled run",
			zap.String("name", s.names[entry.ID]),
			zap.String("next_run", entry.Next.Format(time.RFC3339)))
	}
}

// Stop prevents new runs and waits for runs already in progress to finish.
func (s *Scheduler) Stop() {
	<-s.cron.Stop().Done()
}
