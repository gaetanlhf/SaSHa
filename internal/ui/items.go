package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/gaetanlhf/SaSHa/internal/config"
	"github.com/gaetanlhf/SaSHa/internal/utils"
)

func buildGroupItems(groups []*config.Group, pathPrefix []string, cfg *config.Config) []list.Item {
	var items []list.Item

	for _, group := range groups {
		path := append([]string{}, pathPrefix...)
		path = append(path, group.Name)
		pathStr := strings.Join(path, "/")

		groupColor := utils.ResolveEffectiveColor(cfg, path)

		var descParts []string

		if len(group.Hosts) > 0 {
			hostWord := "host"
			if len(group.Hosts) > 1 {
				hostWord = "hosts"
			}
			descParts = append(descParts, fmt.Sprintf("%d %s", len(group.Hosts), hostWord))
		}

		if len(group.Groups) > 0 {
			subgroupWord := "subgroup"
			if len(group.Groups) > 1 {
				subgroupWord = "subgroups"
			}
			descParts = append(descParts, fmt.Sprintf("%d %s", len(group.Groups), subgroupWord))
		}

		description := "Empty group"
		if len(descParts) > 0 {
			description = fmt.Sprintf("Group with %s", strings.Join(descParts, ", "))
		}

		items = append(items, Item{
			Title:       fmt.Sprintf("📁 %s", group.Name),
			Description: description,
			IsGroup:     true,
			Path:        pathStr,
			Color:       groupColor,
			IsHistory:   false,
			IsFavorite:  false,
			IsMultiline: false,
		})
	}

	return items
}

func buildErrorItems(errors []string) []list.Item {
	var items []list.Item

	for _, err := range errors {
		errorInfo := utils.ClassifyError(err)

		items = append(items, Item{
			Title:       fmt.Sprintf("%s %s", errorInfo.Emoji, errorInfo.Message),
			Description: "",
			IsGroup:     false,
			Path:        "",
			Color:       errorInfo.Color,
			IsMultiline: false,
			IsError:     true,
		})
	}

	return items
}
