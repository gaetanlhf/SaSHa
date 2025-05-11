package main

import (
	"strings"
)

func findGroupByPath(config *Config, path string) *Group {
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return nil
	}

	var currentGroup *Group
	for _, group := range config.Groups {
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

func findGroupByPathSlice(config *Config, path []string) *Group {
	if len(path) == 0 {
		return nil
	}

	var currentGroup *Group
	for _, group := range config.Groups {
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
