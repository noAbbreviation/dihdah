package decode

import (
	"fmt"
	"strings"
	"unicode"

	diacritics "github.com/Regis24GmbH/go-diacritics"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"github.com/noAbbreviation/dihdah/commons"
)

type quoteModel struct {
	drill *commons.Drill
	speed float64

	input       textarea.Model
	showResults bool

	displayedResults string
	corrects         int
	total            int

	toggleSignal chan<- struct{}
}

func newQuoteModel(quote string, speed float64) *quoteModel {
	input := textarea.New()
	input.Placeholder = "?????"
	input.MaxHeight = 5
	input.FocusedStyle = textarea.Style{
		CursorLine: lipgloss.NewStyle(),
	}
	input.Focus()

	return &quoteModel{
		drill: &commons.Drill{
			Text:    quote,
			Correct: make([]bool, len(quote)),
		},
		input: input,
		speed: speed,
	}
}

func (_m *quoteModel) Init() tea.Cmd {
	var playingCmd tea.Cmd
	playingCmd, _m.toggleSignal = initPlayingMorseCodeQuote(_m.drill.Text, _m.speed)

	return tea.Sequence(textarea.Blink, playingCmd)
}

func initPlayingMorseCodeQuote(quote string, speed float64) (tea.Cmd, chan<- struct{}) {
	cleanedQuote := strings.ToLower(diacritics.Normalize(quote))
	runes := []rune(cleanedQuote)

	firstRune := runes[0]

	morseCode := ""
	previouslySpace := false

	if firstRune >= 'a' && firstRune <= 'z' {
		morseCode += commons.MorseCodeLookup[firstRune]
	}

	for _, r := range runes[1:] {
		if r < 'a' {
			r += 'a' - 'A'
		}

		if r < 'a' || r > 'z' {
			previouslySpace = true
			continue
		}

		if previouslySpace {
			morseCode += string(commons.MorseSpaceIndicator)
		}

		previouslySpace = false
		morseCode += " " + commons.MorseCodeLookup[r]
	}

	toggleSignal := make(chan struct{}, 16)
	streamer := commons.MorseCharSound(morseCode, speed)

	quoteBuffer := beep.NewBuffer(commons.AudioFormat)
	quoteBuffer.Append(streamer)

	mixer := &beep.Mixer{}
	playingCmd := func() tea.Msg {
		speaker.Play(mixer)

		for {
			select {
			case _, ok := <-toggleSignal:
				if !ok {
					speaker.Lock()
					mixer.Clear()
					speaker.Unlock()

					return doneMsg{}
				}

				speaker.Lock()

				if mixer.Len() == 0 {
					mixer.Add(quoteBuffer.Streamer(0, quoteBuffer.Len()))
				} else {
					mixer.Clear()
				}

				speaker.Unlock()
			}
		}
	}

	return playingCmd, toggleSignal
}

func (_m *quoteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _m.showResults {
		if key, isKey := msg.(tea.KeyMsg); isKey {
			if key.String() == "esc" || key.String() == "ctrl+c" {
				return _m, tea.Quit
			}
		}
		return _m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		default:
			runes := []rune(msg.String())
			if len(runes) != 1 {
				break
			}
			char := runes[0]

			if char == ' ' {
				break
			}

			if char < 'a' {
				char += 'a' - 'A'
			}

			if char >= 'a' && char <= 'z' {
				break
			}

			return _m, nil
		case "ctrl+c", "esc":
			if len(_m.input.Value()) != 0 {
				_m.input.SetValue("")
				return _m, nil
			}

			return _m, tea.Quit
		case "ctrl+l":
			_m.toggleSignal <- struct{}{}
			return _m, nil
		case "ctrl+s":
			if _m.showResults {
				return _m, nil
			}

			_m.displayedResults, _m.corrects, _m.total = InitQuoteTrainingResults(_m.input.Value(), _m.drill.Text)
			_m.showResults = true

			close(_m.toggleSignal)

			return _m, nil
		}
	}

	var cmd tea.Cmd
	_m.input, cmd = _m.input.Update(msg)
	return _m, cmd
}

