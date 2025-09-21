package components

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"
)

var InvalidInputErr = errors.New("Not a valid input.")

type Input interface {
	Update(tea.Msg) (Input, tea.Cmd)
	View() string
	Init() tea.Cmd

	Focus() tea.Cmd
	Blur()

	Value() InputValue
	SetValue(InputValue) error
	Reset()
}

type InputValue interface{}
