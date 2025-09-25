package decode

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"github.com/noAbbreviation/dihdah/commons"
)

type wordModel struct {
	backReference    tea.Model
	wrongRightSorted bool

	drills  *commons.TrainingModel
	speed   float64
	wordLen uint16

	input        textinput.Model
	resultsTable table.Model
	rows         [][3]table.Row

	userAnswers []string
	showResults bool
	score       int

	wordPlayer   chan<- string
	replaySignal chan<- struct{}
	killSignal   chan<- struct{}
}

func NewWordModel(words []string, wordLen uint16, speed float64, backReference tea.Model) *wordModel {
	drills := []commons.Drill{}

	for _, word := range words {
		drill := commons.Drill{Text: word}
		drills = append(drills, drill)
	}

	input := textinput.New()
	input.CharLimit = 20
	input.Width = 25
	input.Placeholder = "??????????"
	input.Focus()

	return &wordModel{
		backReference: backReference,
		drills: &commons.TrainingModel{
			Drills:  drills,
			Correct: make([]bool, len(drills)),
		},
		input:       input,
		userAnswers: make([]string, len(words)),
		speed:       speed,
		wordLen:     wordLen,
	}
}

func initPlayingMorseCodeWords(speed float64) (
	cmd tea.Cmd,
	wordSignal chan<- string,
	replaySignal chan<- struct{},
	killSignal chan<- struct{},
) {
	_replaySignal := make(chan struct{}, 16)
	newWord := make(chan string, 16)
	_killSignal := make(chan struct{}, 16)

	mixer := &beep.Mixer{}
	var currentSound *beep.Buffer

	playingCmd := func() tea.Msg {
		speaker.Play(mixer)

		for {
			select {
			case <-_killSignal:
				speaker.Lock()
				mixer.Clear()
				speaker.Unlock()

				return doneMsg{}
			default:
			}

			select {
			case word, ok := <-newWord:
				if !ok {
					speaker.Lock()
					mixer.Clear()
					speaker.Unlock()

					return doneMsg{}
				}

				runes := []rune(word)

				firstRune := runes[0]
				if firstRune < 'a' {
					firstRune += 'a' - 'A'
				}

				morseCode := ""
				if firstRune >= 'a' && firstRune <= 'z' {
					morseCode += commons.MorseCodeLookup[firstRune]
				}

				for _, r := range runes[1:] {
					if r == '-' {
						morseCode += "-"
						continue
					}

					if r < 'a' {
						r += 'a' - 'A'
					}

					if r >= 'a' && r <= 'z' {
						morseCode += " " + commons.MorseCodeLookup[r]
					}
				}

				currentStreamer := commons.MorseCharSound(morseCode, speed)
				currentSound = beep.NewBuffer(commons.AudioFormat)
				currentSound.Append(currentStreamer)

				continue
			default:
			}

			select {
			case <-_replaySignal:
				speaker.Lock()

				mixer.Clear()
				mixer.Add(currentSound.Streamer(0, currentSound.Len()))

				speaker.Unlock()

				continue
			default:
			}
		}
	}

	return playingCmd, newWord, _replaySignal, _killSignal
}

func (_m *wordModel) Init() tea.Cmd {
	var playingCmd tea.Cmd
	playingCmd, _m.wordPlayer, _m.replaySignal, _m.killSignal = initPlayingMorseCodeWords(_m.speed)

	_m.wordPlayer <- _m.drills.Drills[_m.drills.CurrentDrill].Text
	_m.replaySignal <- struct{}{}

	return tea.Batch(textinput.Blink, playingCmd)
}

func (_m *wordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	drills := _m.drills

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if _m.backReference == nil {
				return _m, tea.Quit
			}

			_m.killSignal <- struct{}{}
			return _m.backReference, nil
		case "ctrl+c":
			return _m, tea.Quit
		}
	}

	if _m.showResults {
		if key, isKey := msg.(tea.KeyMsg); isKey {
			switch key.String() {
			case "enter", "esc":
				if _m.backReference == nil {
					return _m, tea.Quit
				}

				_m.killSignal <- struct{}{}
				return _m.backReference, nil
			case "s":
				_m.resultsTable = _m.toggleSorted()
			}
		}

		var cmd tea.Cmd
		_m.resultsTable, cmd = _m.resultsTable.Update(msg)
		return _m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ":
			_m.replaySignal <- struct{}{}
			return _m, nil
		default:
			keyMsg := msg.Runes
			if len(keyMsg) != 1 {
				break
			}

			if keyMsg[0] >= 'a' && keyMsg[0] <= 'z' {
				break
			}

			return _m, nil

		case "enter":
			if drills.CurrentDrill >= len(drills.Drills) {
				_m.showResults = true
				return _m, nil
			}

			userAnswer := _m.input.Value()
			currentWord := drills.Drills[drills.CurrentDrill].Text

			if len(userAnswer) == 0 {
				_m.replaySignal <- struct{}{}
				return _m, nil
			}

			_m.userAnswers[drills.CurrentDrill] = userAnswer
			if userAnswer == string(currentWord) {
				drills.Correct[drills.CurrentDrill] = true
			}

			drills.CurrentDrill += 1

			if drills.CurrentDrill >= len(drills.Drills) {
				close(_m.wordPlayer)
				correctWords := 0
				for _, correctAnswer := range _m.drills.Correct {
					if correctAnswer {
						correctWords += 1
					}
				}

				_m.rows = _m.initResultsTable()
				_m.wrongRightSorted = true
				_m.resultsTable = _m.toggleSorted()

				_m.score = correctWords
				_m.showResults = true

				return _m, nil
			}

			_m.input.Reset()
			_m.wordPlayer <- drills.Drills[drills.CurrentDrill].Text
			_m.replaySignal <- struct{}{}

			return _m, nil
		}

	case doneMsg:
		return _m, tea.Tick(time.Second*3, func(_ time.Time) tea.Msg {
			return quitMsg{}
		})

	case quitMsg:
		return _m, tea.Quit
	}

	var cmd tea.Cmd
	_m.input, cmd = _m.input.Update(msg)

	return _m, cmd
}

