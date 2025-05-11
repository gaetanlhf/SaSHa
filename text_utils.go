package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
	"github.com/muesli/reflow/wordwrap"
)

func truncateBreadcrumb(breadcrumb string, maxWidth int) string {
	if lipgloss.Width(breadcrumb) <= maxWidth {
		return breadcrumb
	}

	indicator := "..."
	indicatorWidth := lipgloss.Width(indicator)

	if maxWidth <= indicatorWidth {
		return indicator
	}

	availableWidth := maxWidth - indicatorWidth

	breadcrumbComponents := strings.Split(breadcrumb, " > ")
	if len(breadcrumbComponents) <= 1 {
		return truncate.StringWithTail(breadcrumb, uint(maxWidth), "...")
	}

	endComponent := breadcrumbComponents[len(breadcrumbComponents)-1]
	endComponentWidth := lipgloss.Width(endComponent)

	if endComponentWidth > availableWidth {
		return indicator + " " + truncate.StringWithTail(endComponent, uint(availableWidth-1), "")
	}

	result := endComponent
	remainingWidth := availableWidth - endComponentWidth

	for i := len(breadcrumbComponents) - 2; i >= 0; i-- {
		component := breadcrumbComponents[i]
		separator := " > "

		requiredWidth := lipgloss.Width(component) + lipgloss.Width(separator)

		if remainingWidth >= requiredWidth {
			result = component + separator + result
			remainingWidth -= requiredWidth
		} else {
			return indicator + " " + result
		}
	}

	return result
}

func truncateText(text string, maxWidth int) string {
	if strings.Contains(text, "\n") {
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			if lipgloss.Width(line) > maxWidth {
				lines[i] = truncate.StringWithTail(line, uint(maxWidth), "...")
			}
		}
		return strings.Join(lines, "\n")
	}

	if lipgloss.Width(text) <= maxWidth {
		return text
	}

	return truncate.StringWithTail(text, uint(maxWidth), "...")
}

func wrapText(text string, maxWidth int) string {
	if lipgloss.Width(text) <= maxWidth {
		return text
	}

	return wordwrap.String(text, maxWidth)
}
