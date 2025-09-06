package decode

import (
	"fmt"
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
	speed       float64

	input        textinput.Model
	resultsTable table.Model

	userAnswers []rune
	showResults bool
	score       int

	charPlayer   chan<- rune
	replaySignal chan<- struct{}
}

func newLetterModel(trainingLetters string, lettersUsed string, speed float64) *letterModel {
	drills := &commons.Drill{
		Text:    trainingLetters,
		Correct: make([]bool, len(trainingLetters)),
	}

	input := textinput.New()
	input.CharLimit = 1
	input.Width = 4
	input.Placeholder = "????"
	input.Focus()

	return &letterModel{
		drill:       drills,
		input:       input,
		lettersUsed: lettersUsed,
		userAnswers: make([]rune, len(trainingLetters)),
		speed:       speed,
	}
}

type doneMsg struct{}

func initPlayingMorseCode(speed float64) (tea.Cmd, chan<- rune, chan<- struct{}) {
	delayBuffer := commons.SoundAssets[commons.ShortDelay]
	delayStreamer := delayBuffer.Streamer(0, delayBuffer.Len())
	delayResampler := beep.ResampleRatio(4, speed, delayStreamer)

	emptyStreamer := beep.NewBuffer(commons.AudioFormat)
	emptyStreamer.Append(delayResampler)

	replaySignal := make(chan struct{}, 16)
	newChar := make(chan rune, 16)

	mixer := &beep.Mixer{}
	var currentSound *beep.Buffer

	playingCmd := func() tea.Msg {
		speaker.Play(mixer)

		for {
			select {
			case c, ok := <-newChar:
				if !ok {
					return doneMsg{}
				}

				morseCode := commons.MorseCodeLookup[c]
				currentStreamer := commons.MorseCharSound(morseCode, speed)

				currentSound = beep.NewBuffer(commons.AudioFormat)
				currentSound.Append(currentStreamer)

				continue
			default:
			}

			select {
			case <-replaySignal:
				speaker.Lock()

				mixer.Clear()
				mixer.Add(currentSound.Streamer(0, currentSound.Len()))

				speaker.Unlock()

				continue
			default:
			}
		}
	}

	return playingCmd, newChar, replaySignal
}

func (_m *letterModel) Init() tea.Cmd {
	var playingCmd tea.Cmd
	playingCmd, _m.charPlayer, _m.replaySignal = initPlayingMorseCode(_m.speed)

	_m.charPlayer <- rune(_m.drill.Text[_m.drill.Current])
	_m.replaySignal <- struct{}{}

	return tea.Batch(textinput.Blink, playingCmd)
}

type quitMsg struct{}
type replayMsg struct{}

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
			if drill.Current >= len(drill.Text) {
				_m.showResults = true
				return _m, nil
			}

			userAnswer := _m.input.Value()
			currentChar := drill.Text[drill.Current]

			if len(userAnswer) == 0 {
				_m.replaySignal <- struct{}{}
				return _m, nil
			}

			_m.userAnswers[drill.Current] = rune(userAnswer[0])
			if userAnswer == string(currentChar) {
				drill.Correct[drill.Current] = true
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

				return _m, nil
			}

			_m.input.Reset()
			_m.charPlayer <- rune(drill.Text[drill.Current])
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
			string(_m.userAnswers[i]),
		}

		rows = append(rows, row)
		j += 1
	}

	columns := []table.Column{
		{Title: "#", Width: 3},
		{Title: "Character", Width: 10},
		{Title: "Correct?", Width: 8},
		{Title: "Answered", Width: 8},
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
				"Decoding training results (%v letters, %v iterations):",
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

	return lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		fmt.Sprintf(
			"Decode training (%v letters) (%v of %v)",
			len(_m.lettersUsed),
			drill.Current+1,
			len(drill.Text),
		),
		_m.input.View(),
		"",
		"(escape/ctrl+c to go back, space to repeat sound, enter to confirm)",
		"",
	)
}
