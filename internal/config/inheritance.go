package config

import "gopkg.in/yaml.v3"

type InheritedSettings struct {
	User      *string
	Port      *int
	Password  *string
	ExtraArgs []string
	SSHBinary *string
	Color     *string
	NoCache   bool
	Auth      *AuthConfig
}

func GetInheritedGroupSettings(group *Group) InheritedSettings {
	inherited := InheritedSettings{
		User:      group.User,
		Port:      group.Port,
		Password:  group.Password,
		SSHBinary: group.SSHBinary,
		Color:     group.Color,
		NoCache:   group.NoCache,
	}

	if len(group.ExtraArgs) > 0 {
		inherited.ExtraArgs = append([]string{}, group.ExtraArgs...)
	}

	inherited.Auth = GetGroupAuth(group)

	if inherited.Auth == nil {
		imports := GetImportsFromGroup(group)
		for _, imp := range imports {
			if imp.Auth != nil {
				inherited.Auth = imp.Auth
				break
			}
		}
	}

	return inherited
}

func PropagateInheritedSettings(cfg *Config) {
	if cfg.Inventory == nil {
		return
	}

	var globalAuth *AuthConfig
	if cfg.Inventory.Auth != nil {
		globalAuth = cfg.Inventory.Auth
	} else {
		configImports := GetImportsFromInventory(cfg.Inventory)
		for _, imp := range configImports {
			if imp.Auth != nil {
				globalAuth = imp.Auth
				break
			}
		}
	}

	for _, group := range cfg.Inventory.Groups {
		propagateGroupSettings(group, InheritedSettings{
			User:      cfg.Inventory.User,
			Port:      cfg.Inventory.Port,
			Password:  cfg.Inventory.Password,
			SSHBinary: cfg.Inventory.SSHBinary,
			ExtraArgs: cfg.Inventory.ExtraArgs,
			NoCache:   cfg.Inventory.NoCache,
			Auth:      globalAuth,
		})
	}

	imports := GetImportsFromInventory(cfg.Inventory)
	if len(imports) > 0 {
		for _, group := range cfg.Inventory.Groups {
			for _, imp := range imports {
				if imp.Path == "" || imp.Path == group.Name {
					PropagateImportSettings(group, imp)
				}
			}
		}
	}
}

func PropagateImportSettings(group *Group, directive ImportDirective) {
	if directive.NoCache {
		group.NoCache = true

		imports := GetImportsFromGroup(group)
		for i := range imports {
			imports[i].NoCache = true
		}
	}

	if directive.Auth != nil {
		imports := GetImportsFromGroup(group)
		for i := range imports {
			if imports[i].Auth == nil {
				imports[i].Auth = directive.Auth
			}
		}
	}

	for _, subgroup := range group.Groups {
		PropagateImportSettings(subgroup, directive)
	}
}

func propagateGroupSettings(group *Group, parentSettings InheritedSettings) {
	settings := InheritedSettings{
		User:      parentSettings.User,
		Port:      parentSettings.Port,
		Password:  parentSettings.Password,
		SSHBinary: parentSettings.SSHBinary,
		NoCache:   parentSettings.NoCache,
		Auth:      parentSettings.Auth,
	}

	if len(parentSettings.ExtraArgs) > 0 {
		settings.ExtraArgs = append([]string{}, parentSettings.ExtraArgs...)
	}

	if group.User != nil {
		settings.User = group.User
	}
	if group.Port != nil {
		settings.Port = group.Port
	}
	if group.Password != nil {
		settings.Password = group.Password
	}
	if group.SSHBinary != nil {
		settings.SSHBinary = group.SSHBinary
	}
	if len(group.ExtraArgs) > 0 {
		settings.ExtraArgs = append([]string{}, group.ExtraArgs...)
	}
	if group.NoCache {
		settings.NoCache = true
	}

	groupAuth := GetGroupAuth(group)
	if groupAuth != nil {
		settings.Auth = groupAuth
	}

	if settings.Auth == nil || settings.Auth == parentSettings.Auth {
		imports := GetImportsFromGroup(group)
		for _, imp := range imports {
			if imp.Auth != nil {
				settings.Auth = imp.Auth
				break
			}
		}
	}

	for _, host := range group.Hosts {
		if host.User == nil && settings.User != nil {
			host.User = settings.User
		}
		if host.Port == nil && settings.Port != nil {
			host.Port = settings.Port
		}
		if host.Password == nil && settings.Password != nil {
			host.Password = settings.Password
		}
		if host.SSHBinary == nil && settings.SSHBinary != nil {
			host.SSHBinary = settings.SSHBinary
		}
		if len(host.ExtraArgs) == 0 && len(settings.ExtraArgs) > 0 {
			host.ExtraArgs = append([]string{}, settings.ExtraArgs...)
		}
	}

	for _, subgroup := range group.Groups {
		propagateGroupSettings(subgroup, settings)
	}
}

