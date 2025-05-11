package main

type inheritedSettings struct {
	User      string
	Port      int
	ExtraArgs []string
	SSHBinary string
	Color     string
	NoCache   bool
	Auth      *AuthConfig
}

func getInheritedGroupSettings(group *Group) inheritedSettings {
	inherited := inheritedSettings{
		User:      group.User,
		Port:      group.Port,
		SSHBinary: group.SSHBinary,
		Color:     group.Color,
		NoCache:   group.NoCache,
		Auth:      nil,
	}

	if len(group.ExtraArgs) > 0 {
		inherited.ExtraArgs = append([]string{}, group.ExtraArgs...)
	}

	imports := getImportsFromGroup(group)
	if len(imports) > 0 {
		for _, imp := range imports {
			if imp.Auth != nil && inherited.Auth == nil {
				inherited.Auth = imp.Auth
			}
		}
	}

	return inherited
}

func propagateInheritedSettings(config *Config) {
	for _, group := range config.Groups {
		propagateGroupSettings(group, inheritedSettings{
			User:      group.User,
			Port:      group.Port,
			SSHBinary: group.SSHBinary,
			Color:     group.Color,
			ExtraArgs: group.ExtraArgs,
			NoCache:   group.NoCache,
			Auth:      nil,
		})
	}

	imports := getImportsFromConfig(config)
	if len(imports) > 0 {
		for _, group := range config.Groups {
			for _, imp := range imports {
				if imp.Path == "" || imp.Path == group.Name {
					propagateImportSettings(group, imp)
				}
			}
		}
	}
}

func propagateImportSettings(group *Group, directive ImportDirective) {
	if directive.NoCache {
		group.NoCache = true

		imports := getImportsFromGroup(group)
		for i := range imports {
			imports[i].NoCache = true
		}
	}

	if directive.Auth != nil {
		imports := getImportsFromGroup(group)
		for i := range imports {
			if imports[i].Auth == nil {
				imports[i].Auth = directive.Auth
			}
		}
	}

	for _, subgroup := range group.Groups {
		propagateImportSettings(subgroup, directive)
	}
}

func propagateGroupSettings(group *Group, parentSettings inheritedSettings) {
	settings := inheritedSettings{
		User:      parentSettings.User,
		Port:      parentSettings.Port,
		SSHBinary: parentSettings.SSHBinary,
		Color:     parentSettings.Color,
		NoCache:   parentSettings.NoCache,
		Auth:      parentSettings.Auth,
	}

	if len(parentSettings.ExtraArgs) > 0 {
		settings.ExtraArgs = append([]string{}, parentSettings.ExtraArgs...)
	}

	if group.User != "" {
		settings.User = group.User
	}
	if group.Port != 0 {
		settings.Port = group.Port
	}
	if group.SSHBinary != "" {
		settings.SSHBinary = group.SSHBinary
	}
	if group.Color != "" {
		settings.Color = group.Color
	}
	if len(group.ExtraArgs) > 0 {
		settings.ExtraArgs = append([]string{}, group.ExtraArgs...)
	}
	if group.NoCache {
		settings.NoCache = true
	}

	imports := getImportsFromGroup(group)
	if len(imports) > 0 {
		for _, imp := range imports {
			if imp.Auth != nil {
				settings.Auth = imp.Auth
			}
		}
	}

	for _, host := range group.Hosts {
		if host.User == "" && settings.User != "" {
			host.User = settings.User
		}
		if host.Port == 0 && settings.Port != 0 {
			host.Port = settings.Port
		}
		if host.SSHBinary == "" && settings.SSHBinary != "" {
			host.SSHBinary = settings.SSHBinary
		}
		if host.Color == "" && settings.Color != "" {
			host.Color = settings.Color
		}
		if len(host.ExtraArgs) == 0 && len(settings.ExtraArgs) > 0 {
			host.ExtraArgs = append([]string{}, settings.ExtraArgs...)
		}
	}

	for _, subgroup := range group.Groups {
		propagateGroupSettings(subgroup, settings)
	}
}
