package app

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func openPath(path string) error {
	command, err := openCommand(runtime.GOOS, path)
	if err != nil {
		return err
	}

	if err := command.Start(); err != nil {
		return err
	}

	return command.Process.Release()
}

func openURL(rawURL string) error {
	command, err := openURLCommand(runtime.GOOS, rawURL)
	if err != nil {
		return err
	}

	if err := command.Start(); err != nil {
		return err
	}

	return command.Process.Release()
}

func copyToClipboard(content string) error {
	command, err := clipboardCommand(runtime.GOOS)
	if err != nil {
		return err
	}

	command.Stdin = strings.NewReader(content)
	if err := command.Run(); err != nil {
		return fmt.Errorf("could not copy to clipboard: %w", err)
	}

	return nil
}

func openCommand(goos string, path string) (*exec.Cmd, error) {
	cleanPath := filepath.Clean(path)

	switch goos {
	case "darwin":
		return exec.Command("open", cleanPath), nil
	case "linux":
		return exec.Command("xdg-open", cleanPath), nil
	case "windows":
		return exec.Command("explorer", cleanPath), nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", goos)
	}
}

func openURLCommand(goos string, rawURL string) (*exec.Cmd, error) {
	switch goos {
	case "darwin":
		return exec.Command("open", rawURL), nil
	case "linux":
		return exec.Command("xdg-open", rawURL), nil
	case "windows":
		return exec.Command("explorer", rawURL), nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", goos)
	}
}

func clipboardCommand(goos string) (*exec.Cmd, error) {
	switch goos {
	case "darwin":
		return exec.Command("pbcopy"), nil
	case "linux":
		for _, candidate := range [][]string{
			{"wl-copy"},
			{"xclip", "-selection", "clipboard"},
			{"xsel", "--clipboard", "--input"},
		} {
			if _, err := lookPath(candidate[0]); err == nil {
				return exec.Command(candidate[0], candidate[1:]...), nil
			}
		}
		return nil, fmt.Errorf("clipboard support requires wl-copy, xclip, or xsel")
	case "windows":
		return exec.Command("cmd", "/c", "clip"), nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", goos)
	}
}
