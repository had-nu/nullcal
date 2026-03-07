package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/had-nu/nullcal/internal/config"
	"github.com/had-nu/nullcal/internal/store"
	"github.com/had-nu/nullcal/pkg/timeutil"
)

type focusPane int

const (
	focusCalendar focusPane = iota
	focusTodo
	focusKanban
)

type dashboardView struct {
	focus       focusPane
	currentWeek time.Time

	// Selection state within panes
	calCol   int // 0-6 (Mon-Sun)
	calTask  int
	todoTask int
	kanCol   int // 0-2 (Backlog, Doing, Done)
	kanTask  int
}

func newDashboardView(pane string) dashboardView {
	focus := focusCalendar
	if pane == "todo" { focus = focusTodo }
	if pane == "kanban" { focus = focusKanban }

	return dashboardView{
		focus:       focus,
		currentWeek: time.Now(),
		calCol:      int(time.Now().Weekday()+6) % 7, // Monday=0
	}
}

func (d *dashboardView) cycleFocus() {
	// disabled in isolated pane mode. Native OS triggers splits.
}

func (d dashboardView) update(msg tea.KeyMsg) (dashboardView, tea.Cmd) {
	// We use the default keymap from model, but hardcode the standard vim keys here for simplicity in the view delegate
	switch d.focus {
	case focusCalendar:
		switch msg.String() {
		case "up", "k":
			if d.calTask > 0 {
				d.calTask--
			}
		case "down", "j":
			d.calTask++
		case "left", "h":
			d.currentWeek = d.currentWeek.AddDate(0, 0, -7)
			return d, nil // usually triggers loadTasks in main loop, we'll return a stub message if needed
		case "right", "l":
			d.currentWeek = d.currentWeek.AddDate(0, 0, 7)
			return d, nil
		case "H":
			if d.calCol > 0 {
				d.calCol--
				d.calTask = 0
			}
		case "L":
			if d.calCol < 6 {
				d.calCol++
				d.calTask = 0
			}
		}

	case focusTodo:
		switch msg.String() {
		case "up", "k":
			if d.todoTask > 0 {
				d.todoTask--
			}
		case "down", "j":
			d.todoTask++
		}

	case focusKanban:
		switch msg.String() {
		case "left", "h":
			if d.kanCol > 0 {
				d.kanCol--
				d.kanTask = 0
			}
		case "right", "l":
			if d.kanCol < 2 {
				d.kanCol++
				d.kanTask = 0
			}
		case "up", "k":
			if d.kanTask > 0 {
				d.kanTask--
			}
		case "down", "j":
			d.kanTask++
		}
	}
	return d, nil
}

func (d dashboardView) selectedTask(allTasks []store.Task) *store.Task {
	switch d.focus {
	case focusCalendar:
		days := timeutil.DaysOfWeek(d.currentWeek)
		day := days[d.calCol]
		idx := 0
		for i := range allTasks {
			if allTasks[i].DueAt != nil && sameDay(*allTasks[i].DueAt, day) {
				if idx == d.calTask {
					return &allTasks[i]
				}
				idx++
			}
		}
	case focusTodo:
		idx := 0
		for i := range allTasks {
			if allTasks[i].Status != store.TaskStatusDone {
				if idx == d.todoTask {
					return &allTasks[i]
				}
				idx++
			}
		}
	case focusKanban:
		status := kanbanColumns[d.kanCol]
		idx := 0
		for i := range allTasks {
			if allTasks[i].Status == status {
				if idx == d.kanTask {
					return &allTasks[i]
				}
				idx++
			}
		}
	}
	return nil
}

