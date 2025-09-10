package decode

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

const defaultQuotesFile = "./assets/quotes.txt"

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

		if len(quotesFile) == 0 {
			quotesFile = defaultQuotesFile
		}

		file, err := os.Open(quotesFile)
		if err != nil {
			return fmt.Errorf("Error reading %v: %v", quotesFile, err)
		}

		defer file.Close()

		quotes := []string(nil)
		scanner := bufio.NewScanner(file)

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

		p := tea.NewProgram(newQuoteModel(randomQuote, speed))
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("Error running the program: %v", err)
		}

		return nil
	},
}
