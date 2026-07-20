package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func readFileForContext(path string) (string, error) {
	path = filepath.Clean(path)
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("cannot access %q: %w", path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%q is a directory, not a file", path)
	}
	if info.Size() > 100*1024 {
		return "", fmt.Errorf("%q is too large (%d bytes, max 100KB)", path, info.Size())
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %q: %w", path, err)
	}

	return fmt.Sprintf("File: %s\n```\n%s\n```", path, string(data)), nil
}

func extractFileRefs(text string) []string {
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

func injectFileContext(sysCtx string, filePaths []string) string {
	if len(filePaths) == 0 {
		return sysCtx
	}

	var b strings.Builder
	b.WriteString(sysCtx)
	b.WriteString("\n\nIncluded files:\n")

	for _, fp := range filePaths {
		content, err := readFileForContext(fp)
		if err != nil {
			b.WriteString(fmt.Sprintf("\n[Error reading %s: %v]", fp, err))
		} else {
			b.WriteString("\n" + content)
		}
	}

	return b.String()
}

func handleReadTool(script string) (string, bool) {
	refs := extractFileRefs(script)
	if len(refs) == 0 {
		return "", false
	}

	var context strings.Builder
	for _, ref := range refs {
		content, err := readFileForContext(ref)
		if err != nil {
			context.WriteString(fmt.Sprintf("\n[Error: %v]", err))
		} else {
			context.WriteString("\n" + content)
		}
	}

	return strings.TrimSpace(context.String()), true
}
