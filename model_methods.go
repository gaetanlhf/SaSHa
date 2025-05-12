package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.History):
			m.toggleHistoryView()
			return m, nil

		case key.Matches(msg, m.keys.Favorites):
			m.toggleFavoritesView()
			return m, nil

		case key.Matches(msg, m.keys.Favorite):
			m.toggleCurrentServerFavorite()
			return m, nil

		case key.Matches(msg, m.keys.Back):
			if m.inHistoryView || m.inFavoritesView {
				m.inHistoryView = false
				m.inFavoritesView = false
				m.updateColorBasedOnCurrentPath()
				m.updateListItems()
				m.list.Select(0)
				if m.list.FilterState() != list.Unfiltered {
					m.list.ResetFilter()
				}
				return m, nil
			} else if len(m.currentPath) > 0 {
				m.currentPath = m.currentPath[:len(m.currentPath)-1]
				m.breadcrumbColors = m.breadcrumbColors[:len(m.breadcrumbColors)-1]
				m.updateColorBasedOnCurrentPath()
				m.updateListItems()
				m.list.Select(0)
				if m.list.FilterState() != list.Unfiltered {
					m.list.ResetFilter()
				}
				return m, nil
			}

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			m.adjustListHeight()
			return m, nil
		}

		if key.Matches(msg, m.keys.Enter) {
			if i, ok := m.list.SelectedItem().(item); ok {
				if m.inHistoryView {
					if i.historyEntry != nil {
						server := &i.historyEntry.Server
						serverPath := i.historyEntry.Path

						currentGroup := findGroupByPathSlice(&m.config, serverPath)
						m.sshCommand = buildSSHCommand(server, currentGroup)

						addToHistory(server, serverPath, &m.config)

						m.quitting = true
						return m, tea.Quit
					}
				} else if m.inFavoritesView {
					if i.favoriteEntry != nil {
						server := &i.favoriteEntry.Server
						serverPath := i.favoriteEntry.Path

						currentGroup := findGroupByPathSlice(&m.config, serverPath)
						m.sshCommand = buildSSHCommand(server, currentGroup)

						addToHistory(server, serverPath, &m.config)

						m.quitting = true
						return m, tea.Quit
					}
				} else if i.isGroup {
					groupName := strings.TrimPrefix(i.title, "📁 ")
					if i.path != "" {
						parts := strings.Split(i.path, "/")
						groupName = parts[len(parts)-1]
					}
					m.currentPath = append(m.currentPath, groupName)

					colorToStore := m.currentColor
					if i.color != "" {
						colorToStore = i.color
						m.currentColor = i.color
						initStyles(i.color)
					} else {
						m.updateColorBasedOnCurrentPath()
					}
					m.breadcrumbColors = append(m.breadcrumbColors, colorToStore)

					m.updateListItems()
					m.list.Select(0)

					if m.list.FilterState() != list.Unfiltered {
						m.list.ResetFilter()
					}
				} else {
					serverName := strings.TrimPrefix(i.title, "💻 ")
					serverName = strings.TrimPrefix(serverName, "⭐ ")
					server := m.findServer(serverName)
					if server != nil {
						addToHistory(server, m.currentPath, &m.config)

						m.sshCommand = buildSSHCommand(server, findGroupByPathSlice(&m.config, m.currentPath))
						m.quitting = true
						return m, tea.Quit
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.width = msg.Width - h
		m.height = msg.Height - v
		m.help.Width = m.width

		m.adjustListHeight()
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *model) adjustListHeight() {
	if m.height == 0 {
		return
	}

	top, right, bottom, left := 2, 2, 1, 2
	helpHeight := 1

	if m.help.ShowAll {
		helpText := m.help.View(m.keys)
		helpHeight = strings.Count(helpText, "\n") + 1
	}

	m.list.SetSize(
		m.width-left-right,
		m.height-top-bottom-helpHeight,
	)
}

func (m model) View() string {
	if m.list.Width() == 0 {
		return appStyle.Render("Loading...")
	}

	var content strings.Builder

	homeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	historyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(historyColor))
	favoritesStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(favoritesColor))
	separatorStyle := helpStyle

	maxBreadcrumbWidth := m.width - 4

	if m.inHistoryView {
		breadcrumb := historyStyle.Render("🕒 History")
		breadcrumb = truncateBreadcrumb(breadcrumb, maxBreadcrumbWidth)
		content.WriteString(breadcrumb + "\n\n")
	} else if m.inFavoritesView {
		breadcrumb := favoritesStyle.Render("⭐ Favorites")
		breadcrumb = truncateBreadcrumb(breadcrumb, maxBreadcrumbWidth)
		content.WriteString(breadcrumb + "\n\n")
	} else {
		breadcrumb := homeStyle.Render("🏠 Home")
		if len(m.currentPath) > 0 {
			for i, path := range m.currentPath {
				separator := separatorStyle.Render(" > ")
				if i == 0 {
					breadcrumb += separator
				}

				pathColor := "#FFFFFF"
				if i < len(m.breadcrumbColors) {
					pathColor = m.breadcrumbColors[i]
				}

				pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(pathColor))
				breadcrumb += pathStyle.Render(path)

				if i < len(m.currentPath)-1 {
					breadcrumb += separator
				}
			}
		}
		breadcrumb = truncateBreadcrumb(breadcrumb, maxBreadcrumbWidth)
		content.WriteString(breadcrumb + "\n\n")
	}

	isFiltering := m.list.FilterState() == list.Filtering

	if !isFiltering {
		categoryColor := "#FFFFFF"
		categoryName := "Home"

		if m.inHistoryView {
			categoryColor = historyColor
			categoryName = "History"
		} else if m.inFavoritesView {
			categoryColor = favoritesColor
			categoryName = "Favorites"
		} else if len(m.currentPath) > 0 {
			currentGroup := findGroupByPathSlice(&m.config, m.currentPath)
			if currentGroup != nil {
				categoryColor = m.currentColor
				categoryName = currentGroup.Name
			}
		}

		textColor := getContrastColor(categoryColor)

		categoryStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(textColor)).
			Background(lipgloss.Color(categoryColor)).
			Bold(true).
			Padding(0, 1).
			MarginLeft(2)

		categoryLabel := categoryStyle.Render(categoryName)
		content.WriteString(categoryLabel + "\n")
	}

	content.WriteString(m.list.View())

	if isFiltering {
		content.WriteString("\n")
	}

	helpView := m.help.View(m.keys)

	versionDisplay := version
	if versionDisplay == "" {
		versionDisplay = "dev"
	}
	versionStr := fmt.Sprintf("SaSHa %s", versionDisplay)
	versionStyle := helpStyle

	if m.help.ShowAll {
		lines := strings.Split(helpView, "\n")
		if len(lines) > 0 {
			lastLine := lines[len(lines)-1]

			paddingWidth := m.width - lipgloss.Width(lastLine) - lipgloss.Width(versionStr)
			if paddingWidth < 0 {
				paddingWidth = 0
			}
			paddingStr := strings.Repeat(" ", paddingWidth)

			lines[len(lines)-1] = lastLine + paddingStr + versionStyle.Render(versionStr)
			helpView = strings.Join(lines, "\n")
		}
	} else {
		paddingWidth := m.width - lipgloss.Width(helpView) - lipgloss.Width(versionStr)
		if paddingWidth < 0 {
			paddingWidth = 0
		}
		paddingStr := strings.Repeat(" ", paddingWidth)
		helpView = helpView + paddingStr + versionStyle.Render(versionStr)
	}

	content.WriteString("\n" + helpView)

	contentStr := content.String()

	return appStyle.Render(contentStr)
}
