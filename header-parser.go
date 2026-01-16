package main

import "strings"

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

type HeaderInfo struct {
	Title string
	Icon  string
}
