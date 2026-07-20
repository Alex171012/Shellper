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

func explainCmd(app *appContext) *cobra.Command {
	var files []string

	c := &cobra.Command{
		Use:   "explain [query]",
		Short: "Generate a script with explanations and plan",
		Long: `Ask Shellper to generate a script that teaches you.
Each command comes with a comment explaining what it does,
making it perfect for learning shell scripting.

Always shows a plan before the script. Use this to learn what each command does.

Examples:
  shellper explain "how to find and delete old logs"
  shellper explain "backup my home directory to /backup"
  shellper explain --file docker-compose.yml "explain this compose file"`,
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

			if err := handleScriptExecution(ctx, app, query, "explain", true); err != nil {
				return fmt.Errorf("explain: %w", err)
			}
			return nil
		},
	}

	c.Flags().StringSliceVar(&files, "file", nil, "Include file contents as context (can repeat)")
	return c
}
