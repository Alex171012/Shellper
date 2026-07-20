package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ReadFileForContext(path string) (string, error) {
	path = filepath.Clean(path)
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("cannot access %q: %w", path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%q is a directory", path)
	}
	if info.Size() > 100*1024 {
		return "", fmt.Errorf("%q is too large (%d bytes)", path, info.Size())
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %q: %w", path, err)
	}

	return fmt.Sprintf("=== %s ===\n%s\n=== end ===", path, string(data)), nil
}

func ExtractFileRefs(text string) []string {
	var refs []string
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "@read ") {
			path := strings.TrimSpace(trimmed[6:])
			if path != "" {
				refs = append(refs, path)
			}
		}
	}
	return refs
}

func HandleReadTool(script string) (string, bool) {
	refs := ExtractFileRefs(script)
	if len(refs) == 0 {
		return "", false
	}

	var b strings.Builder
	for _, ref := range refs {
		content, err := ReadFileForContext(ref)
		if err != nil {
			b.WriteString(fmt.Sprintf("\n[Error: %v]", err))
		} else {
			b.WriteString("\n" + content)
		}
	}

	return strings.TrimSpace(b.String()), true
}
