package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"shellper/internal/llm"
	"github.com/spf13/cobra"
)

type sessionFile struct {
	Name      string        `json:"name"`
	CreatedAt time.Time     `json:"created_at"`
	Messages  []llm.Message `json:"messages"`
}

func sessionDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "shellper", "sessions")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func sessionCmd() *cobra.Command {
	var name string

	saveCmd := &cobra.Command{
		Use:   "save [name]",
		Short: "Save current REPL session",
		Long: `Save the current REPL conversation to a file.
If no name is given, uses a timestamp.

Sessions are stored in ~/.config/shellper/sessions/`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionName := name
			if sessionName == "" {
				if len(args) > 0 {
					sessionName = args[0]
				} else {
					sessionName = time.Now().Format("session-20060102-150405")
				}
			}

			dir, err := sessionDir()
			if err != nil {
				return fmt.Errorf("session dir: %w", err)
			}

			path := filepath.Join(dir, sessionName+".json")
			sess := sessionFile{
				Name:      sessionName,
				CreatedAt: time.Now(),
				Messages:  replHistory,
			}

			data, err := json.MarshalIndent(sess, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal session: %w", err)
			}

			if err := os.WriteFile(path, data, 0644); err != nil {
				return fmt.Errorf("write session: %w", err)
			}

			fmt.Printf("Session saved: %s (%d messages)\n", path, len(sess.Messages))
			return nil
		},
	}
	saveCmd.Flags().StringVar(&name, "name", "", "Session name (overrides positional arg)")

	loadCmd := &cobra.Command{
		Use:   "load [name]",
		Short: "Load a REPL session",
		Long: `Load a previously saved REPL conversation.

Examples:
  shellper session load mysession
  shellper session load    # lists available sessions`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := sessionDir()
			if err != nil {
				return fmt.Errorf("session dir: %w", err)
			}

			if len(args) == 0 {
				entries, err := os.ReadDir(dir)
				if err != nil {
					if os.IsNotExist(err) {
						fmt.Println("No saved sessions found.")
						return nil
					}
					return fmt.Errorf("read sessions: %w", err)
				}

				if len(entries) == 0 {
					fmt.Println("No saved sessions found.")
					return nil
				}

				fmt.Println("Saved sessions:")
				for _, e := range entries {
					name := e.Name()
					if len(name) > 5 && name[len(name)-5:] == ".json" {
						data, err := os.ReadFile(filepath.Join(dir, name))
						if err != nil {
							continue
						}
						var sess sessionFile
						if json.Unmarshal(data, &sess) == nil {
							fmt.Printf("  %-30s %d msgs, %s\n",
								name[:len(name)-5],
								len(sess.Messages),
								sess.CreatedAt.Format("2006-01-02 15:04"),
							)
						}
					}
				}
				return nil
			}

			sessionName := args[0]
			path := filepath.Join(dir, sessionName+".json")
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("session %q not found: %w", sessionName, err)
			}

			var sess sessionFile
			if err := json.Unmarshal(data, &sess); err != nil {
				return fmt.Errorf("parse session: %w", err)
			}

			replHistory = sess.Messages
			fmt.Printf("Loaded session %q with %d messages.\n", sessionName, len(sess.Messages))
			fmt.Println("Start the REPL with 'shellper' to continue the conversation.")
			return nil
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List saved sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := sessionDir()
			if err != nil {
				return fmt.Errorf("session dir: %w", err)
			}

			entries, err := os.ReadDir(dir)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("No saved sessions found.")
					return nil
				}
				return fmt.Errorf("read sessions: %w", err)
			}

			if len(entries) == 0 {
				fmt.Println("No saved sessions found.")
				return nil
			}

			fmt.Println("Saved sessions:")
			for _, e := range entries {
				name := e.Name()
				if len(name) > 5 && name[len(name)-5:] == ".json" {
					data, err := os.ReadFile(filepath.Join(dir, name))
					if err != nil {
						continue
					}
					var sess sessionFile
					if json.Unmarshal(data, &sess) == nil {
						fmt.Printf("  %-30s %d msgs, %s\n",
							name[:len(name)-5],
							len(sess.Messages),
							sess.CreatedAt.Format("2006-01-02 15:04"),
						)
					}
				}
			}
			return nil
		},
	}

	sc := &cobra.Command{
		Use:   "session",
		Short: "Manage REPL sessions",
		Long: `Save, load, and list Shellper REPL sessions.

Sessions store your conversation history. Save a session to continue later.

Examples:
  shellper session save mywork
  shellper session load mywork
  shellper session list`,
	}

	sc.AddCommand(saveCmd)
	sc.AddCommand(loadCmd)
	sc.AddCommand(listCmd)
	return sc
}
