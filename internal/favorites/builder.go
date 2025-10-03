package favorites

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/gaetanlhf/sasha/internal/config"
)

func BuildItems(
	favoritesData Data,
	cfg config.Config,
	getPathColors func(config.Config, []string) []string,
) []list.Item {
	var items []list.Item

	for _, entry := range favoritesData.Entries {
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

		var descLines []string
		descLines = append(descLines, connDetails)
		if pathLine != "" {
			descLines = append(descLines, pathLine)
		}

		description := strings.Join(descLines, "\n")

		serverColor := "#FFFFFF"
		if entry.Server.Color != nil && *entry.Server.Color != "" {
			serverColor = *entry.Server.Color
		}

		items = append(items, FavoriteItem{
			title:         fmt.Sprintf("💻 %s", entry.Server.Name),
			description:   description,
			path:          strings.Join(entry.Path, "/"),
			color:         serverColor,
			favoriteEntry: &entry,
			pathEntries:   entry.Path,
			pathColors:    pathColors,
		})
	}

	return items
}

type FavoriteItem struct {
	title         string
	description   string
	path          string
	color         string
	favoriteEntry *Entry
	pathEntries   []string
	pathColors    []string
}

func (i FavoriteItem) Title() string            { return i.title }
func (i FavoriteItem) Description() string      { return i.description }
func (i FavoriteItem) FilterValue() string      { return i.title }
func (i FavoriteItem) GetFavoriteEntry() *Entry { return i.favoriteEntry }
func (i FavoriteItem) GetPath() string          { return i.path }
func (i FavoriteItem) GetColor() string         { return i.color }
func (i FavoriteItem) GetPathEntries() []string { return i.pathEntries }
func (i FavoriteItem) GetPathColors() []string  { return i.pathColors }
func (i FavoriteItem) IsMultiline() bool        { return true }
