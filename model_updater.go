package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"strings"
)

func (m *model) updateColorBasedOnCurrentPath() {
	if m.inHistoryView {
		m.currentColor = historyColor
		initStyles(m.currentColor)

		m.list.Styles.PaginationStyle = titleStyle
		m.list.Styles.HelpStyle = helpStyle

		m.list.FilterInput.PromptStyle = filterPromptStyle
		m.list.FilterInput.TextStyle = filterTextStyle
		m.list.FilterInput.Cursor.Style = filterCursorStyle

		newDelegate := NewColoredDelegate()
		newDelegate.currentColor = "#FFFFFF"
		newDelegate.inHistoryView = true
		newDelegate.inFavoritesView = false
		m.list.SetDelegate(newDelegate)
		return
	}

	if m.inFavoritesView {
		m.currentColor = favoritesColor
		initStyles(m.currentColor)

		m.list.Styles.PaginationStyle = titleStyle
		m.list.Styles.HelpStyle = helpStyle

		m.list.FilterInput.PromptStyle = filterPromptStyle
		m.list.FilterInput.TextStyle = filterTextStyle
		m.list.FilterInput.Cursor.Style = filterCursorStyle

		newDelegate := NewColoredDelegate()
		newDelegate.currentColor = "#FFFFFF"
		newDelegate.inHistoryView = false
		newDelegate.inFavoritesView = true
		m.list.SetDelegate(newDelegate)
		return
	}

	colorToUse := "#FFFFFF"

	if len(m.currentPath) > 0 {
		pathSoFar := []string{}

		for _, part := range m.currentPath {
			pathSoFar = append(pathSoFar, part)
			tempGroup := findGroupByPathSlice(&m.config, pathSoFar)

			if tempGroup != nil && tempGroup.Color != "" {
				colorToUse = tempGroup.Color
			}
		}
	}

	if colorToUse != m.currentColor {
		m.currentColor = colorToUse
		initStyles(m.currentColor)

		m.list.Styles.PaginationStyle = titleStyle
		m.list.Styles.HelpStyle = helpStyle

		m.list.FilterInput.PromptStyle = filterPromptStyle
		m.list.FilterInput.TextStyle = filterTextStyle
		m.list.FilterInput.Cursor.Style = filterCursorStyle

		newDelegate := NewColoredDelegate()
		newDelegate.currentColor = colorToUse
		newDelegate.inHistoryView = false
		newDelegate.inFavoritesView = false
		m.list.SetDelegate(newDelegate)
	}
}

