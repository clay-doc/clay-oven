package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// DirNode represents a file or directory in the documentation tree.
type DirNode struct {
	PathName    string
	IsDir       bool
	FrontMatter HeaderInfo
	Contents    []DirNode
}

// LoadDirectoryTree recursively loads a directory tree starting at path,
// populating the given parent node with its contents.
func LoadDirectoryTree(parent DirNode, path string) (DirNode, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return parent, fmt.Errorf("reading directory %q: %w", path, err)
	}

	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())

		if entry.IsDir() {
			child := DirNode{PathName: entry.Name(), IsDir: true}
			child, err = LoadDirectoryTree(child, fullPath)
			if err != nil {
				return parent, err
			}
			parent.Contents = append(parent.Contents, child)
		} else {
			content, err := os.ReadFile(fullPath)
			if err != nil {
				return parent, fmt.Errorf("reading file %q: %w", fullPath, err)
			}

			frontMatter := ParseHeader(string(content))
			parent.Contents = append(parent.Contents, DirNode{
				PathName:    entry.Name(),
				FrontMatter: frontMatter,
			})
		}
	}

	return parent, nil
}
