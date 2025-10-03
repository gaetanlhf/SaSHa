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

		connDetails := entry.Server.Host
		if entry.Server.User != nil && *entry.Server.User != "" {
			connDetails = fmt.Sprintf("%s@%s", *entry.Server.User, connDetails)
		}
		if entry.Server.Port != nil && *entry.Server.Port != 0 && *entry.Server.Port != 22 {
			connDetails = fmt.Sprintf("%s:%d", connDetails, *entry.Server.Port)
		}

		pathLine := ""
		if len(entry.Path) > 0 {
			pathStr := strings.Join(entry.Path, " > ")
			pathLine = fmt.Sprintf("📁 %s", pathStr)
		}

		pathColors := getPathColors(cfg, entry.Path)

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
		if entry.Server.Color != nil && *entry.Server.Color != "" {
			serverColor = *entry.Server.Color
		}

		favoriteStatus := false
		if favoritesData != nil {
			favoriteStatus = isServerFavorited(&entry.Server, favoritesData)
		}

		title := fmt.Sprintf("💻 %s", entry.Server.Name)
		if favoriteStatus {
			title = fmt.Sprintf("⭐ %s", entry.Server.Name)
		}

		items = append(items, HistoryItem{
			title:          title,
			description:    description,
			path:           strings.Join(entry.Path, "/"),
			color:          serverColor,
			historyEntry:   &entry,
			pathEntries:    entry.Path,
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
	historyEntry   *Entry
	pathEntries    []string
	pathColors     []string
	favoriteStatus bool
}

func (i HistoryItem) Title() string            { return i.title }
func (i HistoryItem) Description() string      { return i.description }
func (i HistoryItem) FilterValue() string      { return i.title }
func (i HistoryItem) GetHistoryEntry() *Entry  { return i.historyEntry }
func (i HistoryItem) GetPath() string          { return i.path }
func (i HistoryItem) GetColor() string         { return i.color }
func (i HistoryItem) GetPathEntries() []string { return i.pathEntries }
func (i HistoryItem) GetPathColors() []string  { return i.pathColors }
func (i HistoryItem) IsFavorite() bool         { return i.favoriteStatus }
func (i HistoryItem) IsMultiline() bool        { return true }
