package ui

import (
	"fmt"
	"strings"

	"github.com/gaetanlhf/SaSHa/internal/config"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gaetanlhf/SaSHa/internal/history"
	"github.com/gaetanlhf/SaSHa/internal/ssh"
	"github.com/gaetanlhf/SaSHa/internal/utils"
)

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.InErrorView {
			switch {
			case key.Matches(msg, m.Keys.Quit):
				return m, tea.Quit
			case key.Matches(msg, m.Keys.Back):
				return m, tea.Quit
			case key.Matches(msg, m.Keys.Enter):
				m.InErrorView = false
				m.Keys = newKeyMap(m.Config.Features.HistorySize > 0, m.Config.Features.FavoritesEnabled, false)
				delegate := NewColoredDelegate()
				delegate.CurrentColor = "#FFFFFF"
				m.List.SetDelegate(delegate)
				m.updateColorBasedOnCurrentPath()
				m.updateListItems()
				m.List.Select(0)
				if m.StartInGroup && m.Config.Inventory != nil && len(m.Config.Inventory.Groups) == 1 && len(m.Config.Inventory.Hosts) == 0 {
					singleGroup := m.Config.Inventory.Groups[0]
					m.CurrentPath = []string{singleGroup.Name}
					if singleGroup.Color != nil && *singleGroup.Color != "" {
						m.CurrentColor = *singleGroup.Color
						initStyles(m.CurrentColor)
						delegate.CurrentColor = m.CurrentColor
						m.List.SetDelegate(delegate)
					}
					m.BreadcrumbColors = []string{m.CurrentColor}
					m.updateColorBasedOnCurrentPath()
					m.updateListItems()
				}
				return m, nil
			}
			return m, nil
		}

		if m.List.FilterState() == list.Filtering {
			m.List, cmd = m.List.Update(msg)
			return m, cmd
		}

		switch {
		case key.Matches(msg, m.Keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.Keys.History):
			m.toggleHistoryView()
			return m, nil

		case key.Matches(msg, m.Keys.Favorites):
			m.toggleFavoritesView()
			return m, nil

		case key.Matches(msg, m.Keys.Favorite):
			m.toggleCurrentServerFavorite()
			return m, nil

		case key.Matches(msg, m.Keys.Back):
			if m.InHistoryView || m.InFavoritesView {
				m.InHistoryView = false
				m.InFavoritesView = false
				m.updateColorBasedOnCurrentPath()
				m.updateListItems()
				m.List.Select(0)
				if m.List.FilterState() != list.Unfiltered {
					m.List.ResetFilter()
				}
				return m, nil
			} else if len(m.CurrentPath) > 0 {
				if m.StartInGroup && len(m.CurrentPath) == 1 {
					return m, tea.Quit
				}
				m.CurrentPath = m.CurrentPath[:len(m.CurrentPath)-1]
				m.BreadcrumbColors = m.BreadcrumbColors[:len(m.BreadcrumbColors)-1]

				var previousSelection int
				if len(m.SelectionStack) > 0 {
					previousSelection = m.SelectionStack[len(m.SelectionStack)-1]
					m.SelectionStack = m.SelectionStack[:len(m.SelectionStack)-1]
				}

				m.updateColorBasedOnCurrentPath()
				m.updateListItems()

				if previousSelection < len(m.List.Items()) {
					m.List.Select(previousSelection)
				} else {
					m.List.Select(0)
				}

				if m.List.FilterState() != list.Unfiltered {
					m.List.ResetFilter()
				}
				return m, nil
			}

		case key.Matches(msg, m.Keys.Help):
			m.Help.ShowAll = !m.Help.ShowAll
			m.adjustListHeight()
			return m, nil
		}

		if key.Matches(msg, m.Keys.Enter) {
			historyPath, _ := utils.GetHistoryFilePath()

			if m.InHistoryView {
				if he, ok := m.List.SelectedItem().(interface{ GetServer() *config.Server }); ok {
					server := he.GetServer()
					if server != nil {
						if pathGetter, ok := m.List.SelectedItem().(interface{ GetPathEntries() []string }); ok {
							serverPath := pathGetter.GetPathEntries()

							if m.StartInGroup && m.Config.Inventory != nil && len(m.Config.Inventory.Groups) == 1 {
								rootGroupName := m.Config.Inventory.Groups[0].Name
								if len(serverPath) == 0 || serverPath[0] != rootGroupName {
									newPath := append([]string{rootGroupName}, serverPath...)
									serverPath = newPath
								}
							}

							currentGroup := utils.FindGroupByPathSlice(&m.Config, serverPath)
							m.SSHCommand = ssh.BuildCommand(server, currentGroup)

							history.Add(historyPath, server, serverPath, &m.Config)

							m.Quitting = true
							return m, tea.Quit
						}
					}
				}
			} else if m.InFavoritesView {
				if fe, ok := m.List.SelectedItem().(interface{ GetServer() *config.Server }); ok {
					server := fe.GetServer()
					if server != nil {
						if pathGetter, ok := m.List.SelectedItem().(interface{ GetPathEntries() []string }); ok {
							serverPath := pathGetter.GetPathEntries()

							if m.StartInGroup && m.Config.Inventory != nil && len(m.Config.Inventory.Groups) == 1 {
								rootGroupName := m.Config.Inventory.Groups[0].Name
								if len(serverPath) == 0 || serverPath[0] != rootGroupName {
									newPath := append([]string{rootGroupName}, serverPath...)
									serverPath = newPath
								}
							}

							currentGroup := utils.FindGroupByPathSlice(&m.Config, serverPath)
							m.SSHCommand = ssh.BuildCommand(server, currentGroup)

							history.Add(historyPath, server, serverPath, &m.Config)

							m.Quitting = true
							return m, tea.Quit
						}
					}
				}
			} else if i, ok := m.List.SelectedItem().(Item); ok {
				if i.IsGroup {
					currentSelection := m.List.Index()
					m.SelectionStack = append(m.SelectionStack, currentSelection)

					groupName := strings.TrimPrefix(i.Title, "📁 ")
					if i.Path != "" {
						parts := strings.Split(i.Path, "/")
						groupName = parts[len(parts)-1]
					}
					m.CurrentPath = append(m.CurrentPath, groupName)

					colorToStore := m.CurrentColor
					if i.Color != "" {
						colorToStore = i.Color
						m.CurrentColor = i.Color
						initStyles(i.Color)
					} else {
						m.updateColorBasedOnCurrentPath()
					}
					m.BreadcrumbColors = append(m.BreadcrumbColors, colorToStore)

					m.updateListItems()
					m.List.Select(0)

					if m.List.FilterState() != list.Unfiltered {
						m.List.ResetFilter()
					}
				} else {
					serverName := strings.TrimPrefix(i.Title, "💻 ")
					serverName = strings.TrimPrefix(serverName, "⭐ ")
					server := m.findServer(serverName)
					if server != nil {
						history.Add(historyPath, server, m.CurrentPath, &m.Config)

						m.SSHCommand = ssh.BuildCommand(server, utils.FindGroupByPathSlice(&m.Config, m.CurrentPath))
						m.Quitting = true
						return m, tea.Quit
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.Width = msg.Width - h
		m.Height = msg.Height - v
		m.Help.Width = m.Width

		m.adjustListHeight()
	}

	if !m.InErrorView {
		m.List, cmd = m.List.Update(msg)
	}
	return m, cmd
}

func (m *Model) adjustListHeight() {
	if m.Height == 0 {
		return
	}

	top, right, bottom, left := 2, 2, 1, 2
	helpHeight := 1

	if m.Help.ShowAll && !m.InErrorView {
		helpText := m.Help.View(m.Keys)
		helpHeight = strings.Count(helpText, "\n") + 1
	}

	if m.InErrorView {
		helpHeight = 1
	}

	m.List.SetSize(
		m.Width-left-right,
		m.Height-top-bottom-helpHeight,
	)
}

func (m Model) View() string {
	if m.List.Width() == 0 {
		return appStyle.Render("Loading...")
	}

	var content strings.Builder

	homeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	historyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(historyColor))
	favoritesStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(favoritesColor))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	separatorStyle := helpStyle

	maxBreadcrumbWidth := m.Width - 4

	if m.InErrorView {
		errorTitle := "🚨 Error"
		if len(m.Config.ImportErrors) > 1 {
			errorTitle = "🚨 Errors"
		}
		breadcrumb := errorStyle.Render(errorTitle)
		breadcrumb = utils.TruncateBreadcrumb(breadcrumb, maxBreadcrumbWidth)
		content.WriteString(breadcrumb + "\n\n")
	} else if m.InHistoryView {
		breadcrumb := historyStyle.Render("🕘 History")
		breadcrumb = utils.TruncateBreadcrumb(breadcrumb, maxBreadcrumbWidth)
		content.WriteString(breadcrumb + "\n\n")
	} else if m.InFavoritesView {
		breadcrumb := favoritesStyle.Render("⭐ Favorites")
		breadcrumb = utils.TruncateBreadcrumb(breadcrumb, maxBreadcrumbWidth)
		content.WriteString(breadcrumb + "\n\n")
	} else {
		breadcrumb := homeStyle.Render("🏠 Home")
		if len(m.CurrentPath) > 0 {
			startIndex := 0
			if m.StartInGroup {
				startIndex = 1
			}

			for i := startIndex; i < len(m.CurrentPath); i++ {
				separator := separatorStyle.Render(" > ")
				if i == startIndex {
					breadcrumb += separator
				}

				pathColor := "#FFFFFF"
				if i < len(m.BreadcrumbColors) {
					pathColor = m.BreadcrumbColors[i]
				}

				pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(pathColor))
				breadcrumb += pathStyle.Render(m.CurrentPath[i])

				if i < len(m.CurrentPath)-1 {
					breadcrumb += separator
				}
			}
		}
		breadcrumb = utils.TruncateBreadcrumb(breadcrumb, maxBreadcrumbWidth)
		content.WriteString(breadcrumb + "\n\n")
	}

	isFiltering := m.List.FilterState() == list.Filtering

	if !isFiltering && !m.InErrorView {
		categoryColor := "#FFFFFF"
		categoryName := "Home"

		if m.InHistoryView {
			categoryColor = historyColor
			categoryName = "History"
		} else if m.InFavoritesView {
			categoryColor = favoritesColor
			categoryName = "Favorites"
		} else if len(m.CurrentPath) > 0 {
			currentGroup := utils.FindGroupByPathSlice(&m.Config, m.CurrentPath)
			if currentGroup != nil {
				if m.StartInGroup && len(m.CurrentPath) == 1 {
					categoryColor = "#FFFFFF"
					categoryName = "Home"
				} else {
					categoryColor = m.CurrentColor
					categoryName = currentGroup.Name
				}
			}
		}

		textColor := utils.GetContrastColor(categoryColor)

		categoryStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(textColor)).
			Background(lipgloss.Color(categoryColor)).
			Bold(true).
			Padding(0, 1).
			MarginLeft(2)

		categoryLabel := categoryStyle.Render(categoryName)
		content.WriteString(categoryLabel + "\n")
	} else if m.InErrorView {
		categoryName := "Error"
		if len(m.Config.ImportErrors) > 1 {
			categoryName = "Errors"
		}

		textColor := utils.GetContrastColor("#FF6B6B")

		categoryStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(textColor)).
			Background(lipgloss.Color("#FF6B6B")).
			Bold(true).
			Padding(0, 1).
			MarginLeft(2)

		categoryLabel := categoryStyle.Render(categoryName)
		content.WriteString(categoryLabel + "\n")
	}

	content.WriteString(m.List.View())

	if isFiltering {
		content.WriteString("\n")
	}

	helpView := m.Help.View(m.Keys)

	versionDisplay := version
	if versionDisplay == "" {
		versionDisplay = "dev"
	}
	versionStr := fmt.Sprintf("SaSHa %s", versionDisplay)
	versionStyle := helpStyle

	if m.Help.ShowAll && !m.InErrorView {
		lines := strings.Split(helpView, "\n")
		if len(lines) > 0 {
			lastLine := lines[len(lines)-1]

			paddingWidth := m.Width - lipgloss.Width(lastLine) - lipgloss.Width(versionStr)
			if paddingWidth < 0 {
				paddingWidth = 0
			}
			paddingStr := strings.Repeat(" ", paddingWidth)

			lines[len(lines)-1] = lastLine + paddingStr + versionStyle.Render(versionStr)
			helpView = strings.Join(lines, "\n")
		}
	} else {
		paddingWidth := m.Width - lipgloss.Width(helpView) - lipgloss.Width(versionStr)
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
