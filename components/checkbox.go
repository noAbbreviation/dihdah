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
}

func (m *Checkbox) SetChecked() {
	m.checked = true
}

func (m *Checkbox) SetUnchecked() {
	m.checked = false
}

func (m *Checkbox) Value() bool {
	return m.checked
}

func (m *Checkbox) Reset() {
	m.SetUnchecked()
}

func (m *Checkbox) Update(msg tea.Msg) (*Checkbox, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ", "enter":
			m.Toggle()
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
