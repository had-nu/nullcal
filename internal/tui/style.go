// Package tui provides the terminal user interface for nullcal.
//
// The TUI is built with Bubbletea and Lipgloss. It provides two views:
// a week calendar view and a kanban board, plus an editor modal for
// creating and editing tasks.
package tui

import "github.com/charmbracelet/lipgloss"

// Color palette. Monochrome with accent for active elements.
var (
	colorFg        = lipgloss.Color("#e0e0e0")
	colorDim       = lipgloss.Color("#666680")
	colorAccent    = lipgloss.Color("#00d4aa")
	colorWarning   = lipgloss.Color("#ffaa00")
	colorBorder    = lipgloss.Color("#333355")
	colorHeaderBg  = lipgloss.Color("#0f0f23")
	colorDoneFg    = lipgloss.Color("#44aa44")
	colorDoingFg   = lipgloss.Color("#dddd44")
	colorBacklogFg = lipgloss.Color("#8888aa")
)

// Header styles.
var (
	headerStyle = lipgloss.NewStyle().
			Background(colorHeaderBg).
			Foreground(colorAccent).
			Bold(true).
			Padding(0, 1)

	weekInfoStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Padding(0, 1)
)

// Column styles (shared by week view and kanban).
var (
	columnStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	activeColumnStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorAccent).
				Padding(0, 1)

	columnHeaderStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true).
				Align(lipgloss.Center)
)

// Task item styles.
var (
	taskStyle = lipgloss.NewStyle().
			Foreground(colorFg)

	selectedTaskStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	completedTaskStyle = lipgloss.NewStyle().
				Foreground(colorDoneFg).
				Strikethrough(true)
)

// Routine block style.
var routineBlockStyle = lipgloss.NewStyle().
	Foreground(colorWarning).
	Bold(true)

// Status bar and help styles.
var (
	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorDim)
)

// Editor modal styles.
var (
	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colorAccent).
			Padding(1, 2).
			Width(50)

	inputLabelStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	inputStyle = lipgloss.NewStyle().
			Foreground(colorFg).
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	activeInputStyle = lipgloss.NewStyle().
				Foreground(colorFg).
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorAccent).
				Padding(0, 1)
)

// Kanban-specific status label styles.
var (
	backlogLabelStyle = lipgloss.NewStyle().
				Foreground(colorBacklogFg).
				Bold(true)

	doingLabelStyle = lipgloss.NewStyle().
			Foreground(colorDoingFg).
			Bold(true)

	doneLabelStyle = lipgloss.NewStyle().
			Foreground(colorDoneFg).
			Bold(true)
)
