package decode

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/noAbbreviation/dihdah/assets"
	"github.com/spf13/cobra"
)

var maxWordLenPerLevel = [...]int{
	5,
	7,
	10,
	16,
}

func init() {
	WordCmd.Flags().Uint16P("iterations", "n", 5, "Training iterations.")
	WordCmd.Flags().Float64P("speed", "s", 1, "Speed ratio to train with.")
	WordCmd.Flags().Uint16P("w-length", "m", 0, "Length of maximum word length for training.")

	WordCmd.Flags().Uint16P("level", "l", 0,
		"Level to have for training. Each level increases the length of the word available."+
			fmt.Sprintf(" Max: %v", len(maxWordLenPerLevel)),
	)

	WordCmd.Flags().String("words", "", "Custom word file to train on. You probably should start by using --level.")
	WordCmd.MarkFlagsOneRequired("level", "words")
	WordCmd.MarkFlagsMutuallyExclusive("w-length", "level")
}

var WordCmd = &cobra.Command{
	Use:     "word",
	Short:   "Train for decoding words.",
	Aliases: []string{"words"},
	RunE: func(cmd *cobra.Command, args []string) error {
		iterations, _ := cmd.Flags().GetUint16("iterations")
		if iterations == 0 {
			return fmt.Errorf("--iterations is set to zero.")
		}

		levelArg, _ := cmd.Flags().GetUint16("level")
		wordLength, _ := cmd.Flags().GetUint16("w-length")

		if wordLength == 0 {
			if levelArg == 0 {
				return fmt.Errorf("--level and --w-length are both set to zero.\n")
			}

			if int(levelArg) > len(maxWordLenPerLevel) {
				cmd.PrintErrf("Warning: Max level for decoding words is %v.\n", len(maxWordLenPerLevel))
				levelArg = uint16(len(maxWordLenPerLevel))
			}

			wordLength = uint16(maxWordLenPerLevel[levelArg-1])
		}

		wordFile, _ := cmd.Flags().GetString("words")
		fileReader := io.Reader(strings.NewReader(assets.Words))

		if len(wordFile) != 0 {
			file, err := os.Open(wordFile)
			if err != nil {
				return fmt.Errorf("Error opening %v: %v\n", wordFile, file)
			}

			defer file.Close()
			fileReader = io.Reader(file)
		} else {
			wordFile = "(the default word file)"
		}

		wordPool := []string(nil)
		scanner := bufio.NewScanner(fileReader)

		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			word := scanner.Text()
			word = strings.Map(func(r rune) rune {
				if r >= 'a' && r <= 'z' {
					return r
				}

				if r >= 'A' && r <= 'Z' {
					return r
				}

				if r == '-' {
					return r
				}

				return -1
			}, word)

			if len(word) <= int(wordLength) || int(wordLength) >= len(maxWordLenPerLevel) {
				wordPool = append(wordPool, word)
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("Error reading through %v: %v", wordFile, err)
		}

		words := []string(nil)
		for range min(len(wordPool), int(iterations)) {
			wordIdx := rand.Intn(len(wordPool))
			words = append(words, wordPool[wordIdx])

			wordPool[wordIdx] = wordPool[len(wordPool)-1]
			wordPool = wordPool[:len(wordPool)-1]
		}

		speed, _ := cmd.Flags().GetFloat64("speed")
		p := tea.NewProgram(newWordModel(words, wordLength, speed))

		if _, err := p.Run(); err != nil {
			return fmt.Errorf("Error running the model: %v\n", err)
		}

		return nil
	},
	Long: `The 'decode words' command gives the user drills to decode morse code words.
The flags in this command should be self-explanatory.

# How it works

For each item, you will be given a sound clip to listen. Input the word corresponding
to the sound. Pressing space or hitting enter when empty will repeat the sound.
Enter to confirm the answer.

====================================================================
Decode training (5 letter limit) (1 of 5)
> egg

(escape/ctrl+c to go back, space to repeat sound, enter to confirm)
====================================================================

At the end of the training session, you will be presented with the correct words
together with your input. Use that as learning and/or feedback for the next training session.

================================================================
Decoding words training results (5 letter limit, 5 iterations):

 #    Word   Correct?  Input
 1    egg    yes       egg

 2    type   no        tove
                        ??
 3    smart  yes       smart

 4    loose  no        loost
                           ?
 5    part   yes       part

(2/5 mistakes) (escape / ctrl+c / enter to go back)
================================================================

# Extras

This is the default word length limit if you specify --level/-l:
    =============================
    | Level | Word Length Limit |
    | ----- | ----------------- |
    | 1     | 5                 |
    | 2     | 7                 |
    | 3     | 10                |
    | 4+    | 16                |
    =============================

NOTE:
- For the convenience and the challenge, --speed can be used to slow down or speed
up the sound being played.`,
}
