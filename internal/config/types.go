package config

import "time"

type Config struct {
	Inventory    *Inventory `yaml:"inventory,omitempty"`
	Features     *Features  `yaml:"features,omitempty"`
	ImportErrors []string   `yaml:"-"`
}

type Inventory struct {
	User      *string     `yaml:"user,omitempty"`
	Port      *int        `yaml:"port,omitempty"`
	Password  *string     `yaml:"password,omitempty"`
	SSHBinary *string     `yaml:"ssh_binary,omitempty"`
	ExtraArgs []string    `yaml:"extra_args,omitempty"`
	Color     *string     `yaml:"color,omitempty"`
	NoCache   bool        `yaml:"no_cache,omitempty"`
	Auth      *AuthConfig `yaml:"auth,omitempty"`
	Groups    []*Group    `yaml:"groups,omitempty"`
	Hosts     []*Server   `yaml:"hosts,omitempty"`
}

type Features struct {
	HistorySize      int    `yaml:"history_size,omitempty"`
	FavoritesEnabled bool   `yaml:"favorites_enabled,omitempty"`
	CacheSchedule    string `yaml:"cache_schedule,omitempty"`
	AllowWebImports  bool   `yaml:"allow_web_imports,omitempty"`
}

type Server struct {
	Name      string   `yaml:"name"`
	Host      string   `yaml:"host"`
	User      *string  `yaml:"user,omitempty"`
	Port      *int     `yaml:"port,omitempty"`
	Password  *string  `yaml:"password,omitempty"`
	SSHBinary *string  `yaml:"ssh_binary,omitempty"`
	Color     *string  `yaml:"color,omitempty"`
	ExtraArgs []string `yaml:"extra_args,omitempty"`
	Group     string   `yaml:"group,omitempty"`
}

type Group struct {
	Name      string            `yaml:"name"`
	User      *string           `yaml:"user,omitempty"`
	Port      *int              `yaml:"port,omitempty"`
	Password  *string           `yaml:"password,omitempty"`
	SSHBinary *string           `yaml:"ssh_binary,omitempty"`
	Color     *string           `yaml:"color,omitempty"`
	ExtraArgs []string          `yaml:"extra_args,omitempty"`
	NoCache   bool              `yaml:"no_cache,omitempty"`
	Groups    []*Group          `yaml:"groups,omitempty"`
	Hosts     []*Server         `yaml:"hosts,omitempty"`
	Imports   []ImportDirective `yaml:"imports,omitempty"`
	Auth      *AuthConfig       `yaml:"auth,omitempty"`
}

type ImportDirective struct {
	File      string      `yaml:"file"`
	Path      string      `yaml:"path,omitempty"`
	User      *string     `yaml:"user,omitempty"`
	Port      *int        `yaml:"port,omitempty"`
	Password  *string     `yaml:"password,omitempty"`
	SSHBinary *string     `yaml:"ssh_binary,omitempty"`
	Color     *string     `yaml:"color,omitempty"`
	ExtraArgs []string    `yaml:"extra_args,omitempty"`
	NoCache   bool        `yaml:"no_cache,omitempty"`
	Auth      *AuthConfig `yaml:"auth,omitempty"`
}

type ImportData struct {
	Groups []*Group  `yaml:"groups,omitempty"`
	Hosts  []*Server `yaml:"hosts,omitempty"`
}

type ImportConfig struct {
	Imports []ImportDirective `yaml:"imports,omitempty"`
}

type AuthConfig struct {
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Token    string `yaml:"token,omitempty"`
	Header   string `yaml:"header,omitempty"`
}

type CacheMetadata struct {
	URL        string    `yaml:"url"`
	LastUpdate time.Time `yaml:"last_update"`
}

type CachedImport struct {
	Metadata CacheMetadata `yaml:"metadata"`
	Data     ImportData    `yaml:"data"`
}
