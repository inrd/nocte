package app

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

var (
	markdownImagePattern = regexp.MustCompile(`^!\[(.*?)\]\((.+)\)$`)
	lookPath             = exec.LookPath
	runCommandOutput     = func(name string, args ...string) ([]byte, error) {
		return exec.Command(name, args...).CombinedOutput()
	}
	previewImageCache = struct {
		sync.Mutex
		rendered map[string][]string
	}{
		rendered: make(map[string][]string),
	}
)

type markdownImage struct {
	alt    string
	target string
}

func isMarkdownImageLine(line string) bool {
	_, ok := parseMarkdownImageLine(line)
	return ok
}

func parseMarkdownImageLine(line string) (markdownImage, bool) {
	match := markdownImagePattern.FindStringSubmatch(line)
	if len(match) != 3 {
		return markdownImage{}, false
	}

	target := strings.TrimSpace(match[2])
	if target == "" {
		return markdownImage{}, false
	}

	return markdownImage{
		alt:    strings.TrimSpace(match[1]),
		target: trimMarkdownImageTitle(target),
	}, true
}

func trimMarkdownImageTitle(target string) string {
	target = strings.TrimSpace(target)
	if strings.HasPrefix(target, "<") && strings.HasSuffix(target, ">") {
		return strings.TrimSpace(target[1 : len(target)-1])
	}

	for _, quote := range []string{` "`, " '", ` "`} {
		index := strings.LastIndex(target, quote)
		if index <= 0 || !strings.HasSuffix(target, string(quote[len(quote)-1])) {
			continue
		}
		return strings.TrimSpace(target[:index])
	}

	return target
}

func renderMarkdownImagePreview(notePath string, line string, width int) []string {
	image, ok := parseMarkdownImageLine(line)
	if !ok {
		return appendWrapped(nil, lipgloss.NewStyle(), line, width)
	}

	resolvedPath, err := resolveMarkdownImagePath(notePath, image.target)
	if err != nil {
		return renderMarkdownImageFallback(image, width, err.Error())
	}

	rendered, err := renderImageWithChafa(resolvedPath, width)
	if err != nil {
		return renderMarkdownImageFallback(image, width, err.Error())
	}

	return rendered
}

func resolveMarkdownImagePath(notePath string, target string) (string, error) {
	if target == "" {
		return "", errors.New("image path is empty")
	}
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return "", errors.New("remote images are not supported")
	}

	expandedTarget, err := expandMarkdownImagePath(target)
	if err != nil {
		return "", err
	}

	if filepath.IsAbs(expandedTarget) {
		return filepath.Clean(expandedTarget), nil
	}
	if notePath == "" {
		return "", errors.New("could not resolve relative image path")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(notePath), expandedTarget)), nil
}

func expandMarkdownImagePath(target string) (string, error) {
	target = os.ExpandEnv(strings.TrimSpace(target))
	if target == "~" || strings.HasPrefix(target, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("locate user home dir: %w", err)
		}
		if target == "~" {
			return home, nil
		}
		return filepath.Join(home, strings.TrimPrefix(target, "~/")), nil
	}

	return target, nil
}

func renderImageWithChafa(path string, width int) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("%s:%d:%d:%d", path, width, info.Size(), info.ModTime().UnixNano())

	previewImageCache.Lock()
	cached, ok := previewImageCache.rendered[cacheKey]
	previewImageCache.Unlock()
	if ok {
		return cached, nil
	}

	bin, err := lookPath("chafa")
	if err != nil {
		return nil, errors.New("install chafa to preview images")
	}

	height := max(4, min(12, width/2))
	output, err := runCommandOutput(bin, "--format=symbols", fmt.Sprintf("--size=%dx%d", width, height), path)
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			return nil, err
		}
		return nil, errors.New(message)
	}

	rendered := splitRenderedImageLines(string(output))
	if len(rendered) == 0 {
		return nil, errors.New("chafa returned no preview output")
	}

	previewImageCache.Lock()
	previewImageCache.rendered[cacheKey] = rendered
	previewImageCache.Unlock()

	return rendered, nil
}

func splitRenderedImageLines(output string) []string {
	output = strings.TrimRight(output, "\n")
	if output == "" {
		return nil
	}
	return strings.Split(output, "\n")
}

func renderMarkdownImageFallback(image markdownImage, width int, reason string) []string {
	label := "Image"
	if image.alt != "" {
		label = fmt.Sprintf("Image: %s", image.alt)
	}

	lines := appendWrapped(nil, previewMutedStyle, label, width)
	if image.target != "" {
		lines = appendWrapped(lines, previewMutedStyle, image.target, width)
	}
	if reason != "" {
		lines = appendWrapped(lines, previewMutedStyle, reason, width)
	}
	return lines
}

func clearPreviewImageCache() {
	previewImageCache.Lock()
	defer previewImageCache.Unlock()
	previewImageCache.rendered = make(map[string][]string)
}
