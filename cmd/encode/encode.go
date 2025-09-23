package encode

import (
	"fmt"
	"math/rand"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var NewLettersPerLevel = [...]string{
	"the",
	"dog",
	"brown",
	"jumps",
	"foxover", // fox over
	"quick",
	"lazy",
}

func init() {
	Cmd.Flags().UintP("iterations", "n", 0, "How many items for the training session.")
	Cmd.Flags().BoolP("recap", "a", false, "To train for all letters in the letter pool at once.")

	Cmd.Flags().Uint16P("level", "l", 0, fmt.Sprintf(
		"Level to use for training. Each level adds 3-5 new letters for training. Max level: %v",
		len(NewLettersPerLevel),
	))
	Cmd.Flags().String("letters", "", "Custom alphabet pool to train. You probably should start by using --level.")

	Cmd.MarkFlagsOneRequired("level", "letters")
	Cmd.MarkFlagsMutuallyExclusive("level", "letters")
}

var Cmd = &cobra.Command{
	Use:   "encode",
	Short: "Drills for encoding the morse code alphabet.",
	RunE: func(cmd *cobra.Command, args []string) error {
		detectedNonAlphabet := false
		letters, _ := cmd.Flags().GetString("letters")

		letters = strings.ToLower(letters)
		letters = strings.Map(func(r rune) rune {
			if r >= 'A' && r <= 'Z' {
				return r + ('a' - 'A')
			}

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

			if int(levelArg) > len(NewLettersPerLevel) {
				cmd.PrintErrf("Warning: Level is at most %v. Will be set to max.\n", len(NewLettersPerLevel))
				levelArg = uint16(len(NewLettersPerLevel))
			}

			if levelArg == 0 {
				return fmt.Errorf("Error: --letters is empty.")
			}

			for i := range levelArg {
				letters += NewLettersPerLevel[i]
			}
		}

		dedupedLetters := DedupCleanLetters(letters)

		doAllLetters, _ := cmd.Flags().GetBool("recap")
		if doAllLetters {
			p := tea.NewProgram(NewLetterModel(dedupedLetters, nil))
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("Error running the program: %v", err)
			}

			return nil
		}

		iterations, _ := cmd.Flags().GetUint("iterations")
		if iterations == 0 {
			iterations = max(uint(len(dedupedLetters)/2), 3)
		}

		trainingLetters := ""
		for range iterations {
			randomLetter := letters[rand.Intn(len(letters))]
			trainingLetters += string(randomLetter)
		}

		p := tea.NewProgram(NewLetterModel(trainingLetters, nil))

		if _, err := p.Run(); err != nil {
			return fmt.Errorf("Error running the program: %v", err)
		}

		return nil
	},
	Long: `The encode command gives the user drills to internalize the morse code alphabet.
The flags in this command should be self-explanatory.

# How it works

For each item, you will be given a letter to encode to morse code. Input the
corresponding morse code using commas{,} as dashes and periods{.} as dots.
Enter to confirm the answer.

=============================================
Encode training (3 letters)
Letter 'h' (1 of 3)
> ....

(escape/ctrl+c to go back, enter to confirm)
=============================================

At the end of the training session, you will be presented with the correct morse
code characters. Use that as learning and/or feedback for the next training.

====================================================
Encode training results (3 letters, 3 iterations):

 #    Character   Correct?  Answer
 1    h           yes       ....
 2    t           yes       ,
 3    t           yes       ,

(all correct!) (escape / ctrl+c / enter to go back)
====================================================

# Extras

The 'encode' command is analogous to 'decode letters', but since the 'encode' command only has
letter drills, it was decided to only have 'encode'.

This is the default letter pool if you specify --level/-l:
    ========================================
    | Level | Letter Pool                  |
    | ----- | ---------------------------- |
    | 1     | the                          |
    | 2     | thedog                       |
    | 3     | thedogbrown                  |
    | 4     | thedogbrownjumps             |
    | 5     | thedogbrownjumpsfoxover      |
    | 6     | thedogbrownjumpsfoxoverquick |
    | 7+    | (everything)                 |
    ========================================

NOTES:
  - If the user is having difficulty differentiating letters, it is recommended
    to run this command with --letters.
  - After being comfortable with a certain --level, it is also recommended to
    run --level with --recap before proceeding with the next --level.`,
}

func DedupCleanLetters(str string) string {
	runes := []rune(strings.ToLower(str))
	firstLetter := runes[0]

	letters := string(firstLetter)

	for _, rune := range runes[1:] {
		if strings.Contains(letters, string(rune)) {
			continue
		}

		if rune >= 'a' && rune <= 'z' {
			letters += string(rune)
		}
	}

	return letters
}
