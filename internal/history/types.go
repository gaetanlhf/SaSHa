package history

import (
	"github.com/gaetanlhf/SaSHa/internal/config"
	"time"
)

type HistoryItem struct {
	title          string
	description    string
	path           string
	color          string
	server         *config.Server
	pathEntries    []string
	pathColors     []string
	favoriteStatus bool
}

type Entry struct {
	Fingerprint string    `json:"fingerprint"`
	Timestamp   time.Time `json:"timestamp"`
}

type Data struct {
	Entries []Entry `json:"entries"`
}

type ResolvedEntry struct {
	Server *config.Server
	Path   []string
}