func GetImportsFromInventory(inventory *Inventory) []ImportDirective {
	var importConfig ImportConfig

	data, err := yaml.Marshal(inventory)
	if err != nil {
		return nil
	}

	err = yaml.Unmarshal(data, &importConfig)
	if err != nil {
		return nil
	}

	return importConfig.Imports
}

func GetImportsFromGroup(group *Group) []ImportDirective {
	var importConfig ImportConfig

	data, err := yaml.Marshal(group)
	if err != nil {
		return nil
	}

	err = yaml.Unmarshal(data, &importConfig)
	if err != nil {
		return nil
	}

	return importConfig.Imports
}

func GetGroupAuth(group *Group) *AuthConfig {
	data, err := yaml.Marshal(group)
	if err != nil {
		return nil
	}

	type GroupWithAuth struct {
		Auth *AuthConfig `yaml:"auth,omitempty"`
	}

	var groupWithAuth GroupWithAuth
	if err := yaml.Unmarshal(data, &groupWithAuth); err != nil {
		return nil
	}

	return groupWithAuth.Auth
}

func ApplyGlobalSettings(cfg *Config) {
	if cfg.Inventory == nil {
		return
	}

	for _, server := range cfg.Inventory.Hosts {
		applyGlobalSettingsToServer(cfg, server)
	}

	for _, group := range cfg.Inventory.Groups {
		applyGlobalSettingsToGroup(cfg, group)
	}
}

func applyGlobalSettingsToGroup(cfg *Config, group *Group) {
	if cfg.Inventory == nil {
		return
	}
	if group.User == nil && cfg.Inventory.User != nil {
		group.User = cfg.Inventory.User
	}
	if group.Port == nil && cfg.Inventory.Port != nil {
		group.Port = cfg.Inventory.Port
	}
	if group.Password == nil && cfg.Inventory.Password != nil {
		group.Password = cfg.Inventory.Password
	}
	if group.SSHBinary == nil && cfg.Inventory.SSHBinary != nil {
		group.SSHBinary = cfg.Inventory.SSHBinary
	}
	if len(group.ExtraArgs) == 0 && len(cfg.Inventory.ExtraArgs) > 0 {
		group.ExtraArgs = append([]string{}, cfg.Inventory.ExtraArgs...)
	}
	if group.Auth == nil && cfg.Inventory.Auth != nil {
		group.Auth = cfg.Inventory.Auth
	}
	if cfg.Inventory.NoCache {
		group.NoCache = true
	}

	for _, server := range group.Hosts {
		applyGlobalSettingsToServer(cfg, server)
	}

	for _, subgroup := range group.Groups {
		applyGlobalSettingsToGroup(cfg, subgroup)
	}
}

func applyGlobalSettingsToServer(cfg *Config, server *Server) {
	if cfg.Inventory == nil {
		return
	}
	if server.User == nil && cfg.Inventory.User != nil {
		server.User = cfg.Inventory.User
	}
	if server.Port == nil && cfg.Inventory.Port != nil {
		server.Port = cfg.Inventory.Port
	}
	if server.Password == nil && cfg.Inventory.Password != nil {
		server.Password = cfg.Inventory.Password
	}
	if server.SSHBinary == nil && cfg.Inventory.SSHBinary != nil {
		server.SSHBinary = cfg.Inventory.SSHBinary
	}
	if len(server.ExtraArgs) == 0 && len(cfg.Inventory.ExtraArgs) > 0 {
		server.ExtraArgs = append([]string{}, cfg.Inventory.ExtraArgs...)
	}
}
