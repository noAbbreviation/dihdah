package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
)

const float64_epsilon = 1.11e-16

type Number struct {
	min float64
	max float64

	value float64

	Default float64
	delta   float64

	DecrementSymbol rune
	IncrementSymbol rune

	Cursor  cursor.Model
	focused bool
}

// Copied from the golang standard library "math" package
// isNaN reports whether f is an IEEE 754 “not-a-number” value.
func isNaN(f float64) (is bool) {
	// IEEE 754 says that only NaNs satisfy f != f.
	return f != f
}

func abs(f float64) float64 {
	if f > 0 {
		return f
	}

	return -f
}

func NewNumber(minValue float64, maxValue float64) *Number {
	cursor := cursor.New()
	cursor.SetChar(" ")

	if isNaN(minValue) {
		minValue = 0
	}

	if isNaN(maxValue) {
		maxValue = 0
	}

	if minValue > maxValue {
		minValue, maxValue = maxValue, minValue
	}

	return &Number{
		min:             minValue,
		max:             maxValue,
		Default:         minValue,
		value:           minValue,
		delta:           1,
		DecrementSymbol: '<',
		IncrementSymbol: '>',
		Cursor:          cursor,
	}
}

// Sets to default value
func (m *Number) Reset() {
	m.SetValue(m.Default)
}

// Returns a wrapped float64
func (m Number) Value() InputValue {
	return m.value
}

func (m *Number) SetValue(value InputValue) error {
	num, ok := value.(float64)
	if !ok {
		return InvalidInputErr
	}

	m.value = num
	m.ClampValue()

	return nil
}

func (m *Number) ClampValue() {
	m.value = clamp(m.value, m.max, m.min)
}

func clamp(value float64, _max float64, _min float64) float64 {
	if _max < _min {
		_min, _max = _max, _min
	}

	value = min(_max, value)
	if abs(_max-value) < float64_epsilon {
		value = _max
		return value
	}

	value = max(_min, value)
	if abs(_min-value) < float64_epsilon {
		value = _min
	}

	return value
}

func (m *Number) Increment() {
	if !m.focused {
		return
	}

	m.value += m.delta
	m.ClampValue()
}

func (m *Number) Decrement() {
	if !m.focused {
		return
	}

	m.value -= m.delta
	m.ClampValue()
}

func (m *Number) SetDelta(delta float64) {
	if isNaN(delta) {
		delta = 0
		return
	}

	if delta > 0 {
		m.delta = delta
		return
	}

	m.delta = -delta
}

func (m *Number) Blur() {
	m.focused = false
	m.Cursor.Blur()
}

func (m *Number) Focus() tea.Cmd {
	m.focused = true
	return m.Cursor.Focus()
}

func (m *Number) Init() tea.Cmd {
	m.Reset()
	return nil
}

func (m *Number) Update(msg tea.Msg) (Input, tea.Cmd) {
	return m.update(msg)
}

func (m *Number) update(msg tea.Msg) (*Number, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "+":
			m.Increment()
		case "-":
			m.Decrement()
		}

	case tea.MouseMsg:
		if msg.Action != tea.MouseActionPress {
			break
		}

		switch msg.Button {
		case tea.MouseButtonWheelUp:
			m.Increment()
		case tea.MouseButtonWheelDown:
			m.Decrement()
		}
	}

	var cmd tea.Cmd
	m.Cursor, cmd = m.Cursor.Update(msg)

	return m, cmd
}

func (m Number) View() string {
	viewStr := ""

	if m.min < m.value {
		viewStr += string(m.DecrementSymbol)
	} else {
		viewStr += " "
	}

	viewStr += fmt.Sprintf("  %v%v ", m.value, m.Cursor.View())

	if m.max > m.value {
		viewStr += string(m.IncrementSymbol)
	} else {
		viewStr += " "
	}

	return viewStr
}
