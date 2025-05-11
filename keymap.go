package main

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
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

func newKeyMap(historyEnabled bool, favoritesEnabled bool) keyMap {
	historyBinding := key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("h", "history"),
	)

	if !historyEnabled {
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

	if !favoritesEnabled {
		favoritesBinding.SetEnabled(false)
		favoriteBinding.SetEnabled(false)
	}

	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "backspace"),
			key.WithHelp("esc", "back"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		History:   historyBinding,
		Favorites: favoritesBinding,
		Favorite:  favoriteBinding,
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
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
