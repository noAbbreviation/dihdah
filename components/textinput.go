package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type TextInput struct {
	Input            textinput.Model
	InvalidHighlight string

	reacted bool
}

func (m *TextInput) Init() tea.Cmd {
	return nil
}

func (m *TextInput) Update(msg tea.Msg) (Input, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if len(msg.Runes) == 1 {
			m.reacted = true
		}
	}

	m.Input, cmd = m.Input.Update(msg)
	return m, cmd
}
func (m *TextInput) View() string {
	invalidHighlight := ""
	if m.Input.Err != nil {
		invalidHighlight = m.InvalidHighlight
	}

	return fmt.Sprintf("%v %v %v", invalidHighlight, m.Input.View(), invalidHighlight)
}

// Returns a wrapped string
func (m *TextInput) Value() InputValue {
	return m.Input.Value()
}

func (m *TextInput) SetValue(str InputValue) error {
	value, ok := str.(string)
	if !ok {
		return InvalidInputErr
	}

	m.Input.SetValue(value)
	m.reacted = true
	return nil
}

func (m *TextInput) Reset() {
	m.Input.SetValue("")
	m.Input.SetCursor(0)
}

func (m *TextInput) Focus() tea.Cmd {
	return m.Input.Focus()
}

func (m *TextInput) Blur() {
	m.Input.Blur()
}

func (m *TextInput) HasReacted() bool {
	return m.reacted
}

func (m *TextInput) ReactFlush() {
	m.reacted = false
}
