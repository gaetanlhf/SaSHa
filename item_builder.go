package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
)

type item struct {
	title          string
	description    string
	isGroup        bool
	path           string
	color          string
	isHistory      bool
	historyEntry   *HistoryEntry
	isFavorite     bool
	favoriteEntry  *FavoriteEntry
	favoriteStatus bool
	isMultiline    bool
	pathEntries    []string
	pathColors     []string
	isError        bool
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

func buildGroupItems(groups []*Group, pathPrefix []string, defaultColor string) []list.Item {
	var items []list.Item

	for _, group := range groups {
		path := append([]string{}, pathPrefix...)
		path = append(path, group.Name)
		pathStr := strings.Join(path, "/")

		groupColor := group.Color
		if groupColor == "" {
			groupColor = defaultColor
		}

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

		items = append(items, item{
			title:       fmt.Sprintf("📁 %s", group.Name),
			description: description,
			isGroup:     true,
			path:        pathStr,
			color:       groupColor,
			isHistory:   false,
			isFavorite:  false,
			isMultiline: false,
		})
	}

	return items
}

func buildErrorItems(errors []string) []list.Item {
	var items []list.Item

	for _, err := range errors {
		errorInfo := classifyError(err)

		items = append(items, item{
			title:       fmt.Sprintf("%s %s", errorInfo.Emoji, errorInfo.Message),
			description: "",
			isGroup:     false,
			path:        "",
			color:       errorInfo.Color,
			isMultiline: false,
			isError:     true,
		})
	}

	return items
}
