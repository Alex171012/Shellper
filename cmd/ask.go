package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"
	"shellper/internal/tools"
)

func askCmd(app *appContext) *cobra.Command {
	var files []string

	c := &cobra.Command{
		Use:   "ask [query]",
		Short: "Generate and execute a shell script",
		Long: `Ask Shellper to perform a task. It generates a shell script,
checks it for safety, and executes it after confirmation if needed.

By default, ask mode is fast — no planning step. Use --think to enable it.

Examples:
  shellper ask "create a file called hello.txt with content 'Hello World'"
  shellper ask "find all files larger than 100MB in /tmp"
  shellper ask --think "backup my home directory"
  shellper ask --file main.go "review this file for issues"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.Join(args, " ")

			if len(files) > 0 {
				var parts []string
				for _, f := range files {
					content, err := tools.ReadFileForContext(f)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
						continue
					}
					parts = append(parts, content)
				}
				if len(parts) > 0 {
					query = query + "\n\n" + strings.Join(parts, "\n")
				}
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

	c.Flags().StringSliceVar(&files, "file", nil, "Include file contents as context (can repeat)")
	return c
}
