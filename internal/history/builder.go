package history

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/gaetanlhf/sasha/internal/config"
)

func BuildItems(
	historyData Data,
	cfg config.Config,
	favoritesData interface{},
	getPathColors func(config.Config, []string) []string,
	isServerFavorited func(*config.Server, interface{}) bool,
) []list.Item {
	var items []list.Item

	for i := len(historyData.Entries) - 1; i >= 0; i-- {
		entry := historyData.Entries[i]
		resolved := ResolveEntry(entry, &cfg)
		if resolved == nil || resolved.Server == nil {
			continue
		}

		server := resolved.Server
		path := resolved.Path

		connDetails := server.Host
		if server.User != nil && *server.User != "" {
			connDetails = fmt.Sprintf("%s@%s", *server.User, connDetails)
		}
		if server.Port != nil && *server.Port != 0 && *server.Port != 22 {
			connDetails = fmt.Sprintf("%s:%d", connDetails, *server.Port)
		}

		pathLine := ""
		if len(path) > 0 {
			pathStr := strings.Join(path, " > ")
			pathLine = fmt.Sprintf("📁 %s", pathStr)
		}

		pathColors := getPathColors(cfg, path)

		timeStr := entry.Timestamp.Format(time.RFC822)
		timeLine := fmt.Sprintf("🕒 %s", timeStr)

		var descLines []string
		descLines = append(descLines, connDetails)
		if pathLine != "" {
			descLines = append(descLines, pathLine)
		}
		descLines = append(descLines, timeLine)

		description := strings.Join(descLines, "\n")

		serverColor := "#FFFFFF"
		if server.Color != nil && *server.Color != "" {
			serverColor = *server.Color
		}

		favoriteStatus := false
		if favoritesData != nil {
			favoriteStatus = isServerFavorited(server, favoritesData)
		}

		title := fmt.Sprintf("💻 %s", server.Name)
		if favoriteStatus {
			title = fmt.Sprintf("⭐ %s", server.Name)
		}

		items = append(items, HistoryItem{
			title:          title,
			description:    description,
			path:           strings.Join(path, "/"),
			color:          serverColor,
			server:         server,
			pathEntries:    path,
			pathColors:     pathColors,
			favoriteStatus: favoriteStatus,
		})
	}

	return items
}

type HistoryItem struct {
	title          string
	description    string
	path           string
	color          string
	server         *config.Server
	pathEntries    []string
	pathColors     []string
	favoriteStatus bool
}

func (i HistoryItem) Title() string             { return i.title }
func (i HistoryItem) Description() string       { return i.description }
func (i HistoryItem) FilterValue() string       { return i.title }
func (i HistoryItem) GetServer() *config.Server { return i.server }
func (i HistoryItem) GetPath() string           { return i.path }
func (i HistoryItem) GetColor() string          { return i.color }
func (i HistoryItem) GetPathEntries() []string  { return i.pathEntries }
func (i HistoryItem) GetPathColors() []string   { return i.pathColors }
func (i HistoryItem) IsFavorite() bool          { return i.favoriteStatus }
func (i HistoryItem) IsMultiline() bool         { return true }
