// Package tui provides the terminal user interface for nullcal.
//
// The TUI is built with Bubbletea and Lipgloss. It provides two views:
// a week calendar view and a kanban board, plus an editor modal for
// creating and editing tasks.
package tui

import "github.com/charmbracelet/lipgloss"

// Color palette. Pure monochrome for CORE retro aesthetic.
var (
	colorBg       = lipgloss.Color("#000000") // Reserved for terminal bg if needed
	colorFg       = lipgloss.Color("#ffffff") // Pure white
	colorDim      = lipgloss.Color("#888888") // Mid gray
	colorAccent   = lipgloss.Color("#cccccc") // Light gray highlight
	colorWarning  = lipgloss.Color("#ffffff") // Warning (retro bright white override)
	colorBorder   = lipgloss.Color("#444444") // Dark gray structural lines
	colorHeaderBg = lipgloss.Color("#111111") // Slight offset for top bar
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
			Padding(1, 1)

	activeColumnStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorAccent).
				Padding(1, 1)

	columnHeaderStyle = lipgloss.NewStyle().
				Foreground(colorFg).
				Bold(true).
				Align(lipgloss.Left) // Align left like the sample
)

// Task item styles.
var (
	taskStyle = lipgloss.NewStyle().
			Foreground(colorFg)

	selectedTaskStyle = lipgloss.NewStyle().
				Foreground(colorBg).
				Background(colorFg).
				Bold(true)

	completedTaskStyle = lipgloss.NewStyle().
				Foreground(colorDim).
				Strikethrough(true)
)

// Routine block style.
var routineBlockStyle = lipgloss.NewStyle().
	Foreground(colorDim).
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
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorFg).
			Padding(1, 2).
			Width(50)

	inputLabelStyle = lipgloss.NewStyle().
			Foreground(colorFg).
			Bold(true)

	inputStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	activeInputStyle = lipgloss.NewStyle().
				Foreground(colorFg).
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorFg).
				Padding(0, 1)
)

// Kanban-specific status label styles.
var (
	backlogLabelStyle = lipgloss.NewStyle().
				Foreground(colorDim).
				Bold(true)

	doingLabelStyle = lipgloss.NewStyle().
			Foreground(colorFg).
			Bold(true)

	doneLabelStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Bold(true).
			Strikethrough(true)
)
