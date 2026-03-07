package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/had-nu/nullcal/internal/config"
	"github.com/had-nu/nullcal/internal/store"
	"github.com/had-nu/nullcal/pkg/timeutil"
)

// ViewMode defines the active view.
type ViewMode int

// ViewWeek and ViewKanban define the active TUI view.
const (
	ViewWeek ViewMode = iota
	ViewKanban
)

// Model is the root Bubbletea model for nullcal.
type Model struct {
	store  *store.Store
	config *config.Config
	keys   KeyMap
	help   help.Model

	view   ViewMode
	week   weekView
	kanban kanbanView
	editor editor

	editing    bool
	confirmDel bool
	deleteID   string
	showHelp   bool

	tasks  []store.Task
	width  int
	height int
	err    string
}

// New creates a new TUI model with the given store and config.
func New(s *store.Store, cfg *config.Config) Model {
	return Model{
		store:  s,
		config: cfg,
		keys:   DefaultKeyMap(),
		help:   help.New(),
		view:   ViewWeek,
		week:   newWeekView(),
		kanban: newKanbanView(),
	}
}

// Init implements tea.Model. Loads tasks on startup.
func (m Model) Init() tea.Cmd {
	return m.loadTasks
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tasksLoadedMsg:
		m.tasks = msg.tasks
		m.err = ""
		return m, nil

	case errMsg:
		m.err = msg.Error()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var main string
	switch {
	case m.editing:
		main = m.viewWithModal()
	case m.confirmDel:
		main = m.viewWithConfirm()
	default:
		main = m.viewMain()
	}

	// Status bar.
	status := statusBarStyle.Render(m.statusText())
	helpView := helpStyle.Render(m.help.ShortHelpView(m.keys.ShortHelp()))

	// Error display.
	if m.err != "" {
		errLine := lipgloss.NewStyle().Foreground(colorWarning).Render("! " + m.err)
		return lipgloss.JoinVertical(lipgloss.Left, main, errLine, status, helpView)
	}

	return lipgloss.JoinVertical(lipgloss.Left, main, status, helpView)
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Modal states take priority.
	if m.editing {
		return m.handleEditorKey(msg)
	}
	if m.confirmDel {
		return m.handleConfirmKey(msg)
	}

	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.SwitchTab):
		if m.view == ViewWeek {
			m.view = ViewKanban
		} else {
			m.view = ViewWeek
		}
		return m, nil

	case key.Matches(msg, m.keys.New):
		m.editor = newEditor()
		m.editing = true
		return m, m.editor.inputs[0].Focus()

	case key.Matches(msg, m.keys.Edit):
		if t := m.selectedTask(); t != nil {
			m.editor = editTask(*t)
			m.editing = true
			return m, m.editor.inputs[0].Focus()
		}
		return m, nil

	case key.Matches(msg, m.keys.Delete):
		if t := m.selectedTask(); t != nil {
			m.confirmDel = true
			m.deleteID = t.ID
		}
		return m, nil

	case key.Matches(msg, m.keys.Toggle):
		return m.toggleSelectedTask()

	case key.Matches(msg, m.keys.Move):
		if m.view == ViewKanban {
			return m.moveSelectedTask()
		}
		return m, nil

	case key.Matches(msg, m.keys.Help):
		m.showHelp = !m.showHelp
		m.help.ShowAll = m.showHelp
		return m, nil
	}

	// Navigation.
	return m.handleNavigation(msg)
}

func (m Model) handleEditorKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.editing = false
		m.editor.err = ""
		return m, nil

	case key.Matches(msg, m.keys.Confirm):
		if err := m.editor.validate(); err != nil {
			m.editor.err = err.Error()
			return m, nil
		}

		task := m.editor.toTask()
		if m.editor.isEdit {
			// Preserve existing status.
			for _, t := range m.tasks {
				if t.ID == task.ID {
					task.Status = t.Status
					task.CompletedAt = t.CompletedAt
					break
				}
			}
			if err := m.store.UpdateTask(&task); err != nil {
				m.err = fmt.Sprintf("update failed: %v", err)
				m.editing = false
				return m, nil
			}
		} else {
			if err := m.store.CreateTask(&task); err != nil {
				m.err = fmt.Sprintf("create failed: %v", err)
				m.editing = false
				return m, nil
			}
		}

		m.editing = false
		return m, m.loadTasks
	}

	var cmd tea.Cmd
	m.editor, cmd = updateEditor(m.editor, msg)
	return m, cmd
}

func (m Model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if err := m.store.DeleteTask(m.deleteID); err != nil {
			m.err = fmt.Sprintf("delete failed: %v", err)
		}
		m.confirmDel = false
		m.deleteID = ""
		return m, m.loadTasks
	default:
		m.confirmDel = false
		m.deleteID = ""
		return m, nil
	}
}

