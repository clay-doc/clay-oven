package main

import "strings"

// StructureFile holds the lines that make up the generated structure.yaml.
type StructureFile struct {
	Lines []string
}

// GenerateStructureFile recursively walks the directory tree and meta tree,
// appending formatted lines to the structure file.
func GenerateStructureFile(structure *StructureFile, dirTree DirNode, metaTree *MetaNode, indentLevel int, sink OutputSink) {
	for _, entry := range dirTree.Contents {
		if entry.IsDir {
			curIconNode := findIconNode(metaTree.Children, entry.PathName)
			sink.Verbose("Processing directory: " + entry.PathName)

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
			GenerateStructureFile(structure, entry, node, indentLevel+1, sink)
			continue
		}

		sink.Verbose("Processing file: " + entry.PathName)
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
	indent := strings.Repeat("    ", indentLevel)
	return indent + "- \"" + path + "#" + name + "#" + icon + "\"" + end + "\n"
}
