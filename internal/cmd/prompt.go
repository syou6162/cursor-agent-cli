package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

func resolvePrompt(flagValue string, stdin io.Reader, isTerminal func(io.Reader) bool) (string, error) {
	if strings.TrimSpace(flagValue) != "" {
		return flagValue, nil
	}
	check := isTerminal
	if check == nil {
		check = defaultIsTerminal
	}
	if stdin == nil || check(stdin) {
		return "", fmt.Errorf("--prompt is required")
	}
	data, err := io.ReadAll(stdin)
	if err != nil {
		return "", fmt.Errorf("read prompt from stdin: %w", err)
	}
	prompt := strings.TrimSpace(string(data))
	if prompt == "" {
		return "", fmt.Errorf("--prompt is required")
	}
	return prompt, nil
}

func defaultIsTerminal(reader io.Reader) bool {
	if reader == nil {
		return true
	}
	f, ok := reader.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}
