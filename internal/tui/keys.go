package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keybindings for nullcal.
type KeyMap struct {
	Left      key.Binding
	Right     key.Binding
	Up        key.Binding
	Down      key.Binding
	New       key.Binding
	Edit      key.Binding
	Delete    key.Binding
	Toggle    key.Binding
	Move      key.Binding
	SwitchTab key.Binding
	Confirm   key.Binding
	Cancel    key.Binding
	Help      key.Binding
	Quit      key.Binding
}

// DefaultKeyMap returns the default keybindings for nullcal.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/left", "previous"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l/right", "next"),
		),
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/up", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/down", "down"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new task"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Delete: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "delete"),
		),
		Toggle: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "toggle done"),
		),
		Move: key.NewBinding(
			key.WithKeys("m", "enter"),
			key.WithHelp("m/enter", "move status"),
		),
		SwitchTab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch view"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// ShortHelp returns the help text for the main view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.New, k.Edit, k.Delete, k.Toggle, k.SwitchTab, k.Quit, k.Help,
	}
}

// FullHelp returns the full help text.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right, k.Up, k.Down},
		{k.New, k.Edit, k.Delete, k.Toggle},
		{k.Move, k.SwitchTab, k.Quit},
	}
}
