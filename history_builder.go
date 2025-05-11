package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"strings"
	"time"
)

func buildHistoryItems(historyData HistoryData, config Config, favoritesData FavoritesData) []list.Item {
	var items []list.Item

	for i := len(historyData.Entries) - 1; i >= 0; i-- {
		entry := historyData.Entries[i]

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

		timeStr := entry.Timestamp.Format(time.RFC822)
		timeLine := fmt.Sprintf("🕓 %s", timeStr)

		var descLines []string
		descLines = append(descLines, connDetails)
		if pathLine != "" {
			descLines = append(descLines, pathLine)
		}
		descLines = append(descLines, timeLine)

		description := strings.Join(descLines, "\n")

		serverColor := entry.Server.Color
		if serverColor == "" {
			serverColor = "#FFFFFF"
		}

		favoriteStatus := false
		if config.FavoritesEnabled {
			favoriteStatus = isServerFavorited(&entry.Server, favoritesData)
		}

		title := fmt.Sprintf("💻 %s", entry.Server.Name)
		if favoriteStatus {
			title = fmt.Sprintf("⭐ %s", entry.Server.Name)
		}

		items = append(items, item{
			title:          title,
			description:    description,
			isGroup:        false,
			path:           strings.Join(entry.Path, "/"),
			color:          serverColor,
			isHistory:      true,
			historyEntry:   &entry,
			isMultiline:    true,
			pathEntries:    entry.Path,
			pathColors:     pathColors,
			favoriteStatus: favoriteStatus,
		})
	}

	return items
}
