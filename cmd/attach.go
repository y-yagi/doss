package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/y-yagi/doss/attach"
)

var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach a container",
	Run: func(cmd *cobra.Command, args []string) {
		executer, err := attach.NewExecuter(cmd, args, os.Stdout)
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
	rootCmd.AddCommand(attachCmd)
}
