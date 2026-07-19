package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

func askCmd(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "ask [query]",
		Short: "Generate and execute a shell script",
		Long: `Ask Shellper to perform a task. It generates a shell script,
checks it for safety, and executes it after confirmation if needed.

By default, ask mode is fast — no planning step. Use --think to enable it.

Examples:
  shellper ask "create a file called hello.txt with content 'Hello World'"
  shellper ask "find all files larger than 100MB in /tmp"
  shellper ask --think "backup my home directory"`,
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

			think := app.cfg.thinkEnabled
			if err := handleScriptExecution(ctx, app, query, "ask", think); err != nil {
				return fmt.Errorf("ask: %w", err)
			}
			return nil
		},
	}
}
