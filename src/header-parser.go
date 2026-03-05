package main

import "strings"

// HeaderInfo holds front-matter metadata extracted from a markdown file.
type HeaderInfo struct {
	Title string
	Icon  string
}

// ParseHeader extracts YAML front matter (between --- delimiters) from
// markdown content and returns the parsed header info.
func ParseHeader(content string) HeaderInfo {
	// Front matter format:
	// ---
	// title: Document Title
	// icon: fa-icon
	// ---

	var headerLines []string
	lines := strings.Split(content, "\n")
	inHeader := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "---" {
			if !inHeader {
				inHeader = true
			} else {
				break
			}
		} else if inHeader {
			headerLines = append(headerLines, line)
		}
	}

	headerInfo := HeaderInfo{}
	for _, line := range headerLines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "title":
			headerInfo.Title = value
		case "icon":
			headerInfo.Icon = value
		}
	}

	return headerInfo
}