func (d dashboardView) view(tasks []store.Task, blocks []config.RoutineBlock, width int, height int) string {
	// ASCII Dashboard Header (Minimalized for individual pane usage)
	dateStr := time.Now().Format("02. 01. 26")
	headerInfo := lipgloss.JoinHorizontal(lipgloss.Bottom,
		lipgloss.NewStyle().Foreground(colorFg).Bold(true).Render("NULLCAL"),
		"    ",
		headerStyle.Render(fmt.Sprintf("%s    %s", dateStr, strings.ToUpper(time.Now().Weekday().String()))),
	)
	
	header := headerInfo + "\n\n"

	h := height - 4 // subtract header and status bar boundaries
	if h < 10 {
		h = 10
	}

	var pane string
	switch d.focus {
	case focusCalendar:
		pane = d.renderCalendar(tasks, blocks, width, h)
	case focusTodo:
		pane = d.renderTodo(tasks, width, h)
	case focusKanban:
		pane = d.renderKanban(tasks, width, h)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, pane)
}

func (d dashboardView) renderCalendar(tasks []store.Task, blocks []config.RoutineBlock, width, height int) string {
	bColor := colorBorder
	if d.focus == focusCalendar {
		bColor = colorFg
	}
	style := columnStyle.Width(width - 2).Height(height).BorderForeground(bColor)
	
	_, _ = timeutil.WeekBounds(d.currentWeek) // Remove if monday unused, kept for future use if needed
	days := timeutil.DaysOfWeek(d.currentWeek)

	// Calculate column width for the 7 days
	colWidth := (width - 4) / 7
	if colWidth < 8 {
		colWidth = 8
	}

	var columns []string
	for i, day := range days {
		var lines []string
		
		// Day Header
		headerLabel := fmt.Sprintf("%s %02d", shortWeekday(day.Weekday()), day.Day())
		hStyle := columnHeaderStyle.Width(colWidth - 2)
		if i == d.calCol && d.focus == focusCalendar {
			hStyle = hStyle.Foreground(colorAccent).Background(colorDim)
		}
		lines = append(lines, hStyle.Render(headerLabel), "")

		// RoutineBlocks
		for _, rb := range blocks {
			if rb.Weekday == day.Weekday() {
				lines = append(lines,
					routineBlockStyle.Render(fmt.Sprintf("%s", rb.Label)),
					routineBlockStyle.Render(fmt.Sprintf("%s-%s", rb.StartTime, rb.EndTime)),
					"",
				)
			}
		}

		// Tasks
		taskIdx := 0
		for _, t := range tasks {
			if t.DueAt == nil || !sameDay(*t.DueAt, day) {
				continue
			}

			prefix := "- "
			tStyle := taskStyle
			
			// Semantic coloring
			if t.Status == store.TaskStatusDone {
				prefix = "x "
				tStyle = completedTaskStyle
			} else {
				hoursUntil := time.Until(*t.DueAt).Hours()
				if hoursUntil < 0 {
					tStyle = overdueTaskStyle
				} else if hoursUntil < 48 {
					tStyle = dueSoonTaskStyle
				}
			}

			if d.focus == focusCalendar && i == d.calCol && taskIdx == d.calTask {
				prefix = "> "
				tStyle = selectedTaskStyle
			}

			lines = append(lines, tStyle.Render(prefix+truncate(t.Title, colWidth-4)))
			taskIdx++
		}

		content := strings.Join(lines, "\n")
		// We use a basic padding style for the inner columns
		colStyle := lipgloss.NewStyle().Width(colWidth).Padding(0, 1)
		columns = append(columns, colStyle.Render(content))
	}

	grid := lipgloss.JoinHorizontal(lipgloss.Top, columns...)
	return style.Render(grid)
}

func shortWeekday(wd time.Weekday) string {
	names := [7]string{"SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"}
	return names[wd]
}

