package decode

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
	LetterCmd.Flags().UintP("iterations", "n", 0, "Training iterations.")
	LetterCmd.Flags().Float64P("speed", "s", 1, "Speed ratio to train with.")
	LetterCmd.Flags().BoolP("recap", "a", false, "To train for all letters (in the level if applicable).")

	LetterCmd.Flags().Uint16P("level", "l", 0, fmt.Sprintf(
		"Level to have for training. Each level adds 3-5 new letters to train. Max level: %v",
		len(NewLettersPerLevel),
	))
	LetterCmd.Flags().String("letters", "", "Custom alphabet pool to train. You probably should start by using --level.")

	LetterCmd.MarkFlagsOneRequired("level", "letters")
	LetterCmd.MarkFlagsMutuallyExclusive("level", "letters")
}

var LetterCmd = &cobra.Command{
	Use:     "letter",
	Short:   "Train for decoding letters.",
	Aliases: []string{"letters"},
	RunE: func(cmd *cobra.Command, args []string) error {
		detectedNonAlphabet := false
		letters, _ := cmd.Flags().GetString("letters")
		if len(letters) != 0 {
			letters = DedupCleanLetters(letters)
		}

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
		speed, _ := cmd.Flags().GetFloat64("speed")

		doAllLetters, _ := cmd.Flags().GetBool("recap")
		if doAllLetters {
			allLettersRand := []rune(dedupedLetters)
			rand.Shuffle(len(allLettersRand), func(i, j int) {
				allLettersRand[i], allLettersRand[j] = allLettersRand[j], allLettersRand[i]
			})

			p := tea.NewProgram(NewLetterModel(string(allLettersRand), dedupedLetters, speed, nil))
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

		p := tea.NewProgram(NewLetterModel(trainingLetters, dedupedLetters, speed, nil))
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("Error running the program: %v", err)
		}

		return nil
	},
	Long: `The 'decode letters' command gives the user drills to decode the morse code alphabet.
The flags in this command should be self-explanatory.

# How it works

For each item, you will be given a sound clip to listen. Input the letter corresponding
to the sound. Pressing space or hitting enter when empty will repeat the sound.
Enter to confirm the answer.

====================================================================
Decode training (3 letters) (1 of 3)
> t

(escape/ctrl+c to go back, space to repeat sound, enter to confirm)
====================================================================


At the end of the training session, you will be presented with the correct characters
played. Use that as learning and/or feedback for the next training.

=====================================================
Decoding training results (3 letters, 3 iterations):

 #    Character   Correct?  Answered
 1    t           yes       t
 2    e           yes       e
 3    e           yes       e

(all correct!) (escape / ctrl+c / enter to go back)
=====================================================

# Extras

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
    run --level with --recap before proceeding with the next --level.
  - For the convenience and the challenge for the user, --speed can be used to
	slow down or speed up the sound being played.`,
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
