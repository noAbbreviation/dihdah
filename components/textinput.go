package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type TextInput struct {
	Input textinput.Model
}

func (m *TextInput) Init() tea.Cmd {
	return nil
}

func (m *TextInput) Update(msg tea.Msg) (Input, tea.Cmd) {
	var cmd tea.Cmd

	m.Input, cmd = m.Input.Update(msg)
	return m, cmd
}
func (m *TextInput) View() string {
	return m.Input.View()
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
	return nil
}

func (m *TextInput) Reset() {
	m.Reset()
}

func (m *TextInput) Focus() tea.Cmd {
	return m.Input.Focus()
}

func (m *TextInput) Blur() {
	m.Input.Blur()
}
