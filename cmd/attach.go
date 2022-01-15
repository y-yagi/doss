package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/y-yagi/doss/attach"
)

var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach a container",
	RunE: func(cmd *cobra.Command, args []string) error {
		executer, err := attach.NewExecuter(cmd, args, os.Stdout)
		if err != nil {
			return err
		}
		return executer.Execute()
	},
}

func init() {
	rootCmd.AddCommand(attachCmd)
}
