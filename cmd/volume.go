package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/y-yagi/doss/volume"
)

var findCmd = &cobra.Command{
	Use:   "volume",
	Short: "Docker volume helper",
	RunE: func(cmd *cobra.Command, args []string) error {
		executer, err := volume.NewExecuter(cmd, args, os.Stdout)
		if err != nil {
			return err
		}
		return executer.Execute()
	},
}

func init() {
	rootCmd.AddCommand(findCmd)
	findCmd.Flags().StringP("find", "f", "", "Find files by a specified pattern")
	findCmd.Flags().Bool("list", false, "Show Docker volume list")
}
