//go:build darwin
// +build darwin

package cmd

import (
	"bytes"
	"os/exec"
)

// copyToClipboard uses pbcopy on macOS
func copyToClipboard(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = bytes.NewReader([]byte(text))
	return cmd.Run()
}

// pasteFromClipboard uses pbpaste on macOS
func pasteFromClipboard() (string, error) {
	cmd := exec.Command("pbpaste")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}