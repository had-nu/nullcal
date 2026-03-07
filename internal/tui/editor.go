package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/had-nu/nullcal/internal/store"
)

// editorField identifies which field is currently focused in the editor.
type editorField int

const (
	fieldTitle editorField = iota
	fieldDescription
	fieldProjectTag
	fieldDueDate
	fieldCount
)

// editor is a modal overlay for creating and editing tasks.
type editor struct {
	inputs [fieldCount]textinput.Model
	focus  editorField
	taskID string // empty for new tasks
	isEdit bool
	err    string
}

func newEditor() editor {
	var inputs [fieldCount]textinput.Model

	inputs[fieldTitle] = textinput.New()
	inputs[fieldTitle].Placeholder = "Task title"
	inputs[fieldTitle].CharLimit = 100
	inputs[fieldTitle].Width = 40
	inputs[fieldTitle].Focus()

	inputs[fieldDescription] = textinput.New()
	inputs[fieldDescription].Placeholder = "Description (optional)"
	inputs[fieldDescription].CharLimit = 256
	inputs[fieldDescription].Width = 40

	inputs[fieldProjectTag] = textinput.New()
	inputs[fieldProjectTag].Placeholder = "Project tag (optional)"
	inputs[fieldProjectTag].CharLimit = 30
	inputs[fieldProjectTag].Width = 40

	inputs[fieldDueDate] = textinput.New()
	inputs[fieldDueDate].Placeholder = "Due date: YYYY-MM-DD (optional)"
	inputs[fieldDueDate].CharLimit = 10
	inputs[fieldDueDate].Width = 40

	return editor{inputs: inputs}
}

// editTask prepares the editor for editing an existing task.
func editTask(t store.Task) editor {
	ed := newEditor()
	ed.taskID = t.ID
	ed.isEdit = true

	ed.inputs[fieldTitle].SetValue(t.Title)
	ed.inputs[fieldDescription].SetValue(t.Description)
	ed.inputs[fieldProjectTag].SetValue(t.ProjectTag)
	if t.DueAt != nil {
		ed.inputs[fieldDueDate].SetValue(t.DueAt.Format("2006-01-02"))
	}

	return ed
}

// updateEditor handles input events for the editor modal.
func updateEditor(ed editor, msg tea.KeyMsg) (editor, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field"))):
		ed.inputs[ed.focus].Blur()
		ed.focus = (ed.focus + 1) % fieldCount
		cmd := ed.inputs[ed.focus].Focus()
		return ed, cmd

	case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev field"))):
		ed.inputs[ed.focus].Blur()
		ed.focus = (ed.focus - 1 + fieldCount) % fieldCount
		cmd := ed.inputs[ed.focus].Focus()
		return ed, cmd
	}

	var cmd tea.Cmd
	ed.inputs[ed.focus], cmd = ed.inputs[ed.focus].Update(msg)
	return ed, cmd
}

// validate checks whether the editor has valid input.
func (ed *editor) validate() error {
	title := strings.TrimSpace(ed.inputs[fieldTitle].Value())
	if title == "" {
		return fmt.Errorf("title is required")
	}

	dateStr := strings.TrimSpace(ed.inputs[fieldDueDate].Value())
	if dateStr != "" {
		if _, err := time.Parse("2006-01-02", dateStr); err != nil {
			return fmt.Errorf("invalid date format, expected YYYY-MM-DD")
		}
	}

	return nil
}

// toTask converts the editor state into a Task.
func (ed *editor) toTask() store.Task {
	t := store.Task{
		ID:          ed.taskID,
		Title:       strings.TrimSpace(ed.inputs[fieldTitle].Value()),
		Description: strings.TrimSpace(ed.inputs[fieldDescription].Value()),
		ProjectTag:  strings.TrimSpace(ed.inputs[fieldProjectTag].Value()),
		Status:      store.TaskStatusBacklog,
	}

	dateStr := strings.TrimSpace(ed.inputs[fieldDueDate].Value())
	if dateStr != "" {
		if d, err := time.Parse("2006-01-02", dateStr); err == nil {
			t.DueAt = &d
		}
	}

	return t
}

// renderEditor renders the editor modal.
func renderEditor(ed editor) string {
	title := "New Task"
	if ed.isEdit {
		title = "Edit Task"
	}

	labels := [fieldCount]string{
		"Title:",
		"Description:",
		"Project Tag:",
		"Due Date:",
	}

	var lines []string
	lines = append(lines, inputLabelStyle.Render(title))
	lines = append(lines, "")

	for i := editorField(0); i < fieldCount; i++ {
		lines = append(lines, inputLabelStyle.Render(labels[i]))
		style := inputStyle
		if i == ed.focus {
			style = activeInputStyle
		}
		lines = append(lines, style.Render(ed.inputs[i].View()))
		lines = append(lines, "")
	}

	if ed.err != "" {
		lines = append(lines, lipgloss.NewStyle().
			Foreground(colorWarning).Render("! "+ed.err))
		lines = append(lines, "")
	}

	lines = append(lines, helpStyle.Render("tab: next field | enter: save | esc: cancel"))

	return modalStyle.Render(strings.Join(lines, "\n"))
}
