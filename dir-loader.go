package main

import "os"

func LoadDirectoryTree(parent DirNode, path string) DirNode {
	entries, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			newNode := DirNode{PathName: entry.Name(), IsDir: true}
			subDir := LoadDirectoryTree(newNode, path+"/"+entry.Name())
			parent.Contents = append(parent.Contents, subDir)
		} else {
			content, err := os.ReadFile(path + "/" + entry.Name())
			if err != nil {
				panic(err)
			}

			frontMatter := ParseHeader(string(content))
			parent.Contents = append(parent.Contents, DirNode{PathName: entry.Name(), FrontMatter: frontMatter})
		}
	}

	return parent
}

type DirNode struct {
	PathName    string
	IsDir       bool
	FrontMatter HeaderInfo
	Contents    []DirNode
}
