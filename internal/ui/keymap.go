package ui

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Enter     key.Binding
	Quit      key.Binding
	Back      key.Binding
	Help      key.Binding
	History   key.Binding
	Favorites key.Binding
	Favorite  key.Binding
	Filter    key.Binding
}

func newKeyMap(historyEnabled bool, favoritesEnabled bool, inErrorView bool) KeyMap {
	historyBinding := key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("h", "history"),
	)

	if !historyEnabled || inErrorView {
		historyBinding.SetEnabled(false)
	}

	favoritesBinding := key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "favorites"),
	)

	favoriteBinding := key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "toggle favorite"),
	)

	if !favoritesEnabled || inErrorView {
		favoritesBinding.SetEnabled(false)
		favoriteBinding.SetEnabled(false)
	}

	helpBinding := key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	)

	filterBinding := key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	)

	if inErrorView {
		helpBinding.SetEnabled(false)
		filterBinding.SetEnabled(false)
	}

	upBinding := key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	)

	downBinding := key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	)

	if inErrorView {
		upBinding.SetEnabled(false)
		downBinding.SetEnabled(false)
	}

	backHelp := "back"
	if inErrorView {
		backHelp = "exit"
	}

	enterHelp := "select"
	if inErrorView {
		enterHelp = "continue"
	}

	return KeyMap{
		Up:   upBinding,
		Down: downBinding,
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", enterHelp),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "backspace"),
			key.WithHelp("esc", backHelp),
		),
		Help:      helpBinding,
		Filter:    filterBinding,
		History:   historyBinding,
		Favorites: favoritesBinding,
		Favorite:  favoriteBinding,
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	if k.Help.Enabled() {
		return []key.Binding{k.Help}
	}
	return []key.Binding{k.Enter, k.Back}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	if !k.Help.Enabled() {
		return [][]key.Binding{
			{k.Enter, k.Back, k.Quit},
		}
	}

	if !k.History.Enabled() && !k.Favorites.Enabled() {
		return [][]key.Binding{
			{k.Up, k.Down, k.Enter, k.Back},
			{k.Filter, k.Help, k.Quit},
		}
	}

	if !k.History.Enabled() {
		return [][]key.Binding{
			{k.Up, k.Down, k.Enter, k.Back},
			{k.Filter, k.Favorites, k.Favorite},
			{k.Help, k.Quit},
		}
	}

	if !k.Favorites.Enabled() {
		return [][]key.Binding{
			{k.Up, k.Down, k.Enter, k.Back},
			{k.Filter, k.History},
			{k.Help, k.Quit},
		}
	}

	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Back},
		{k.Filter, k.Favorites, k.Favorite, k.History},
		{k.Help, k.Quit},
	}
}
