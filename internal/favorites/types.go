package favorites

import "github.com/gaetanlhf/SaSHa/internal/config"

type FavoriteItem struct {
	title       string
	description string
	path        string
	color       string
	server      *config.Server
	pathEntries []string
	pathColors  []string
}

type Entry struct {
	Fingerprint string `json:"fingerprint"`
}

type Data struct {
	Entries []Entry `json:"entries"`
}

type ResolvedEntry struct {
	Server *config.Server
	Path   []string
}
