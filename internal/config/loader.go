package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func Load(path string) (Config, error) {
	var cfg Config

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return cfg, err
	}

	ApplyGlobalSettings(&cfg)

	if cfg.Features == nil {
		cfg.Features = &Features{}
	}

	return cfg, nil
}

func FilterByGroups(cfg Config, filterTopLevelGroups []string) (Config, error) {
	if cfg.Inventory == nil {
		return cfg, fmt.Errorf("no inventory defined")
	}

	var filteredGroups []*Group
	var foundGroups []string

	for _, groupName := range filterTopLevelGroups {
		found := false
		for _, group := range cfg.Inventory.Groups {
			if group.Name == groupName {
				filteredGroups = append(filteredGroups, group)
				foundGroups = append(foundGroups, groupName)
				found = true
				break
			}
		}
		if !found {
			return cfg, fmt.Errorf("top-level group '%s' not found", groupName)
		}
	}

	filteredConfig := cfg
	filteredConfig.Inventory.Groups = filteredGroups

	return filteredConfig, nil
}
