package decode

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/noAbbreviation/dihdah/assets"
	"github.com/spf13/cobra"
)

func init() {
	QuoteCmd.Flags().Float64P("speed", "s", 1, "Speed to do the training with.")
	QuoteCmd.Flags().String("quotes", "", "Custom quote file to use for training.")
}

var QuoteCmd = &cobra.Command{
	Use:     "quote",
	Short:   "Train for decoding quotes.",
	Aliases: []string{"quotes"},
	RunE: func(cmd *cobra.Command, args []string) error {
		quotesFile, _ := cmd.Flags().GetString("quotes")
		fileReader := io.Reader(strings.NewReader(assets.Quotes))

		if len(quotesFile) != 0 {
			file, err := os.Open(quotesFile)
			if err != nil {
				return fmt.Errorf("Error reading %v: %v", quotesFile, err)
			}

			defer file.Close()
			fileReader = io.Reader(file)
		} else {
			quotesFile = "(the default quotes file)"
		}

		quotes := []string(nil)
		scanner := bufio.NewScanner(fileReader)

		for scanner.Scan() {
			quote := strings.TrimSpace(scanner.Text())
			quotes = append(quotes, quote)
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("Error scanning %v: %v", quotesFile, err)
		}

		randomQuote := quotes[rand.Intn(len(quotes))]

		speed, _ := cmd.Flags().GetFloat64("speed")
		if speed == 0 {
			return fmt.Errorf("Speed must not be zero.")
		}

		p := tea.NewProgram(NewQuoteModel(randomQuote, speed, nil))
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("Error running the program: %v", err)
		}

		return nil
	},
	Long: `The 'decode quotes' command gives the user drills to decode sentences.
The flags should be self-explanatory.

# How it works

You will be given a long sound clip, which is an encoded morse code sentence. Ctrl+l
will either stop or play the clip. Ctrl+c will either clear your input or go back.
Ctrl+s will confirm your input.

===========================================================================================
Decode quote training

┃  1 ?????
┃
┃
┃
┃
┃

(ctrl+l to stop/restart playing, ctrl+s to confirm answer, ctrl+c/esc to clear or go back)
===========================================================================================

At the end of the training session, you will be presented with the correct quote
played. Use that as learning and/or feedback for the next training session.

=========================================
Decode quote training results

   Do all things with love. - Og Mandino
      ?         ?
>> do nll things_with love    og mandino

(1/28 mistakes) (ctrl+c/esc to go back)
=========================================

NOTE:
- For the convenience and challenge, --speed can be used to slow down or speed
up the sound being played.`,
}
