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

func qaCmd(app *appContext) *cobra.Command {
	var files []string

	c := &cobra.Command{
		Use:   "qa [question]",
		Short: "Ask a Linux/shell question (no execution)",
		Long: `Ask Shellper a question about Linux or shell scripting.
This mode only returns information — no commands are executed.

Examples:
  shellper qa "what is a symlink"
  shellper qa "how does piping work in bash"
  shellper qa --file /var/log/syslog "analyze this log file"`,
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

			messages := buildQAPrompt(query, buildSystemContext())
			response, err := getLLMResponse(ctx, app, messages)
			if err != nil {
				return fmt.Errorf("qa: %w", err)
			}

			if app.cfg.renderMD {
				fmt.Println(renderMarkdown(response))
			} else {
				fmt.Println(response)
			}
			return nil
		},
	}

	c.Flags().StringSliceVar(&files, "file", nil, "Include file contents as context (can repeat)")
	return c
}
