package decode

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "decode",
	Short: "Drills for decoding morse code",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hello from sub-command")
	},
}

func init() {
	Cmd.AddCommand(WordCmd)
}
