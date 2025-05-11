package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"strings"
)

func buildFavoritesItems(favoritesData FavoritesData, config Config) []list.Item {
	var items []list.Item

	for _, entry := range favoritesData.Entries {
		connDetails := entry.Server.Host
		if entry.Server.User != "" {
			connDetails = fmt.Sprintf("%s@%s", entry.Server.User, connDetails)
		}
		if entry.Server.Port != 0 && entry.Server.Port != 22 {
			connDetails = fmt.Sprintf("%s:%d", connDetails, entry.Server.Port)
		}

		pathLine := ""
		if len(entry.Path) > 0 {
			pathStr := strings.Join(entry.Path, " > ")
			pathLine = fmt.Sprintf("📁 %s", pathStr)
		}

		pathColors := getPathColors(config, entry.Path)

		extraInfo := ""
		if entry.Server.SSHBinary != "" && entry.Server.SSHBinary != "ssh" {
			extraInfo = fmt.Sprintf("SSH Client: %s", entry.Server.SSHBinary)
		}

		var descLines []string
		descLines = append(descLines, connDetails)
		if pathLine != "" {
			descLines = append(descLines, pathLine)
		}
		if extraInfo != "" {
			descLines = append(descLines, extraInfo)
		}

		description := strings.Join(descLines, "\n")

		serverColor := entry.Server.Color
		if serverColor == "" {
			serverColor = "#FFFFFF"
		}

		items = append(items, item{
			title:          fmt.Sprintf("💻 %s", entry.Server.Name),
			description:    description,
			isGroup:        false,
			path:           strings.Join(entry.Path, "/"),
			color:          serverColor,
			isFavorite:     true,
			favoriteEntry:  &entry,
			favoriteStatus: true,
			isMultiline:    true,
			pathEntries:    entry.Path,
			pathColors:     pathColors,
		})
	}

	return items
}
