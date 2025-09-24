package components

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FilePicker struct {
	picker       filepicker.Model
	allowedTypes []string

	viewport      viewport.Model
	SelectingFile bool

	Cursor  cursor.Model
	focused bool

	reacted bool
}

func NewFilePicker(width int, allowedTypes ...string) *FilePicker {
	cursor := cursor.New()
	cursor.SetChar(" ")

	if width <= 0 {
		width = 20
	}

	viewport := viewport.New(width, 1)
	viewport.SetHorizontalStep(2)

	return &FilePicker{
		Cursor:       cursor,
		picker:       newInternalFilePicker(allowedTypes),
		allowedTypes: allowedTypes,
		viewport:     viewport,
	}
}

func newInternalFilePicker(allowedTypes []string) filepicker.Model {
	filePicker := filepicker.New()
	filePicker.SetHeight(15)
	filePicker.ShowPermissions = false
	filePicker.CurrentDirectory, _ = os.Getwd()

	if len(allowedTypes) == 0 {
		allowedTypes = []string{".txt"}
	}
	filePicker.AllowedTypes = allowedTypes
	return filePicker
}

func (m *FilePicker) Focus() tea.Cmd {
	m.focused = true
	return m.Cursor.Focus()
}

func (m *FilePicker) Blur() {
	m.focused = false
	m.Cursor.Blur()
}

// Returns a wrapped string
func (m *FilePicker) Value() InputValue {
	return m.picker.Path
}

func (m *FilePicker) SetValue(filePath InputValue) error {
	// TODO: Maybe validate if a valid file is presented, but eh ¯\_(ツ)_/¯
	value, ok := filePath.(string)
	if !ok {
		return InvalidInputErr
	}

	fileTree := strings.Split(filepath.Clean(value), string(filepath.Separator))
	slices.Reverse(fileTree)

	fileTree = fileTree[:min(2, len(fileTree))]
	slices.Reverse(fileTree)

	displayedFilePath := strings.Join(fileTree, string(filepath.Separator))
	m.setViewportContent(fmt.Sprintf(" ../%v ", displayedFilePath))
	m.picker.Path = filepath.Clean(value)

	m.setReact()
	return nil
}

func (m *FilePicker) Reset() {
	m.picker = newInternalFilePicker(m.allowedTypes)
	m.viewport.SetContent("")
	m.setReact()
}

func (m *FilePicker) Init() tea.Cmd {
	return m.picker.Init()
}

func (m *FilePicker) Update(msg tea.Msg) (Input, tea.Cmd) {
	return m.update(msg)
}

func (m *FilePicker) update(msg tea.Msg) (*FilePicker, tea.Cmd) {
	if m.SelectingFile {
		m.setReact()

		if msg, isKey := msg.(tea.KeyMsg); isKey {
			if msg.String() == "esc" {
				m.SelectingFile = false
				return m, tea.Batch(tea.ExitAltScreen, m.Focus())
			}
		}

		cmds := make([]tea.Cmd, 3)
		m.picker, cmds[0] = m.picker.Update(msg)

		if didSelect, filePath := m.picker.DidSelectFile(msg); didSelect {
			m.SelectingFile = false
			m.SetValue(filePath)

			cmds[1] = tea.ExitAltScreen
			cmds[2] = m.Focus()
		}

		return m, tea.Batch(cmds[:]...)
	}

	if !m.focused {
		return m, nil
	}

	m.setReact()

	switch msg := msg.(type) {
	default:
		var cmd tea.Cmd

		m.Cursor, cmd = m.Cursor.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":
			m.SelectingFile = true
			return m, tea.Sequence(m.picker.Init(), tea.EnterAltScreen)
		}

		var cmd tea.Cmd

		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
}

func (m *FilePicker) setViewportContent(s string) {
	style := lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Center).
		Width(m.viewport.Width)

	if len(s) > m.viewport.Width {
		style = style.UnsetWidth()
	}

	m.viewport.SetContent(style.Render(s))
}

func (m *FilePicker) View() string {
	if !m.SelectingFile {
		// NOTE: Check m.SetValue(...) for details on why it's 3
		if len(strings.TrimSpace(m.viewport.View())) <= 3 {
			m.setViewportContent(" no file selected ")
		}

		return fmt.Sprintf("[%v]%v", m.viewport.View(), m.Cursor.View())
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		m.picker.View(),
		"",
		fmt.Sprintf("current path: %v", m.picker.CurrentDirectory),
		"(file picker) (up/down to navigate, enter to confirm choice, backspace to go back one directory, escape to cancel)",
	)
}

func (m *FilePicker) HasReacted() bool {
	return m.reacted
}

func (m *FilePicker) ReactFlush() {
	m.reacted = false
}

func (m *FilePicker) setReact() {
	m.reacted = true
}
