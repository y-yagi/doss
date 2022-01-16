package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/y-yagi/doss/volume"
)

var findCmd = &cobra.Command{
	Use:   "volume",
	Short: "Docker volume helper",
	Run: func(cmd *cobra.Command, args []string) {
		executer, err := volume.NewExecuter(cmd, args, os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}

		if err := executer.Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(findCmd)
	findCmd.Flags().StringP("find", "f", "", "Find files by a specified pattern")
	findCmd.Flags().BoolP("remove", "r", false, "Remove a Docker volume")
	findCmd.Flags().BoolP("list", "l", false, "Show Docker volume list")
}
