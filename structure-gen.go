package main

import "fmt"

func GenerateStructureFile(structure *StructureFile, dirTree DirNode, metaTree *MetaNode, indentLevel int) {
	for _, entry := range dirTree.Contents {
		if entry.IsDir {
			curIconNode := findIconNode(metaTree.Children, entry.PathName)
			fmt.Println("Processing directory:", entry.PathName)

			node := metaTree
			var name string
			var icon string

			if curIconNode != nil {
				node = curIconNode
				name = node.Name
				icon = node.Icon
			} else {
				name = entry.PathName
				icon = "fa-solid fa-folder"
			}

			dirLine := generateLine(entry.PathName, name, icon, indentLevel, ":")
			structure.Lines = append(structure.Lines, dirLine)
			GenerateStructureFile(structure, entry, node, indentLevel+1)
			continue
		}

		fmt.Println("Processing file:", entry.PathName)
		var name string
		var icon string

		if entry.FrontMatter.Title != "" {
			name = entry.FrontMatter.Title
		} else {
			name = entry.PathName
		}

		if entry.FrontMatter.Icon != "" {
			icon = entry.FrontMatter.Icon
		} else {
			icon = "fa-solid fa-file"
		}

		line := generateLine(entry.PathName, name, icon, indentLevel, "")
		structure.Lines = append(structure.Lines, line)
	}
}

func findIconNode(iconTree []*MetaNode, name string) *MetaNode {
	for _, node := range iconTree {
		if node.Path == name {
			return node
		}
	}

	return nil
}

func generateLine(path string, name string, icon string, indentLevel int, end string) string {
	indent := ""

	for i := 0; i < indentLevel; i++ {
		indent += "    "
	}

	return indent + "- \"" + path + "#" + name + "#" + icon + "\"" + end + "\n"
}

type StructureFile struct {
	Lines []string
}