func (m *model) updateListItems() {
	var items []list.Item

	if m.inHistoryView {
		if m.config.HistorySize == 0 {
			m.inHistoryView = false
			m.updateColorBasedOnCurrentPath()
			m.updateListItems()
			return
		}

		filteredHistory := filterHistoryByExistingServers(m.historyData, &m.config)
		items = buildHistoryItems(filteredHistory, m.config, m.favoritesData)
		m.list.SetItems(items)

		m.adjustListHeight()
		return
	}

	if m.inFavoritesView {
		if !m.config.FavoritesEnabled {
			m.inFavoritesView = false
			m.updateColorBasedOnCurrentPath()
			m.updateListItems()
			return
		}

		filteredFavorites := filterFavoritesByExistingServers(m.favoritesData, &m.config)
		items = buildFavoritesItems(filteredFavorites, m.config)
		m.list.SetItems(items)

		m.adjustListHeight()
		return
	}

	if len(m.currentPath) == 0 {
		items = buildGroupItems(m.config.Groups, []string{}, m.currentColor)

		for _, server := range m.config.Hosts {
			if server.Group == "" {
				desc := server.Host
				if server.User != "" {
					desc = fmt.Sprintf("%s@%s", server.User, desc)
				}
				if server.Port != 0 && server.Port != 22 {
					desc = fmt.Sprintf("%s:%d", desc, server.Port)
				}

				serverColor := server.Color
				if serverColor == "" {
					serverColor = m.currentColor
				}

				favoriteStatus := false
				if m.config.FavoritesEnabled {
					favoriteStatus = isServerFavorited(server, m.favoritesData)
				}

				title := fmt.Sprintf("💻 %s", server.Name)
				if favoriteStatus {
					title = fmt.Sprintf("⭐ %s", server.Name)
				}

				items = append(items, item{
					title:          title,
					description:    desc,
					isGroup:        false,
					color:          serverColor,
					isHistory:      false,
					favoriteStatus: favoriteStatus,
					isMultiline:    false,
				})
			}
		}
	} else {
		currentGroup := findGroupByPathSlice(&m.config, m.currentPath)

		if currentGroup != nil {
			pathPrefix := append([]string{}, m.currentPath...)
			for _, group := range currentGroup.Groups {
				path := append([]string{}, pathPrefix...)
				path = append(path, group.Name)
				pathStr := strings.Join(path, "/")

				groupColor := group.Color
				if groupColor == "" {
					groupColor = m.currentColor
				}

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

			for _, server := range currentGroup.Hosts {
				desc := server.Host
				if server.User != "" {
					desc = fmt.Sprintf("%s@%s", server.User, desc)
				}
				if server.Port != 0 && server.Port != 22 {
					desc = fmt.Sprintf("%s:%d", desc, server.Port)
				}

				serverColor := server.Color
				if serverColor == "" {
					serverColor = m.currentColor
				}

				favoriteStatus := false
				if m.config.FavoritesEnabled {
					favoriteStatus = isServerFavorited(server, m.favoritesData)
				}

				title := fmt.Sprintf("💻 %s", server.Name)
				if favoriteStatus {
					title = fmt.Sprintf("⭐ %s", server.Name)
				}

				items = append(items, item{
					title:          title,
					description:    desc,
					isGroup:        false,
					color:          serverColor,
					isHistory:      false,
					favoriteStatus: favoriteStatus,
					isMultiline:    false,
				})
			}
		}
	}

	m.list.SetItems(items)
	m.adjustListHeight()
}

func (m *model) toggleHistoryView() {
	if m.config.HistorySize == 0 {
		return
	}

	if m.inFavoritesView {
		m.inFavoritesView = false
	}

	m.inHistoryView = !m.inHistoryView

	if m.inHistoryView {
		historyData, _ := loadHistory()
		m.historyData = historyData
	}

	m.updateColorBasedOnCurrentPath()
	m.updateListItems()
	m.list.Select(0)
}

func (m *model) toggleFavoritesView() {
	if !m.config.FavoritesEnabled {
		return
	}

	if m.inHistoryView {
		m.inHistoryView = false
	}

	m.inFavoritesView = !m.inFavoritesView

	if m.inFavoritesView {
		favoritesData, _ := loadFavorites()
		m.favoritesData = favoritesData
	}

	m.updateColorBasedOnCurrentPath()
	m.updateListItems()
	m.list.Select(0)
}

func (m *model) toggleCurrentServerFavorite() {
	if !m.config.FavoritesEnabled {
		return
	}

	if i, ok := m.list.SelectedItem().(item); ok {
		if !i.isGroup {
			var server *Server
			var path []string
			currentIndex := m.list.Index()

			if m.inHistoryView && i.historyEntry != nil {
				server = &i.historyEntry.Server
				path = i.historyEntry.Path
			} else if m.inFavoritesView && i.favoriteEntry != nil {
				server = &i.favoriteEntry.Server
				path = i.favoriteEntry.Path
			} else {
				serverName := strings.TrimPrefix(i.title, "💻 ")
				serverName = strings.TrimPrefix(serverName, "⭐ ")
				server = m.findServer(serverName)
				path = m.currentPath
			}

			if server != nil {
				isFavorited := isServerFavorited(server, m.favoritesData)

				addToFavorites(server, path, &m.config)

				favoritesData, _ := loadFavorites()
				m.favoritesData = favoritesData

				if m.inFavoritesView && isFavorited {
					m.updateListItems()

					if currentIndex > 0 && currentIndex >= len(m.list.Items()) {
						m.list.Select(currentIndex - 1)
					} else {
						m.list.Select(currentIndex)
					}
				} else {
					m.updateListItems()
				}
			}
		}
	}
}

func (m *model) findServer(serverName string) *Server {
	if len(m.currentPath) > 0 {
		currentGroup := findGroupByPathSlice(&m.config, m.currentPath)
		if currentGroup != nil {
			for _, server := range currentGroup.Hosts {
				if server.Name == serverName {
					return server
				}
			}
		}
		return nil
	}

	for _, server := range m.config.Hosts {
		if server.Name == serverName {
			return server
		}
	}

	return nil
}

func (m *model) findHistoryServer(serverName string) (*Server, []string) {
	for _, entry := range m.historyData.Entries {
		if entry.Server.Name == serverName {
			return &entry.Server, entry.Path
		}
	}
	return nil, nil
}

func (m *model) findFavoriteServer(serverName string) (*Server, []string) {
	for _, entry := range m.favoritesData.Entries {
		if entry.Server.Name == serverName {
			return &entry.Server, entry.Path
		}
	}
	return nil, nil
}
