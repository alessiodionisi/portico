package cmd

import "github.com/spf13/cobra"

func newServer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Run the Portico server",
	}

	return cmd
}
