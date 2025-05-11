package main

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().
			Padding(1, 2, 1, 2).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#626262"))

	defaultColor      = "#FFFFFF"
	historyColor      = "#7D56F4"
	favoritesColor    = "#FFD700"
	titleStyle        lipgloss.Style
	itemStyle         lipgloss.Style
	selectedItemStyle lipgloss.Style
	helpStyle         lipgloss.Style
	breadcrumbStyle   lipgloss.Style

	filterTextStyle   lipgloss.Style
	filterCursorStyle lipgloss.Style
	filterPromptStyle lipgloss.Style
)

func initStyles(baseColor string) {
	if baseColor == "" {
		baseColor = defaultColor
	}

	titleStyle = lipgloss.NewStyle().MarginLeft(2).Bold(true).Foreground(lipgloss.Color(baseColor))
	itemStyle = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color(baseColor))
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	breadcrumbStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(baseColor))

	filterPromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(baseColor))
	filterTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(baseColor))
	filterCursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(baseColor))

	appStyle = appStyle.BorderForeground(lipgloss.Color("#626262"))
}
