package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
)

type Checkbox struct {
	checked         bool
	CheckedSymbol   string
	UncheckedSymbol string

	Cursor  cursor.Model
	focused bool

	reacted bool
}

func NewCheckBox(checked bool) *Checkbox {
	c := cursor.New()
	c.SetChar(" ")

	return &Checkbox{
		Cursor:          c,
		checked:         checked,
		CheckedSymbol:   "x",
		UncheckedSymbol: " ",
	}
}

func (m *Checkbox) Focus() tea.Cmd {
	m.focused = true
	return m.Cursor.Focus()
}

func (m *Checkbox) Blur() {
	m.focused = false
	m.Cursor.Blur()
}

func (m *Checkbox) Toggle() {
	m.checked = !m.checked
	m.reacted = true
}

func (m *Checkbox) SetChecked() {
	m.checked = true
	m.reacted = true
}

func (m *Checkbox) SetUnchecked() {
	m.checked = false
	m.reacted = true
}

// Returns a wrapped bool
func (m *Checkbox) Value() InputValue {
	return m.checked
}

func (m *Checkbox) SetValue(checked InputValue) error {
	value, ok := checked.(bool)
	if !ok {
		return InvalidInputErr
	}

	m.checked = value
	m.reacted = true

	return nil
}

func (m *Checkbox) Reset() {
	m.SetUnchecked()
	m.reacted = true
}

func (m *Checkbox) Init() tea.Cmd {
	return nil
}

func (m *Checkbox) Update(msg tea.Msg) (Input, tea.Cmd) {
	return m.update(msg)
}

func (m *Checkbox) update(msg tea.Msg) (*Checkbox, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == " " {
			m.Toggle()
			m.reacted = true
		}
	}

	var cmd tea.Cmd
	m.Cursor, cmd = m.Cursor.Update(msg)

	return m, cmd
}

func (m *Checkbox) View() string {
	if m.checked {
		return fmt.Sprintf("[%v]%v", m.CheckedSymbol, m.Cursor.View())
	}

	return fmt.Sprintf("[%v]%v", m.UncheckedSymbol, m.Cursor.View())
}

func (m *Checkbox) HasReacted() bool {
	return m.reacted
}

func (m *Checkbox) ReactFlush() {
	m.reacted = false
}
