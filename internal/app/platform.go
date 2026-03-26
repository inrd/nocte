package app

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
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