func (m Model) handleNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.view {
	case ViewWeek:
		switch {
		case key.Matches(msg, m.keys.Left):
			m.week.currentWeek = m.week.currentWeek.AddDate(0, 0, -7)
			return m, m.loadTasks
		case key.Matches(msg, m.keys.Right):
			m.week.currentWeek = m.week.currentWeek.AddDate(0, 0, 7)
			return m, m.loadTasks
		case key.Matches(msg, m.keys.Up):
			if m.week.focusTasks > 0 {
				m.week.focusTasks--
			}
		case key.Matches(msg, m.keys.Down):
			m.week.focusTasks++
			// Clamp in view render.
		}
		// Left/right also navigate days within week with Shift.
		if msg.String() == "H" {
			if m.week.focusCol > 0 {
				m.week.focusCol--
				m.week.focusTasks = 0
			}
		}
		if msg.String() == "L" {
			if m.week.focusCol < 6 {
				m.week.focusCol++
				m.week.focusTasks = 0
			}
		}

	case ViewKanban:
		switch {
		case key.Matches(msg, m.keys.Left):
			if m.kanban.focusCol > 0 {
				m.kanban.focusCol--
				m.kanban.focusTask = 0
			}
		case key.Matches(msg, m.keys.Right):
			if m.kanban.focusCol < 2 {
				m.kanban.focusCol++
				m.kanban.focusTask = 0
			}
		case key.Matches(msg, m.keys.Up):
			if m.kanban.focusTask > 0 {
				m.kanban.focusTask--
			}
		case key.Matches(msg, m.keys.Down):
			m.kanban.focusTask++
		}
	}

	return m, nil
}

func (m Model) toggleSelectedTask() (tea.Model, tea.Cmd) {
	t := m.selectedTask()
	if t == nil {
		return m, nil
	}

	newStatus := store.TaskStatusDone
	if t.Status == store.TaskStatusDone {
		newStatus = store.TaskStatusBacklog
	}

	if err := m.store.SetTaskStatus(t.ID, newStatus); err != nil {
		m.err = fmt.Sprintf("toggle failed: %v", err)
		return m, nil
	}

	return m, m.loadTasks
}

func (m Model) moveSelectedTask() (tea.Model, tea.Cmd) {
	t := m.selectedTask()
	if t == nil {
		return m, nil
	}

	next := nextStatus(t.Status)
	if err := m.store.SetTaskStatus(t.ID, next); err != nil {
		m.err = fmt.Sprintf("move failed: %v", err)
		return m, nil
	}

	return m, m.loadTasks
}

func (m Model) selectedTask() *store.Task {
	switch m.view {
	case ViewWeek:
		days := timeutil.DaysOfWeek(m.week.currentWeek)
		d := days[m.week.focusCol]
		idx := 0
		for i := range m.tasks {
			if m.tasks[i].DueAt == nil {
				continue
			}
			if !sameDay(*m.tasks[i].DueAt, d) {
				continue
			}
			if idx == m.week.focusTasks {
				return &m.tasks[i]
			}
			idx++
		}

	case ViewKanban:
		status := kanbanColumns[m.kanban.focusCol]
		idx := 0
		for i := range m.tasks {
			if m.tasks[i].Status == status {
				if idx == m.kanban.focusTask {
					return &m.tasks[i]
				}
				idx++
			}
		}
	}

	return nil
}

func (m Model) viewMain() string {
	switch m.view {
	case ViewKanban:
		return renderKanban(m.kanban, m.tasks, m.width)
	default:
		return renderWeekView(m.week, m.tasks, m.config.RoutineBlocks, m.width)
	}
}

func (m Model) viewWithModal() string {
	bg := m.viewMain()
	modal := renderEditor(m.editor)
	return lipgloss.Place(m.width, m.height-4,
		lipgloss.Center, lipgloss.Center,
		modal,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(colorDim),
	) + "\n" + lipgloss.NewStyle().Foreground(colorDim).Render(bg[:min(len(bg), 40)])
}

func (m Model) viewWithConfirm() string {
	bg := m.viewMain()
	confirm := modalStyle.Render(
		inputLabelStyle.Render("Delete task?") + "\n\n" +
			helpStyle.Render("y: confirm | any key: cancel"),
	)
	_ = bg
	return lipgloss.Place(m.width, m.height-2,
		lipgloss.Center, lipgloss.Center,
		confirm,
	)
}

func (m Model) statusText() string {
	viewLabel := "WEEK"
	if m.view == ViewKanban {
		viewLabel = "KANBAN"
	}
	return fmt.Sprintf("[%s] %d tasks", viewLabel, len(m.tasks))
}

// Messages for async operations.

type tasksLoadedMsg struct {
	tasks []store.Task
}

type errMsg struct{ error }

func (m Model) loadTasks() tea.Msg {
	var tasks []store.Task
	var err error

	switch m.view {
	case ViewWeek:
		// Load tasks for the visible week.
		days := timeutil.DaysOfWeek(m.week.currentWeek)

		// Load all tasks for the week by querying each day.
		// A more efficient approach would be a range query, but this
		// is sufficient for MVP.
		for _, d := range days {
			dayTasks, e := m.store.ListTasksByDate(d)
			if e != nil {
				return errMsg{e}
			}
			tasks = append(tasks, dayTasks...)
		}

	case ViewKanban:
		tasks, err = m.store.ListTasks()
		if err != nil {
			return errMsg{err}
		}
	}

	return tasksLoadedMsg{tasks: tasks}
}
