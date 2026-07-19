package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

func qaCmd(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "qa [question]",
		Short: "Ask a Linux/shell question (no execution)",
		Long: `Ask Shellper a question about Linux or shell scripting.
This mode only returns information — no commands are executed.

Examples:
  shellper qa "what is a symlink"
  shellper qa "how does piping work in bash"
  shellper qa "difference between hard link and soft link"`,
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
}
