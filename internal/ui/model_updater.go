package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/gaetanlhf/SaSHa/internal/config"
	"github.com/gaetanlhf/SaSHa/internal/favorites"
	"github.com/gaetanlhf/SaSHa/internal/history"
	"github.com/gaetanlhf/SaSHa/internal/utils"
)

func (m *Model) updateColorBasedOnCurrentPath() {
	if m.InHistoryView {
		m.CurrentColor = historyColor
		initStyles(m.CurrentColor)

		m.List.Styles.PaginationStyle = titleStyle
		m.List.Styles.HelpStyle = helpStyle

		m.List.FilterInput.PromptStyle = filterPromptStyle
		m.List.FilterInput.TextStyle = filterTextStyle
		m.List.FilterInput.Cursor.Style = filterCursorStyle

		newDelegate := NewColoredDelegate()
		newDelegate.CurrentColor = "#FFFFFF"
		newDelegate.InHistoryView = true
		newDelegate.InFavoritesView = false
		m.List.SetDelegate(newDelegate)
		return
	}

	if m.InFavoritesView {
		m.CurrentColor = favoritesColor
		initStyles(m.CurrentColor)

		m.List.Styles.PaginationStyle = titleStyle
		m.List.Styles.HelpStyle = helpStyle

		m.List.FilterInput.PromptStyle = filterPromptStyle
		m.List.FilterInput.TextStyle = filterTextStyle
		m.List.FilterInput.Cursor.Style = filterCursorStyle

		newDelegate := NewColoredDelegate()
		newDelegate.CurrentColor = "#FFFFFF"
		newDelegate.InHistoryView = false
		newDelegate.InFavoritesView = true
		m.List.SetDelegate(newDelegate)
		return
	}

	colorToUse := utils.ResolveEffectiveColor(&m.Config, m.CurrentPath)

	if colorToUse != m.CurrentColor {
		m.CurrentColor = colorToUse
		initStyles(m.CurrentColor)

		m.List.Styles.PaginationStyle = titleStyle
		m.List.Styles.HelpStyle = helpStyle

		m.List.FilterInput.PromptStyle = filterPromptStyle
		m.List.FilterInput.TextStyle = filterTextStyle
		m.List.FilterInput.Cursor.Style = filterCursorStyle

		newDelegate := NewColoredDelegate()
		newDelegate.CurrentColor = colorToUse
		newDelegate.InHistoryView = false
		newDelegate.InFavoritesView = false
		m.List.SetDelegate(newDelegate)
	}
}

