package main

type Config struct {
	Groups           []*Group          `yaml:"groups"`
	Hosts            []*Server         `yaml:"hosts"`
	Imports          []ImportDirective `yaml:"imports,omitempty"`
	HistorySize      int               `yaml:"history_size,omitempty"`
	FavoritesEnabled bool              `yaml:"favorites_enabled,omitempty"`
	CacheTimeout     int               `yaml:"cache_timeout,omitempty"`
	ImportErrors     []string          `yaml:"-"`
	NoCache          bool              `yaml:"no_cache,omitempty"`
	Auth             *AuthConfig       `yaml:"auth,omitempty"`
	Color            string            `yaml:"color,omitempty"`
	User             string            `yaml:"user,omitempty"`
	Port             int               `yaml:"port,omitempty"`
	ExtraArgs        []string          `yaml:"extra_args,omitempty"`
	SSHBinary        string            `yaml:"ssh_binary,omitempty"`
}

type Group struct {
	Name      string            `yaml:"name"`
	Hosts     []*Server         `yaml:"hosts"`
	Groups    []*Group          `yaml:"groups,omitempty"`
	User      string            `yaml:"user,omitempty"`
	Port      int               `yaml:"port,omitempty"`
	ExtraArgs []string          `yaml:"extra_args,omitempty"`
	SSHBinary string            `yaml:"ssh_binary,omitempty"`
	Color     string            `yaml:"color,omitempty"`
	Imports   []ImportDirective `yaml:"imports,omitempty"`
	NoCache   bool              `yaml:"no_cache,omitempty"`
	Auth      *AuthConfig       `yaml:"auth,omitempty"`
}

type Server struct {
	Name      string   `yaml:"name"`
	Host      string   `yaml:"host"`
	Port      int      `yaml:"port,omitempty"`
	User      string   `yaml:"user,omitempty"`
	ExtraArgs []string `yaml:"extra_args,omitempty"`
	Group     string   `yaml:"group,omitempty"`
	SSHBinary string   `yaml:"ssh_binary,omitempty"`
	Color     string   `yaml:"color,omitempty"`
}

type ImportConfig struct {
	Imports []ImportDirective `yaml:"imports"`
	NoCache bool              `yaml:"no_cache,omitempty"`
	Auth    *AuthConfig       `yaml:"auth,omitempty"`
}

type ImportDirective struct {
	File      string      `yaml:"file"`
	Path      string      `yaml:"path,omitempty"`
	Group     string      `yaml:"group,omitempty"`
	User      string      `yaml:"user,omitempty"`
	Port      int         `yaml:"port,omitempty"`
	ExtraArgs []string    `yaml:"extra_args,omitempty"`
	SSHBinary string      `yaml:"ssh_binary,omitempty"`
	Color     string      `yaml:"color,omitempty"`
	NoCache   bool        `yaml:"no_cache,omitempty"`
	Auth      *AuthConfig `yaml:"auth,omitempty"`
}

type AuthConfig struct {
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Token    string `yaml:"token,omitempty"`
	Header   string `yaml:"header,omitempty"`
}

type ImportData struct {
	Groups []*Group  `yaml:"groups,omitempty"`
	Hosts  []*Server `yaml:"hosts,omitempty"`
}
