package imports

import (
	"github.com/robfig/cron/v3"
	"time"
)

func NewCronManager() *CronManager {
	return &CronManager{
		parser: cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
}

func (c *CronManager) ShouldUpdate(schedule string, lastUpdate time.Time) bool {
	if schedule == "" {
		schedule = "0 */6 * * *"
	}

	sched, err := c.parser.Parse(schedule)
	if err != nil {
		return true
	}

	nextRun := sched.Next(lastUpdate)
	return time.Now().After(nextRun)
}

func (c *CronManager) GetNextUpdateTime(schedule string, from time.Time) time.Time {
	if schedule == "" {
		schedule = "0 */6 * * *"
	}

	sched, err := c.parser.Parse(schedule)
	if err != nil {
		return from.Add(6 * time.Hour)
	}

	return sched.Next(from)
}
