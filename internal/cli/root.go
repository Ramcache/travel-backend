package cli

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "travel-api",
		Short: "Travel Backend API",
	}

	rootCmd.AddCommand(NewServeCmd())
	rootCmd.AddCommand(NewMigrateCmd())

	return rootCmd
}
