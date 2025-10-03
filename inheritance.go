package main

type inheritedSettings struct {
	User      *string
	Port      *int
	Password  *string
	ExtraArgs []string
	SSHBinary *string
	Color     *string
	NoCache   bool
	Auth      *AuthConfig
}

func getInheritedGroupSettings(group *Group) inheritedSettings {
	inherited := inheritedSettings{
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

	inherited.Auth = getGroupAuth(group)

	if inherited.Auth == nil {
		imports := getImportsFromGroup(group)
		for _, imp := range imports {
			if imp.Auth != nil {
				inherited.Auth = imp.Auth
				break
			}
		}
	}

	return inherited
}

func propagateInheritedSettings(config *Config) {
	if config.Inventory == nil {
		return
	}

	var globalAuth *AuthConfig
	if config.Inventory.Auth != nil {
		globalAuth = config.Inventory.Auth
	} else {
		configImports := getImportsFromInventory(config.Inventory)
		for _, imp := range configImports {
			if imp.Auth != nil {
				globalAuth = imp.Auth
				break
			}
		}
	}

	for _, group := range config.Inventory.Groups {
		propagateGroupSettings(group, inheritedSettings{
			User:      config.Inventory.User,
			Port:      config.Inventory.Port,
			Password:  config.Inventory.Password,
			SSHBinary: config.Inventory.SSHBinary,
			ExtraArgs: config.Inventory.ExtraArgs,
			NoCache:   config.Inventory.NoCache,
			Auth:      globalAuth,
		})
	}

	imports := getImportsFromInventory(config.Inventory)
	if len(imports) > 0 {
		for _, group := range config.Inventory.Groups {
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

	groupAuth := getGroupAuth(group)
	if groupAuth != nil {
		settings.Auth = groupAuth
	}

	if settings.Auth == nil || settings.Auth == parentSettings.Auth {
		imports := getImportsFromGroup(group)
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
