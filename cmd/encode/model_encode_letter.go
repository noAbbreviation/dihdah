package encode

import (
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"github.com/noAbbreviation/dihdah/commons"
)

type letterModel struct {
	drill       *commons.Drill
	lettersUsed string

	input        textinput.Model
	resultsTable table.Model

	showResults bool
	score       int

	charPlayer chan<- rune
}

func newLetterModel(trainingLetters string) *letterModel {
	drills := &commons.Drill{
		Text:    trainingLetters,
		Correct: make([]bool, len(trainingLetters)),
	}

	input := textinput.New()
	input.CharLimit = 32
	input.Width = 10
	input.Placeholder = "????"
	input.Focus()

	return &letterModel{
		drill:       drills,
		input:       input,
		lettersUsed: trainingLetters,
	}
}

type doneMsg struct{}

func initPlayingMorseCode(speed float64) (tea.Cmd, chan<- rune) {
	delayBuffer := commons.SoundAssets[commons.ShortDelay]
	delayStreamer := delayBuffer.Streamer(0, delayBuffer.Len())
	delayResampler := beep.ResampleRatio(4, speed, delayStreamer)

	emptyStreamer := beep.NewBuffer(commons.AudioFormat)
	emptyStreamer.Append(delayResampler)

	playing := sync.WaitGroup{}
	chars := make(chan rune, 256)

	playingCmd := func() tea.Msg {
		for {
			var c rune
			var ok bool

			select {
			case c, ok = <-chars:
				playing.Wait()
				playing.Add(1)

				streamer := emptyStreamer.Streamer(0, emptyStreamer.Len())
				if c == ' ' {
					speaker.Play(
						beep.Seq(
							beep.Loop(7, streamer),
							beep.Callback(playing.Done),
						),
					)

					continue
				}

				morseCode := commons.MorseCodeLookup[c]
				speaker.Play(
					beep.Seq(
						commons.MorseCharSound(morseCode, speed),
						streamer,
						beep.Callback(playing.Done),
					),
				)
			}

			if !ok && len(chars) == 0 {
				return doneMsg{}
			}
		}
	}

	return playingCmd, chars
}

func (_m *letterModel) Init() tea.Cmd {
	var playingCmd tea.Cmd
	playingCmd, _m.charPlayer = initPlayingMorseCode(1)

	return tea.Batch(playingCmd, textinput.Blink)
}

type quitMsg struct{}

func (_m *letterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	drill := _m.drill

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return _m, tea.Quit
		}
	}

	if _m.showResults {
		if key, isKey := msg.(tea.KeyMsg); isKey {
			if key.String() == "enter" {
				return _m, tea.Quit
			}
		}

		var cmd tea.Cmd
		_m.resultsTable, cmd = _m.resultsTable.Update(msg)
		return _m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		default:
			keyMsg := msg.Runes
			if len(keyMsg) != 1 {
				break
			}

			if keyMsg[0] == '.' || keyMsg[0] == ',' {
				break
			}

			return _m, nil

		case "enter":
			if drill.Current >= len(drill.Text) {
				_m.showResults = true
				return _m, nil
			}

			currentChar := drill.Text[drill.Current]

			userAnswer := _m.input.Value()
			morseCodeAnswer := commons.MorseCodeLookup[rune(currentChar)]

			if userAnswer == morseCodeAnswer {
				drill.Correct[drill.Current] = true
				_m.charPlayer <- rune(currentChar)
			}

			drill.Current += 1
			for drill.Current < len(drill.Text) {
				currentChar := drill.Text[drill.Current]
				if currentChar >= 'a' && currentChar <= 'z' {
					break
				}

				drill.Current += 1
			}

			if drill.Current >= len(drill.Text) {
				close(_m.charPlayer)

				_m.resultsTable = _m.initResultsTable()
				_m.score, _ = countCorrectLetters(drill.Text, drill.Correct)
				_m.showResults = true
			}

			_m.input.Reset()
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

func (_m letterModel) initResultsTable() table.Model {
	drill := _m.drill

	j := 1
	rows := []table.Row{}

	for i := 0; i < len(drill.Text); i++ {
		currentChar := rune(drill.Text[i])
		if currentChar < 'a' || currentChar > 'z' {
			continue
		}

		correctString := "yes"
		if !drill.Correct[i] {
			correctString = "no"
		}

		row := table.Row{
			fmt.Sprint(j),
			string(currentChar),
			correctString,
			commons.MorseCodeLookup[currentChar],
		}

		rows = append(rows, row)
		j += 1
	}

	columns := []table.Column{
		{Title: "#", Width: 3},
		{Title: "Character", Width: 10},
		{Title: "Correct?", Width: 8},
		{Title: "Answer", Width: 7},
	}

	tableStyle := table.DefaultStyles()

	return table.New(
		table.WithFocused(true),
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(min(10, len(rows)+1)),
		table.WithStyles(tableStyle),
	)
}

func countCorrectLetters(text string, correct []bool) (int, error) {
	if len([]rune(text)) != len(correct) {
		return -1, fmt.Errorf("Corrects slice is not equal to length of text.")
	}

	correctCount := 0
	for i := 0; i < len(text); i++ {
		currentChar := rune(text[i])
		if currentChar < 'a' || currentChar > 'z' {
			continue
		}

		if correct[i] {
			correctCount += 1
		}
	}

	return correctCount, nil
}

func (_m *letterModel) View() string {
	drill := _m.drill
	if _m.showResults {
		iterations := len(drill.Text)

		scoreText := "(all correct!)"
		if _m.score != iterations {
			mistakes := len(drill.Text) - _m.score
			scoreText = fmt.Sprintf("(%v/%v mistakes)", mistakes, iterations)
		}

		return lipgloss.JoinVertical(
			lipgloss.Left,
			fmt.Sprintf(
				"Encode training results (%v letters, %v iterations):",
				len(_m.lettersUsed),
				len(drill.Text),
			),
			"",
			_m.resultsTable.View(),
			"",
			fmt.Sprintf("%v (escape / ctrl+c / enter to go back)", scoreText),
			"",
		)
	}

	var charView string
	if drill.Current < len(drill.Text) {
		charView = string(drill.Text[drill.Current])
	} else {
		charView = "done"
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		fmt.Sprintf("Encode training (%v letters)", len(_m.lettersUsed)),
		fmt.Sprintf("Letter '%v' (%v of %v)", charView, drill.Current+1, len(drill.Text)),
		_m.input.View(),
		"",
		"(escape/ctrl+c to go back, enter to confirm)",
		"",
	)
}
