//go:build !darwin
// +build !darwin

package cmd

import (
	"golang.design/x/clipboard"
)

// copyToClipboard uses the golang.design/x/clipboard library on non-macOS systems
func copyToClipboard(text string) error {
	return clipboard.Write(clipboard.FmtText, []byte(text))
}

// pasteFromClipboard uses the golang.design/x/clipboard library on non-macOS systems
func pasteFromClipboard() (string, error) {
	data := clipboard.Read(clipboard.FmtText)
	if data == nil {
		return "", nil
	}
	return string(data), nil
}