package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "portico",
		Short: "Portico is a Kubernetes controller that implements Gateway API",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})))

			return nil
		},
	}

	cmd.AddCommand(newController())
	cmd.AddCommand(newServer())

	cmd.PersistentFlags().String("log-level", "info", "set the log level (debug, info, warn, error)")

	return cmd
}
