package app

import (
	"regexp"
	"strings"
)

var (
	markdownLinkPattern = regexp.MustCompile(`\[(.*?)\]\((https?://[^\s)]+)\)`)
	bareURLPattern      = regexp.MustCompile(`https?://[^\s<>()]+`)
)

func extractNoteLinks(content string) []noteLink {
	if strings.TrimSpace(content) == "" {
		return nil
	}

	links := make([]noteLink, 0)
	occupied := make([][2]int, 0)
	seen := make(map[string]struct{})
	seenURLs := make(map[string]struct{})

	for _, match := range markdownLinkPattern.FindAllStringSubmatchIndex(content, -1) {
		if len(match) < 6 {
			continue
		}

		label := strings.TrimSpace(content[match[2]:match[3]])
		url := strings.TrimSpace(content[match[4]:match[5]])
		if url == "" {
			continue
		}

		key := label + "\x00" + url
		if _, ok := seen[key]; ok {
			continue
		}

		links = append(links, noteLink{label: label, url: url})
		occupied = append(occupied, [2]int{match[0], match[1]})
		seen[key] = struct{}{}
		seenURLs[url] = struct{}{}
	}

	for _, match := range bareURLPattern.FindAllStringIndex(content, -1) {
		if overlapsMarkdownLink(match, occupied) {
			continue
		}

		url := strings.TrimRight(content[match[0]:match[1]], ".,;:!?")
		if url == "" {
			continue
		}
		if _, ok := seenURLs[url]; ok {
			continue
		}
		if _, ok := seen["\x00"+url]; ok {
			continue
		}

		links = append(links, noteLink{url: url})
		seen["\x00"+url] = struct{}{}
		seenURLs[url] = struct{}{}
	}

	return links
}

func overlapsMarkdownLink(match []int, occupied [][2]int) bool {
	for _, span := range occupied {
		if match[0] < span[1] && span[0] < match[1] {
			return true
		}
	}

	return false
}

func (m *Model) openLinksDialog() {
	m.activeDialog = "links"
	m.dialogLinks = extractNoteLinks(m.editor.Value())
	m.dialogIndex = -1
	m.dialogOffset = 0
	if len(m.dialogLinks) > 0 {
		m.dialogIndex = 0
	}
	m.status = ""
	m.isError = false
	m.editor.Focus()
}

func (m *Model) openSelectedDialogLink() error {
	if len(m.dialogLinks) == 0 {
		m.closeDialog()
		return nil
	}

	if m.dialogIndex < 0 || m.dialogIndex >= len(m.dialogLinks) {
		m.dialogIndex = 0
	}

	link := m.dialogLinks[m.dialogIndex]
	if err := openURLWithSystemApp(link.url); err != nil {
		m.status = "Could not open link: " + err.Error()
		m.isError = true
		return err
	}

	m.closeDialog()
	m.status = "Opened link in browser"
	m.isError = false
	return nil
}
