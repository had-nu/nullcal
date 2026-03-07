package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/had-nu/nullcal/internal/store"
)

// kanbanView renders a 3-column kanban board: backlog | doing | done.
type kanbanView struct {
	focusCol  int // 0=backlog, 1=doing, 2=done
	focusTask int // index of selected task within the focused column
}

func newKanbanView() kanbanView {
	return kanbanView{}
}

// kanbanColumn maps column index to TaskStatus.
var kanbanColumns = [3]store.TaskStatus{
	store.TaskStatusBacklog,
	store.TaskStatusDoing,
	store.TaskStatusDone,
}

// kanbanColumnLabels maps column index to display label.
var kanbanColumnLabels = [3]string{
	"BACKLOG",
	"DOING",
	"DONE",
}

// renderKanban renders the kanban board with three status columns.
func renderKanban(kv kanbanView, tasks []store.Task, width int) string {
	header := headerStyle.Render("nullcal") + "  " +
		weekInfoStyle.Render("kanban board")

	colWidth := (width - 2) / 3
	if colWidth < 16 {
		colWidth = 16
	}

	// Group tasks by status.
	grouped := map[store.TaskStatus][]store.Task{}
	for _, t := range tasks {
		grouped[t.Status] = append(grouped[t.Status], t)
	}

	var columns []string
	for i, status := range kanbanColumns {
		col := grouped[status]

		// Column header with count.
		labelStyle := backlogLabelStyle
		switch status {
		case store.TaskStatusDoing:
			labelStyle = doingLabelStyle
		case store.TaskStatusDone:
			labelStyle = doneLabelStyle
		}
		colHeader := labelStyle.Render(fmt.Sprintf(" %s (%d) ",
			kanbanColumnLabels[i], len(col)))

		var lines []string
		lines = append(lines, colHeader)
		lines = append(lines, "")

		for j, t := range col {
			prefix := "  "
			style := taskStyle
			if t.Status == store.TaskStatusDone {
				prefix = "x "
				style = completedTaskStyle
			}
			if i == kv.focusCol && j == kv.focusTask {
				prefix = "> "
				style = selectedTaskStyle
			}

			taskLine := style.Render(prefix + truncate(t.Title, colWidth-6))

			// Show project tag if present.
			if t.ProjectTag != "" {
				tag := lipgloss.NewStyle().Foreground(colorDim).Render(
					fmt.Sprintf("  [%s]", t.ProjectTag))
				taskLine += "\n" + tag
			}

			lines = append(lines, taskLine)
		}

		if len(col) == 0 {
			lines = append(lines,
				lipgloss.NewStyle().Foreground(colorDim).Render("  (empty)"))
		}

		content := strings.Join(lines, "\n")
		style := columnStyle.Width(colWidth - 2).Height(16)
		if i == kv.focusCol {
			style = activeColumnStyle.Width(colWidth - 2).Height(16)
		}
		columns = append(columns, style.Render(content))
	}

	grid := lipgloss.JoinHorizontal(lipgloss.Top, columns...)

	return lipgloss.JoinVertical(lipgloss.Left, header, grid)
}

// nextStatus returns the next kanban status in the workflow.
// backlog -> doing -> done. Done wraps back to backlog.
func nextStatus(current store.TaskStatus) store.TaskStatus {
	switch current {
	case store.TaskStatusBacklog:
		return store.TaskStatusDoing
	case store.TaskStatusDoing:
		return store.TaskStatusDone
	default:
		return store.TaskStatusBacklog
	}
}
