package decode

import (
	"fmt"

	"github.com/noAbbreviation/dihdah/cmd"
	"github.com/spf13/cobra"
)

var decodeCmd = &cobra.Command{
	Use:   "decode",
	Short: "Drills for decoding morse code",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hello from sub-command")
	},
}

func init() {
	cmd.Cmd.AddCommand(decodeCmd)
	decodeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
