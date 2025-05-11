package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"strings"
)

type ColoredDelegate struct {
	defaultDelegate list.DefaultDelegate
	currentColor    string
	inHistoryView   bool
	inFavoritesView bool
}

func NewColoredDelegate() ColoredDelegate {
	d := list.NewDefaultDelegate()

	return ColoredDelegate{
		defaultDelegate: d,
		currentColor:    "",
		inHistoryView:   false,
		inFavoritesView: false,
	}
}

func (d ColoredDelegate) Height() int {
	if d.inHistoryView {
		return 4
	} else if d.inFavoritesView {
		return 3
	}
	return d.defaultDelegate.Height()
}

func (d ColoredDelegate) Spacing() int {
	return d.defaultDelegate.Spacing()
}

func (d ColoredDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return d.defaultDelegate.Update(msg, m)
}

func (d ColoredDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	title := i.Title()
	desc := i.Description()

	colorToUse := i.color
	if colorToUse == "" {
		colorToUse = d.currentColor
	}

	if colorToUse == "" {
		colorToUse = "#FFFFFF"
	}

	var maxWidth int
	if m.Width() > 0 {
		maxWidth = m.Width() - 4
	} else {
		maxWidth = 80
	}

	var formattedDesc string
	isSelected := index == m.Index()

	if i.isMultiline && strings.Contains(desc, "\n") {
		lines := strings.Split(desc, "\n")

		for idx, line := range lines {
			if strings.Contains(line, "📁 ") && len(i.pathEntries) > 0 && isSelected {
				pathPrefix := "📁 "

				coloredPath := formatColoredPath(i.pathEntries, i.pathColors, colorToUse)
				lines[idx] = pathPrefix + coloredPath
			} else {
				lines[idx] = truncateText(line, maxWidth)
			}
		}
		formattedDesc = strings.Join(lines, "\n")
	} else {
		title = truncateText(title, maxWidth)
		formattedDesc = truncateText(desc, maxWidth)
	}

	var (
		titleStyle, descStyle lipgloss.Style
	)

	if isSelected {
		titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorToUse)).
			PaddingLeft(2)

		if i.isGroup {
			title = "▶ " + title
		} else {
			title = "▶ " + title
		}

		descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorToUse)).
			PaddingLeft(4)
	} else {
		titleStyle = lipgloss.NewStyle().
			PaddingLeft(4).
			Foreground(lipgloss.Color("#FFFFFF"))

		descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")).
			PaddingLeft(4)
	}

	titleLine := titleStyle.Render(title)
	descLine := descStyle.Render(formattedDesc)

	fmt.Fprintf(w, "%s\n%s", titleLine, descLine)
}

func formatColoredPath(pathEntries []string, pathColors []string, defaultColor string) string {
	if len(pathEntries) == 0 {
		return ""
	}

	separatorStyle := helpStyle
	separator := separatorStyle.Render(" > ")

	parts := make([]string, len(pathEntries))

	for i, entry := range pathEntries {
		color := defaultColor
		if i < len(pathColors) && pathColors[i] != "" {
			color = pathColors[i]
		}

		style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
		parts[i] = style.Render(entry)
	}

	return strings.Join(parts, separator)
}
