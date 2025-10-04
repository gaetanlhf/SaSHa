package imports

import (
	"github.com/gaetanlhf/SaSHa/internal/config"
	"github.com/robfig/cron/v3"
)

type CacheManager struct {
	cronManager *CronManager
}

type CronManager struct {
	parser cron.Parser
}

type ElementPosition struct {
	Type  string
	Index int
	Line  int
}

type OrderTracker struct {
	positions map[string][]ElementPosition
}

type inheritedSettings struct {
	User      *string
	Port      *int
	Password  *string
	ExtraArgs []string
	SSHBinary *string
	Color     *string
	NoCache   bool
	Auth      *config.AuthConfig
}
