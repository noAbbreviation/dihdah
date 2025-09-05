package decode

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

const defaultWordFile = "./assets/words.txt"

var maxWordLenPerLevel = []int{
	5,
	7,
	10,
	16,
}

var WordCmd = &cobra.Command{
	Use:     "word",
	Short:   "Train for decoding words.",
	Aliases: []string{"words"},
	RunE: func(cmd *cobra.Command, args []string) error {
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
		if len(wordFile) == 0 {
			wordFile = defaultWordFile
		}

		file, err := os.Open(wordFile)
		if err != nil {
			return fmt.Errorf("Error opening %v: %v\n", wordFile, file)
		}
		defer file.Close()

		words := []string(nil)

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			word := strings.TrimSpace(scanner.Text())

			if len(word) <= int(wordLength) {
				words = append(words, word)
			}
		}

		if err = scanner.Err(); err != nil {
			return fmt.Errorf("Error reading through %v: %v", wordFile, err)
		}

		fmt.Printf("words length: %v\n", len(words))
		os.Exit(0)

		var model tea.Model

		p := tea.NewProgram(model)
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("Error running the model: %v\n", err)
		}

		return nil
	},
}

func init() {
	WordCmd.Flags().UintP("iterations", "n", 0, "Training iterations.")
	WordCmd.Flags().Uint16P("w-length", "m", 0, "Length of maximum word length for training.")

	WordCmd.Flags().Uint16P("level", "l", 0,
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			"Level to have for training. Each level increases the length of the word available,",
			" so make sure you had finished all levels of 'letters'.",
			fmt.Sprintf(" Max: %v", len(maxWordLenPerLevel)),
		),
	)

	WordCmd.Flags().String("words", "", "Custom word file to train on. You probably should start by using --level.")
	WordCmd.MarkFlagsOneRequired("level", "words")
	WordCmd.MarkFlagsMutuallyExclusive("w-length", "level")
}
