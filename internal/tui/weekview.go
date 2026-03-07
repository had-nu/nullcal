package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/had-nu/nullcal/internal/config"
	"github.com/had-nu/nullcal/internal/store"
	"github.com/had-nu/nullcal/pkg/timeutil"
)

// weekView renders a 7-column calendar grid for the current week.
type weekView struct {
	currentWeek time.Time // any day in the current week (used to derive bounds)
	focusCol    int       // 0-6 (Mon-Sun)
	focusTasks  int       // index of selected task within the focused column
}

func newWeekView() weekView {
	return weekView{
		currentWeek: time.Now(),
		focusCol:    int(time.Now().Weekday()+6) % 7, // Monday=0
	}
}

// renderWeekView renders the complete week view including header, grid, and tasks.
func renderWeekView(wv weekView, tasks []store.Task, blocks []config.RoutineBlock, width int) string {
	monday, _ := timeutil.WeekBounds(wv.currentWeek)
	days := timeutil.DaysOfWeek(wv.currentWeek)
	weekNum := timeutil.WeekNumber(wv.currentWeek)

	_, isoWeek := monday.ISOWeek()
	_ = isoWeek

	// Big pixelated ASCII header for NULLCAL (inspired by CORE sample)
	asciiHeader := `
  _ _ _ _   _       __    __    ___    ____   ___
 | \ | | | | |     / /   / /   / __\  /  _ \ |  |
 | \|  | | | |    / /   / /   / /    /  /_\ \|  |
 |  |  | |_| |   / /__ / /__ / /___ /  ___  \|  |___
 |_/ \_|\___/   /____//____/ \____//__/   \__\_____/
`
	headerTitle := lipgloss.NewStyle().Foreground(colorFg).Bold(true).Render(asciiHeader)

	// Header format: DD.MM.YY and full weekday
	dateStr := wv.currentWeek.Format("02. 01. 26") // '26' is the hardcoded year format string
	headerInfo := lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Render(fmt.Sprintf("%s    %s", dateStr, strings.ToUpper(wv.currentWeek.Weekday().String()))),
		lipgloss.NewStyle().Foreground(colorDim).Render(fmt.Sprintf("W%02d    nullcal v1.0", weekNum)),
	)
	
	header := lipgloss.JoinHorizontal(lipgloss.Bottom, headerTitle, "    ", headerInfo) + "\n"

	// Calculate column width.
	colWidth := (width - 2) / 7
	if colWidth < 12 {
		colWidth = 12
	}

	// Day headers.
	var dayHeaders []string
	for i, d := range days {
		label := fmt.Sprintf("%s %02d", shortWeekday(d.Weekday()), d.Day())
		style := columnHeaderStyle.Width(colWidth - 4)
		if i == wv.focusCol {
			style = style.Foreground(colorAccent)
		}
		dayHeaders = append(dayHeaders, style.Render(label))
	}
	headerRow := lipgloss.JoinHorizontal(lipgloss.Top, dayHeaders...)

	// Build columns with routine blocks and tasks.
	var columns []string
	for i, d := range days {
		var lines []string

		// Routine blocks for this weekday.
		for _, rb := range blocks {
			if rb.Weekday == d.Weekday() {
				lines = append(lines,
					routineBlockStyle.Render(fmt.Sprintf("# %s", rb.Label)),
					routineBlockStyle.Render(fmt.Sprintf("  %s-%s", rb.StartTime, rb.EndTime)),
				)
			}
		}

		// Tasks due this day.
		taskIdx := 0
		for _, t := range tasks {
			if t.DueAt == nil {
				continue
			}
			if !sameDay(*t.DueAt, d) {
				continue
			}

			prefix := "- "
			style := taskStyle
			if t.Status == store.TaskStatusDone {
				prefix = "x "
				style = completedTaskStyle
			} else if i == wv.focusCol && taskIdx == wv.focusTasks {
				prefix = "> "
				style = selectedTaskStyle
			}

			lines = append(lines, style.Render(prefix+truncate(t.Title, colWidth-6)))
			taskIdx++
		}

		if len(lines) == 0 {
			lines = append(lines, lipgloss.NewStyle().Foreground(colorDim).Render("  --"))
		}

		content := strings.Join(lines, "\n")
		style := columnStyle.Width(colWidth - 2).Height(12)
		if i == wv.focusCol {
			style = activeColumnStyle.Width(colWidth - 2).Height(12)
		}
		columns = append(columns, style.Render(content))
	}

	grid := lipgloss.JoinHorizontal(lipgloss.Top, columns...)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		headerRow,
		grid,
	)
}

func shortWeekday(wd time.Weekday) string {
	names := [7]string{"SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"}
	return names[wd]
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 4 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
