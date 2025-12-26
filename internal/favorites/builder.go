package favorites

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/gaetanlhf/SaSHa/internal/config"
)

func BuildItems(
	favoritesData Data,
	cfg config.Config,
	getPathColors func(config.Config, []string) []string,
) []list.Item {
	var items []list.Item

	startInGroup := cfg.Inventory != nil && len(cfg.Inventory.Groups) == 1 && len(cfg.Inventory.Hosts) == 0

	for _, entry := range favoritesData.Entries {
		resolved := ResolveEntry(entry, &cfg)
		if resolved == nil || resolved.Server == nil {
			continue
		}

		server := resolved.Server
		path := resolved.Path
		pathColors := getPathColors(cfg, path)

		if startInGroup && len(path) > 0 && path[0] == cfg.Inventory.Groups[0].Name {
			path = path[1:]
			if len(pathColors) > 0 {
				pathColors = pathColors[1:]
			}
		}

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

		var descLines []string
		descLines = append(descLines, connDetails)
		if pathLine != "" {
			descLines = append(descLines, pathLine)
		}

		description := strings.Join(descLines, "\n")

		serverColor := "#FFFFFF"
		if server.Color != nil && *server.Color != "" {
			serverColor = *server.Color
		}

		items = append(items, FavoriteItem{
			title:       fmt.Sprintf("💻 %s", server.Name),
			description: description,
			path:        strings.Join(path, "/"),
			color:       serverColor,
			server:      server,
			pathEntries: path,
			pathColors:  pathColors,
		})
	}

	return items
}

func (i FavoriteItem) Title() string             { return i.title }
func (i FavoriteItem) Description() string       { return i.description }
func (i FavoriteItem) FilterValue() string       { return i.title }
func (i FavoriteItem) GetServer() *config.Server { return i.server }
func (i FavoriteItem) GetPath() string           { return i.path }
func (i FavoriteItem) GetColor() string          { return i.color }
func (i FavoriteItem) GetPathEntries() []string  { return i.pathEntries }
func (i FavoriteItem) GetPathColors() []string   { return i.pathColors }
func (i FavoriteItem) IsMultiline() bool         { return true }
