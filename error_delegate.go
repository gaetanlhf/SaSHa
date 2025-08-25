package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
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
	i, ok := listItem.(item)
	if !ok {
		return
	}

	var maxWidth int
	if m.Width() > 0 {
		maxWidth = m.Width() - 4
	} else {
		maxWidth = 80
	}

	title := truncateText(i.title, maxWidth)
	str := fmt.Sprintf("  %s", title)

	if i.color != "" {
		color := lipgloss.Color(i.color)
		style := lipgloss.NewStyle().Foreground(color)
		str = style.Render(str)
	}

	fmt.Fprint(w, str)
}