func (_m wordModel) initResultsTable() [][3]table.Row {
	drills := _m.drills

	rows := [][3]table.Row{}
	maxWordWidth := 4
	maxUserWordWidth := 5

	for i, drill := range drills.Drills {
		maxWordWidth = max(maxWordWidth, len(drill.Text))
		maxUserWordWidth = max(maxUserWordWidth, len(_m.userAnswers[i]))

		correctString := "yes"
		if !drills.Correct[i] {
			correctString = "no"
		}

		userAnswer := _m.userAnswers[i]
		realAnswer := []rune(drill.Text)

		correctionString := ""
		for i, userRune := range userAnswer {
			if i >= len(realAnswer) {
				correctionString += "+"
				continue
			}

			realRune := realAnswer[i]
			if realRune == userRune {
				correctionString += " "
			} else {
				correctionString += "?"
			}
		}

		userDisplayedAnswer := userAnswer
		missingLetters := len(drill.Text) - len(userAnswer)

		if missingLetters > 0 {
			userDisplayedAnswer += strings.Repeat("_", missingLetters)
			correctionString += strings.Repeat("?", missingLetters)
		}

		firstRow := table.Row{
			fmt.Sprint(i + 1),
			drill.Text,
			correctString,
			userDisplayedAnswer,
		}

		correctionStringRow := table.Row{
			"",
			"",
			"",
			correctionString,
		}
		paddingRow := table.Row{"", "", "", ""}

		groupedRows := [3]table.Row{
			firstRow,
			correctionStringRow,
			paddingRow,
		}
		rows = append(rows, groupedRows)
	}

	wordResultsColumns[wordWidthIdx].Width = maxWordWidth
	wordResultsColumns[inputStrIdx].Width = maxWordWidth

	return rows
}

const (
	itemNumberIdx int = iota
	wordWidthIdx
	correctStrIdx
	inputStrIdx
)

var wordResultsColumns = []table.Column{
	{Title: "#", Width: 3},
	{Title: "Word"}, // Width: maxWordWidth
	{Title: "Correct?", Width: 8},
	{Title: "Input"}, // Width: maxWordWidth
}

func word_compareCorrectsThenNums(rowA, rowB [3]table.Row) int {
	correctStrIdx := 2
	if rowA[0][correctStrIdx] != rowB[0][correctStrIdx] {
		if rowA[0][correctStrIdx] == "no" {
			return -1
		} else {
			return 1
		}
	}

	return word_compareRowNums(rowA, rowB)
}

func word_compareRowNums(rowA, rowB [3]table.Row) int {
	itemNumberIdx := 0
	numA, _ := strconv.Atoi(rowA[0][itemNumberIdx])
	numB, _ := strconv.Atoi(rowB[0][itemNumberIdx])

	if numA < numB {
		return -1
	} else {
		return 1
	}
}

func (_m *wordModel) toggleSorted() table.Model {
	_m.wrongRightSorted = !_m.wrongRightSorted

	if _m.wrongRightSorted {
		slices.SortFunc(_m.rows, word_compareCorrectsThenNums)
	} else {
		slices.SortFunc(_m.rows, word_compareRowNums)
	}

	rows := []table.Row{}
	for _, r := range _m.rows {
		rows = append(rows, r[:]...)
	}

	return table.New(
		table.WithFocused(true),
		table.WithColumns(wordResultsColumns),
		table.WithRows(rows),
		table.WithHeight(min(10, len(_m.rows)+1)),
	)
}

func (_m *wordModel) View() string {
	drills := _m.drills

	trainingSpecification := fmt.Sprintf("%v letter limit", _m.wordLen)
	if _m.wordLen == 0 {
		trainingSpecification = "custom word pool"
	}

	if _m.showResults {
		iterations := len(drills.Drills)

		scoreText := "(all correct!)"
		if _m.score != iterations {
			mistakes := len(drills.Drills) - _m.score
			scoreText = fmt.Sprintf("(%v/%v mistakes)", mistakes, iterations)
		}

		return lipgloss.JoinVertical(
			lipgloss.Left,
			fmt.Sprintf(
				"Decode word training results (%v, %v iterations):",
				trainingSpecification,
				len(drills.Drills),
			),
			"",
			_m.resultsTable.View(),
			"",
			fmt.Sprintf("%v (escape/enter to go back, s to toggle sort, ctrl+c to exit)", scoreText),
			"",
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		fmt.Sprintf(
			"Decode word training (%v) (%v of %v)",
			trainingSpecification,
			drills.CurrentDrill+1,
			len(drills.Drills),
		),
		_m.input.View(),
		"",
		"(escape to go back, space to repeat sound, enter to confirm, ctrl+c to exit)",
		"",
	)
}
