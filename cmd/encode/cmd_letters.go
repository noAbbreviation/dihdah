package encode

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var newLettersPerLevel = []string{
	"the",
	"dog",
	"brown",
	"jumps",
	"foxover", // fox over
	"quick",
	"lazy",
}

var LetterCmd = &cobra.Command{
	Use:     "letter",
	Short:   "Train for decoding letters.",
	Aliases: []string{"letters"},
	RunE: func(cmd *cobra.Command, args []string) error {
		detectedNonAlphabet := false
		letters, _ := cmd.Flags().GetString("letters")

		letters = strings.ToLower(letters)
		letters = strings.Map(func(r rune) rune {
			if r >= 'a' && r <= 'z' {
				return r
			}

			detectedNonAlphabet = true
			return -1
		}, letters)

		if detectedNonAlphabet {
			if len(letters) == 0 {
				return fmt.Errorf("Error: --letters has effectively nothing in it.")
			} else {
				cmd.PrintErrln("Warning: Detected non-alphabet characters in --letters, removing...")
			}
		}

		if len(letters) == 0 {
			levelArg, _ := cmd.Flags().GetUint16("level")

			if int(levelArg) > len(newLettersPerLevel) {
				cmd.PrintErrf("Warning: Level is at most %v. Will be set to max.\n", len(newLettersPerLevel))
				levelArg = uint16(len(newLettersPerLevel))
			}

			if levelArg == 0 {
				return fmt.Errorf("Error: --letters is empty.")
			}

			for i := range levelArg {
				letters += newLettersPerLevel[i]
			}
		}

		dedupedLetters := dedupLetters(letters)

		doAllLetters, _ := cmd.Flags().GetBool("recap")
		if doAllLetters {
			p := tea.NewProgram(newAllLetterModel(dedupedLetters))
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("Error running the program: %v", err)
			}

			return nil
		}

		iterations, _ := cmd.Flags().GetUint("iterations")
		if iterations == 0 {
			iterations = max(uint(len(dedupedLetters)/2), 3)
		}
		p := tea.NewProgram(newLetterModel(dedupedLetters, iterations))

		if _, err := p.Run(); err != nil {
			return fmt.Errorf("Error running the program: %v", err)
		}

		return nil
	},
}

func dedupLetters(str string) string {
	runes := []rune(str)
	firstLetter := runes[0]

	letters := string(firstLetter)

	for _, rune := range runes[1:] {
		if strings.Contains(letters, string(rune)) {
			continue
		}

		letters += string(rune)
	}

	return letters
}

func init() {
	LetterCmd.Flags().UintP("iterations", "n", 0, "Training iterations.")
	LetterCmd.Flags().BoolP("recap", "a", false, "To train for all letters (in the level if applicable).")

	LetterCmd.Flags().Uint16P("level", "l", 0, fmt.Sprintf(
		"Level to have for training. Each level adds 3-5 new letters to train. Max level: %v",
		len(newLettersPerLevel),
	))
	LetterCmd.Flags().String("letters", "", "Custom alphabet pool to train. You probably should start by using --level.")

	LetterCmd.MarkFlagsOneRequired("level", "letters")
}