func (d dashboardView) renderTodo(tasks []store.Task, width, height int) string {
	bColor := colorBorder
	if d.focus == focusTodo {
		bColor = colorFg
	}
	style := columnStyle.Width(width - 2).Height(height).BorderForeground(bColor)

	var lines []string
	headerLabel := "TO-DO LIST"
	hStyle := columnHeaderStyle
	if d.focus == focusTodo {
		hStyle = hStyle.Foreground(colorAccent).Background(colorDim)
	}
	lines = append(lines, hStyle.Render(headerLabel), "")

	// Render all active tasks (not done) that don't have a due date (backlog concept)
	// For MVP, just render all active tasks as a flat list
	taskIdx := 0
	for _, t := range tasks {
		if t.Status == store.TaskStatusDone {
			continue // Skip done tasks in this simplified view
		}

		prefix := "- "
		tStyle := taskStyle
		
		if t.DueAt != nil {
			hoursUntil := time.Until(*t.DueAt).Hours()
			if hoursUntil < 0 {
				tStyle = overdueTaskStyle
			} else if hoursUntil < 48 {
				tStyle = dueSoonTaskStyle
			}
		}

		if d.focus == focusTodo && taskIdx == d.todoTask {
			prefix = "> "
			tStyle = selectedTaskStyle
		}

		lines = append(lines, tStyle.Render(prefix+truncate(t.Title, width-6)))
		taskIdx++
	}

	content := strings.Join(lines, "\n")
	return style.Render(content)
}

var kanbanColumns = []store.TaskStatus{
	store.TaskStatusBacklog,
	store.TaskStatusDoing,
	store.TaskStatusDone,
}

func (d dashboardView) renderKanban(tasks []store.Task, width, height int) string {
	bColor := colorBorder
	if d.focus == focusKanban {
		bColor = colorFg
	}
	style := columnStyle.Width(width - 2).Height(height).BorderForeground(bColor)

	colWidth := (width - 4) / 3
	if colWidth < 10 {
		colWidth = 10
	}

	var columns []string
	labels := []string{"BACKLOG", "DOING", "DONE"}

	for i, status := range kanbanColumns {
		var lines []string
		
		hStyle := columnHeaderStyle.Width(colWidth - 2)
		if d.focus == focusKanban && i == d.kanCol {
			hStyle = hStyle.Foreground(colorAccent).Background(colorDim)
		}
		lines = append(lines, hStyle.Render(labels[i]), "")

		taskIdx := 0
		for _, t := range tasks {
			if t.Status != status {
				continue
			}

			prefix := "- "
			tStyle := taskStyle
			
			if status == store.TaskStatusDone {
				prefix = "x "
				tStyle = completedTaskStyle
			} else if t.DueAt != nil {
				hoursUntil := time.Until(*t.DueAt).Hours()
				if hoursUntil < 0 {
					tStyle = overdueTaskStyle
				} else if hoursUntil < 48 {
					tStyle = dueSoonTaskStyle
				}
			}

			if d.focus == focusKanban && i == d.kanCol && taskIdx == d.kanTask {
				prefix = "> "
				tStyle = selectedTaskStyle
			}

			lines = append(lines, tStyle.Render(prefix+truncate(t.Title, colWidth-4)))
			
			if t.ProjectTag != "" {
				tag := lipgloss.NewStyle().Foreground(colorBg).Background(colorDim).Padding(0, 1).Render(t.ProjectTag)
				lines = append(lines, "  "+tag)
			}
			lines = append(lines, "") // spacing
			
			taskIdx++
		}

		colStyle := lipgloss.NewStyle().Width(colWidth).Padding(0, 1)
		columns = append(columns, colStyle.Render(strings.Join(lines, "\n")))
	}

	grid := lipgloss.JoinHorizontal(lipgloss.Top, columns...)
	return style.Render(grid)
}

// nextStatus cycles task status for quick moving
func nextStatus(s store.TaskStatus) store.TaskStatus {
	switch s {
	case store.TaskStatusBacklog:
		return store.TaskStatusDoing
	case store.TaskStatusDoing:
		return store.TaskStatusDone
	case store.TaskStatusDone:
		return store.TaskStatusBacklog
	default:
		return store.TaskStatusBacklog
	}
}

// sameDay checks if two dates represent the same calendar day
func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

// truncate shortens a string to fit a visual column
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 4 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