func InitQuoteTrainingResults(userAnswerStr string, realAnswerStr string) (displayedResults string, corrects int, total int) {
	userFields := strings.FieldsFunc(userAnswerStr, func(r rune) bool {
		return !unicode.IsLetter(r)
	})

	realAnswer := []rune(strings.ToLower(diacritics.Normalize(realAnswerStr)))
	userAnswer := []rune(strings.ToLower(strings.Join(userFields, " ")))

	correctionString := strings.Builder{}
	userDisplayedAnswer := strings.Builder{}

	total = 0
	corrects = 0

	userAnswerIdx := 0
	encounteredSpace := false
	extendIncorrectPadding := false

	for _, realRune := range realAnswer {
		if !unicode.IsLetter(realRune) && !encounteredSpace {
			if userAnswerIdx >= len(userAnswer) {
				correctionString.WriteRune('?')
				userDisplayedAnswer.WriteRune('_')
				continue
			}

			encounteredSpace = true
			userRune := userAnswer[userAnswerIdx]

			if userRune != ' ' {
				correctionString.WriteRune('?')
				userDisplayedAnswer.WriteRune('_')

				extendIncorrectPadding = true
			} else {
				correctionString.WriteRune(' ')
				userDisplayedAnswer.WriteRune(' ')

				userAnswerIdx += 1
			}

			continue
		}

		if realRune < 'a' || realRune > 'z' {
			if extendIncorrectPadding {
				userDisplayedAnswer.WriteRune('_')
			} else {
				userDisplayedAnswer.WriteRune(' ')
			}

			correctionString.WriteRune(' ')
			continue
		}

		encounteredSpace = false
		extendIncorrectPadding = false
		total += 1

		if userAnswerIdx >= len(userAnswer) {
			correctionString.WriteRune('?')
			userDisplayedAnswer.WriteRune('_')
			continue
		}

		userRune := userAnswer[userAnswerIdx]
		userDisplayedAnswer.WriteRune(userRune)

		if realRune == userRune {
			correctionString.WriteRune(' ')
			corrects += 1
		} else {
			correctionString.WriteRune('?')
		}

		userAnswerIdx += 1
	}

	_results := [3]string{realAnswerStr, correctionString.String(), userDisplayedAnswer.String()}
	resultsBuilder := []string{}

	maxWidth := 40
	for len(_results[0]) > maxWidth {
		resultsJoined := lipgloss.JoinVertical(
			lipgloss.Left,
			_results[0][:min(maxWidth, len(_results[0]))],
			_results[1][:min(maxWidth, len(_results[1]))],
			_results[2][:min(maxWidth, len(_results[2]))],
			strings.Repeat("-", maxWidth),
		)
		resultsBuilder = append(resultsBuilder, lipgloss.JoinHorizontal(
			lipgloss.Left,
			"   \n   \n>> ",
			resultsJoined,
		))

		_results[0] = _results[0][maxWidth:]
		_results[1] = _results[1][maxWidth:]
		_results[2] = _results[2][maxWidth:]
	}

	resultsJoined := lipgloss.JoinVertical(lipgloss.Left, _results[:]...)
	resultsBuilder = append(resultsBuilder, lipgloss.JoinHorizontal(
		lipgloss.Left,
		"   \n   \n>> ",
		resultsJoined,
	))

	return lipgloss.JoinVertical(lipgloss.Left, resultsBuilder...), corrects, total
}

func (_m *quoteModel) View() string {
	if _m.showResults {
		scoreText := "(all correct!)"
		if _m.corrects != _m.total {
			mistakes := _m.total - _m.corrects
			scoreText = fmt.Sprintf("(%v/%v mistakes)", mistakes, _m.total)
		}

		return lipgloss.JoinVertical(
			lipgloss.Left,
			"Decode quote training results",
			"",
			_m.displayedResults,
			"",
			fmt.Sprintf("%v (ctrl+c/esc to go back)", scoreText),
			"",
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		"Decode quote training",
		"",
		_m.input.View(),
		"",
		"(ctrl+l to stop/restart playing, ctrl+s to confirm answer, ctrl+c/esc to clear or go back)",
		"",
	)
}
