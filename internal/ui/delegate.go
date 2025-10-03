package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gaetanlhf/sasha/internal/utils"
)

type ColoredDelegate struct {
	DefaultDelegate list.DefaultDelegate
	CurrentColor    string
	InHistoryView   bool
	InFavoritesView bool
}

func NewColoredDelegate() ColoredDelegate {
	d := list.NewDefaultDelegate()

	return ColoredDelegate{
		DefaultDelegate: d,
		CurrentColor:    "",
		InHistoryView:   false,
		InFavoritesView: false,
	}
}

func (d ColoredDelegate) Height() int {
	if d.InHistoryView {
		return 4
	} else if d.InFavoritesView {
		return 3
	}
	return d.DefaultDelegate.Height()
}

func (d ColoredDelegate) Spacing() int {
	return d.DefaultDelegate.Spacing()
}

func (d ColoredDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return d.DefaultDelegate.Update(msg, m)
}

func (d ColoredDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	var (
		title, desc, colorToUse string
		isMultiline             bool
		pathEntries, pathColors []string
	)

	// Gestion des différents types d'items
	switch i := listItem.(type) {
	case Item:
		title = i.Title
		desc = i.Description
		colorToUse = i.Color
		isMultiline = i.IsMultiline
		pathEntries = i.PathEntries
		pathColors = i.PathColors
	case interface{ Title() string }:
		title = i.Title()
		desc = ""
		if d, ok := listItem.(interface{ Description() string }); ok {
			desc = d.Description()
		}
		if c, ok := listItem.(interface{ GetColor() string }); ok {
			colorToUse = c.GetColor()
		}
		if ml, ok := listItem.(interface{ IsMultiline() bool }); ok {
			isMultiline = ml.IsMultiline()
		}
		if pe, ok := listItem.(interface{ GetPathEntries() []string }); ok {
			pathEntries = pe.GetPathEntries()
		}
		if pc, ok := listItem.(interface{ GetPathColors() []string }); ok {
			pathColors = pc.GetPathColors()
		}
	default:
		return
	}

	if colorToUse == "" {
		colorToUse = d.CurrentColor
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

	if isMultiline && strings.Contains(desc, "\n") {
		lines := strings.Split(desc, "\n")

		for idx, line := range lines {
			if strings.Contains(line, "📁 ") && len(pathEntries) > 0 && isSelected {
				pathPrefix := "📁 "
				coloredPath := formatColoredPath(pathEntries, pathColors, colorToUse)
				lines[idx] = pathPrefix + coloredPath
			} else {
				lines[idx] = utils.TruncateText(line, maxWidth)
			}
		}
		formattedDesc = strings.Join(lines, "\n")
	} else {
		title = utils.TruncateText(title, maxWidth)
		formattedDesc = utils.TruncateText(desc, maxWidth)
	}

	var titleStyle, descStyle lipgloss.Style

	if isSelected {
		titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorToUse)).
			PaddingLeft(2)

		title = "▶ " + title

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
