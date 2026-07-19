package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

func explainCmd(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "explain [query]",
		Short: "Generate a script with explanations and plan",
		Long: `Ask Shellper to generate a script that teaches you.
Each command comes with a comment explaining what it does,
making it perfect for learning shell scripting.

Always shows a plan before the script. Use this to learn what each command does.

Examples:
  shellper explain "how to find and delete old logs"
  shellper explain "backup my home directory to /backup"
  shellper explain "check disk usage by directory"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			for _, a := range args[1:] {
				query += " " + a
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt)
			go func() {
				<-sigCh
				cancel()
			}()

			if err := handleScriptExecution(ctx, app, query, "explain", true); err != nil {
				return fmt.Errorf("explain: %w", err)
			}
			return nil
		},
	}
}