func (m *Model) updateListItems() {
	var items []list.Item

	if m.InHistoryView {
		if m.Config.Features.HistorySize == 0 {
			m.InHistoryView = false
			m.updateColorBasedOnCurrentPath()
			m.updateListItems()
			return
		}

		filteredHistory := history.FilterByExistingServers(m.HistoryData, &m.Config)
		items = buildHistoryItems(filteredHistory, m.Config, m.FavoritesData)
		m.List.SetItems(items)

		m.adjustListHeight()
		return
	}

	if m.InFavoritesView {
		if !m.Config.Features.FavoritesEnabled {
			m.InFavoritesView = false
			m.updateColorBasedOnCurrentPath()
			m.updateListItems()
			return
		}

		filteredFavorites := favorites.FilterByExistingServers(m.FavoritesData, &m.Config)
		items = buildFavoritesItems(filteredFavorites, m.Config)
		m.List.SetItems(items)

		m.adjustListHeight()
		return
	}

	if len(m.CurrentPath) == 0 {
		items = buildGroupItems(m.Config.Inventory.Groups, []string{}, &m.Config)

		for _, server := range m.Config.Inventory.Hosts {
			if server.Group == "" {
				desc := server.Host
				if server.User != nil && *server.User != "" {
					desc = fmt.Sprintf("%s@%s", *server.User, desc)
				}
				if server.Port != nil && *server.Port != 0 && *server.Port != 22 {
					desc = fmt.Sprintf("%s:%d", desc, *server.Port)
				}

				serverColor := utils.ResolveEffectiveColor(&m.Config, []string{})
				if server.Color != nil && *server.Color != "" {
					serverColor = *server.Color
				}

				favoriteStatus := false
				if m.Config.Features.FavoritesEnabled {
					favoriteStatus = favorites.IsServerFavorited(server, []string{}, m.FavoritesData)
				}

				title := fmt.Sprintf("💻 %s", server.Name)
				if favoriteStatus {
					title = fmt.Sprintf("⭐ %s", server.Name)
				}

				items = append(items, Item{
					Title:          title,
					Description:    desc,
					IsGroup:        false,
					Color:          serverColor,
					IsHistory:      false,
					FavoriteStatus: favoriteStatus,
					IsMultiline:    false,
				})
			}
		}
	} else {
		currentGroup := utils.FindGroupByPathSlice(&m.Config, m.CurrentPath)

		if currentGroup != nil {
			pathPrefix := append([]string{}, m.CurrentPath...)
			for _, group := range currentGroup.Groups {
				path := append([]string{}, pathPrefix...)
				path = append(path, group.Name)
				pathStr := strings.Join(path, "/")

				groupColor := utils.ResolveEffectiveColor(&m.Config, path)

				descParts := []string{}

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

			for _, server := range currentGroup.Hosts {
				desc := server.Host
				if server.User != nil && *server.User != "" {
					desc = fmt.Sprintf("%s@%s", *server.User, desc)
				}
				if server.Port != nil && *server.Port != 0 && *server.Port != 22 {
					desc = fmt.Sprintf("%s:%d", desc, *server.Port)
				}

				serverColor := utils.ResolveEffectiveColor(&m.Config, m.CurrentPath)
				if server.Color != nil && *server.Color != "" {
					serverColor = *server.Color
				}

				favoriteStatus := false
				if m.Config.Features.FavoritesEnabled {
					favoriteStatus = favorites.IsServerFavorited(server, m.CurrentPath, m.FavoritesData)
				}

				title := fmt.Sprintf("💻 %s", server.Name)
				if favoriteStatus {
					title = fmt.Sprintf("⭐ %s", server.Name)
				}

				items = append(items, Item{
					Title:          title,
					Description:    desc,
					IsGroup:        false,
					Color:          serverColor,
					IsHistory:      false,
					FavoriteStatus: favoriteStatus,
					IsMultiline:    false,
				})
			}
		}
	}

	m.List.SetItems(items)
	m.adjustListHeight()
}

func (m *Model) toggleHistoryView() {
	if m.Config.Features.HistorySize == 0 {
		return
	}

	if m.InFavoritesView {
		m.InFavoritesView = false
	}

	m.InHistoryView = !m.InHistoryView

	if m.InHistoryView {
		historyPath, _ := utils.GetHistoryFilePath()
		historyData, _ := history.Load(historyPath)
		m.HistoryData = historyData
	}

	m.updateColorBasedOnCurrentPath()
	m.updateListItems()
	m.List.Select(0)
}

func (m *Model) toggleFavoritesView() {
	if !m.Config.Features.FavoritesEnabled {
		return
	}

	if m.InHistoryView {
		m.InHistoryView = false
	}

	m.InFavoritesView = !m.InFavoritesView

	if m.InFavoritesView {
		favoritesPath, _ := utils.GetFavoritesFilePath()
		favoritesData, _ := favorites.Load(favoritesPath)
		m.FavoritesData = favoritesData
	}

	m.updateColorBasedOnCurrentPath()
	m.updateListItems()
	m.List.Select(0)
}

func (m *Model) toggleCurrentServerFavorite() {
	if !m.Config.Features.FavoritesEnabled {
		return
	}

	favoritesPath, _ := utils.GetFavoritesFilePath()

	if i, ok := m.List.SelectedItem().(Item); ok {
		if !i.IsGroup {
			var server *config.Server
			var path []string
			currentIndex := m.List.Index()

			if m.InHistoryView {
				if serverGetter, ok := m.List.SelectedItem().(interface{ GetServer() *config.Server }); ok {
					server = serverGetter.GetServer()
					if pathGetter, ok := m.List.SelectedItem().(interface{ GetPathEntries() []string }); ok {
						path = pathGetter.GetPathEntries()

						if m.StartInGroup && m.Config.Inventory != nil && len(m.Config.Inventory.Groups) == 1 {
							rootGroupName := m.Config.Inventory.Groups[0].Name
							if len(path) == 0 || path[0] != rootGroupName {
								newPath := append([]string{rootGroupName}, path...)
								path = newPath
							}
						}
					}
				}
			} else if m.InFavoritesView {
				if serverGetter, ok := m.List.SelectedItem().(interface{ GetServer() *config.Server }); ok {
					server = serverGetter.GetServer()
					if pathGetter, ok := m.List.SelectedItem().(interface{ GetPathEntries() []string }); ok {
						path = pathGetter.GetPathEntries()

						if m.StartInGroup && m.Config.Inventory != nil && len(m.Config.Inventory.Groups) == 1 {
							rootGroupName := m.Config.Inventory.Groups[0].Name
							if len(path) == 0 || path[0] != rootGroupName {
								newPath := append([]string{rootGroupName}, path...)
								path = newPath
							}
						}
					}
				}
			} else {
				serverName := strings.TrimPrefix(i.Title, "💻 ")
				serverName = strings.TrimPrefix(serverName, "⭐ ")
				server = m.findServer(serverName)
				path = m.CurrentPath
			}

			if server != nil {
				isFavorited := favorites.IsServerFavorited(server, path, m.FavoritesData)

				favorites.Add(favoritesPath, server, path, &m.Config)

				favoritesData, _ := favorites.Load(favoritesPath)
				m.FavoritesData = favoritesData

				if m.InFavoritesView && isFavorited {
					m.updateListItems()

					if currentIndex > 0 && currentIndex >= len(m.List.Items()) {
						m.List.Select(currentIndex - 1)
					} else {
						m.List.Select(currentIndex)
					}
				} else {
					m.updateListItems()
				}
			}
		}
	}
}

func (m *Model) findServer(serverName string) *config.Server {
	if len(m.CurrentPath) > 0 {
		currentGroup := utils.FindGroupByPathSlice(&m.Config, m.CurrentPath)
		if currentGroup != nil {
			for _, server := range currentGroup.Hosts {
				if server.Name == serverName {
					return server
				}
			}
		}
		return nil
	}

	if m.Config.Inventory != nil {
		for _, server := range m.Config.Inventory.Hosts {
			if server.Name == serverName {
				return server
			}
		}
	}

	return nil
}

func buildHistoryItems(historyData history.Data, cfg config.Config, favoritesData favorites.Data) []list.Item {
	return history.BuildItems(historyData, cfg, favoritesData, utils.GetPathColors, isServerFavoritedWrapper)
}

func buildFavoritesItems(favoritesData favorites.Data, cfg config.Config) []list.Item {
	return favorites.BuildItems(favoritesData, cfg, utils.GetPathColors)
}

func isServerFavoritedWrapper(server *config.Server, data interface{}) bool {
	if favData, ok := data.(favorites.Data); ok {
		return favorites.IsServerFavorited(server, []string{}, favData)
	}
	return false
}
