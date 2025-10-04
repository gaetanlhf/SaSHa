package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gaetanlhf/SaSHa/internal/utils"
)

type errorDelegate struct {
	showDesc bool
}

func newErrorDelegate() errorDelegate {
	return errorDelegate{
		showDesc: false,
	}
}

func (d errorDelegate) Height() int { return 1 }

func (d errorDelegate) Spacing() int { return 0 }

func (d errorDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d errorDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Item)
	if !ok {
		return
	}

	var maxWidth int
	if m.Width() > 0 {
		maxWidth = m.Width() - 4
	} else {
		maxWidth = 80
	}

	title := utils.TruncateText(i.Title, maxWidth)
	str := fmt.Sprintf("  %s", title)

	if i.Color != "" {
		color := lipgloss.Color(i.Color)
		style := lipgloss.NewStyle().Foreground(color)
		str = style.Render(str)
	}

	fmt.Fprint(w, str)
}
