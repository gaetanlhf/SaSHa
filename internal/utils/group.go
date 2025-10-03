package utils

import (
	"strings"

	"github.com/gaetanlhf/sasha/internal/config"
)

func FindGroupByPath(cfg *config.Config, path string) *config.Group {
	if cfg.Inventory == nil {
		return nil
	}

	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return nil
	}

	var currentGroup *config.Group
	for _, group := range cfg.Inventory.Groups {
		if group.Name == parts[0] {
			currentGroup = group
			break
		}
	}

	if currentGroup == nil {
		return nil
	}

	for i := 1; i < len(parts); i++ {
		found := false
		for _, subgroup := range currentGroup.Groups {
			if subgroup.Name == parts[i] {
				currentGroup = subgroup
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}

	return currentGroup
}

func FindGroupByPathSlice(cfg *config.Config, path []string) *config.Group {
	if cfg.Inventory == nil {
		return nil
	}

	if len(path) == 0 {
		return nil
	}

	var currentGroup *config.Group
	for _, group := range cfg.Inventory.Groups {
		if group.Name == path[0] {
			currentGroup = group
			break
		}
	}

	if currentGroup == nil {
		return nil
	}

	for i := 1; i < len(path); i++ {
		found := false
		for _, subgroup := range currentGroup.Groups {
			if subgroup.Name == path[i] {
				currentGroup = subgroup
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}

	return currentGroup
}

func ResolveEffectiveColor(cfg *config.Config, path []string) string {
	defaultColor := "#FFFFFF"

	if cfg.Inventory != nil && cfg.Inventory.Color != nil && *cfg.Inventory.Color != "" {
		defaultColor = *cfg.Inventory.Color
	}

	if len(path) == 0 {
		return defaultColor
	}

	effectiveColor := defaultColor
	pathSoFar := []string{}

	for _, part := range path {
		pathSoFar = append(pathSoFar, part)
		group := FindGroupByPathSlice(cfg, pathSoFar)
		if group != nil && group.Color != nil && *group.Color != "" {
			effectiveColor = *group.Color
		}
	}

	return effectiveColor
}
