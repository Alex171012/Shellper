package cmd

import (
	"shellper/internal/config"
	"shellper/internal/tui"
	"github.com/spf13/cobra"
)

func tuiCmd() *cobra.Command {
	var sessionName string

	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Launch the TUI (terminal user interface)",
		Long: `Launch the interactive TUI — a rich terminal interface for Shellper.

Features:
  - Chat-like conversation layout
  - Auto-showing script preview and output panels
  - Vim-like key bindings
  - Session save/load
  - Mode switching (ask, explain, qa)

Key bindings:
  Enter          Send message
  i / a          Focus input (insert mode)
  Esc            Normal mode (unfocus input)
  j / k          Scroll messages
  J / K          Scroll script panel
  Tab            Cycle panels (script → output → input)
  :              Command mode
  q              Quit
  Ctrl+C         Cancel / Quit`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			opts := tui.TUIOpts{
				Backend:      cfg.Backend,
				Model:        cfg.Model,
				OllamaURL:    cfg.OllamaURL,
				OpenAIBase:   cfg.OpenAIBase,
				OpenAIKey:    cfg.OpenAIKey,
				SafetyMode:   cfg.Safety,
				DefaultShell: cfg.DefaultShell,
				SessionName:  sessionName,
			}

			return tui.StartTUI(opts)
		},
	}

	cmd.Flags().StringVar(&sessionName, "session", "", "Load a saved session")
	return cmd
}
