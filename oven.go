package main

import "fmt"

func RunOven(args map[string]string) {
	var config, docsDir, output, folderMeta string

	noConfirm := args["-nc"] == "true"

	if val, ok := args["-c"]; ok {
		config = val
	} else {
		config = "clay.yaml"
	}

	if val, ok := args["-d"]; ok {
		docsDir = val
	} else {
		docsDir = "./docs"
	}

	if val, ok := args["-o"]; ok {
		output = val
	} else {
		output = "./output"
	}

	if val, ok := args["-fm"]; ok {
		folderMeta = val
	} else {
		folderMeta = "dir-meta.yaml"
	}

	fmt.Println("Oven is running with the following parameters:")
	fmt.Printf("\nConfig file:        %s\n", config)
	fmt.Printf("Doc directory:      %s\n", docsDir)
	fmt.Printf("Output directory:   %s\n", output)
	fmt.Printf("Folder meta file:  %s\n", folderMeta)
	fmt.Printf("No confirm:         %v\n", noConfirm)

	rootNode := DirNode{PathName: "docs"}
	var metaTree MetaNode
	var directoryTree DirNode

	_ = LoadConfigYaml(config)

	directoryTree = LoadDirectoryTree(rootNode, docsDir)
	metaTree = LoadMetaTree(folderMeta)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("\nCould not load: ", r)
			metaTree = MetaNode{Name: ""}
			directoryTree = DirNode{PathName: "docs"}
		}

		genStructureFile(directoryTree, metaTree)
	}()
}

func genStructureFile(dirTree DirNode, metaTree MetaNode) {
	fmt.Printf("\nGenerating structure file...\n\n")

	structure := StructureFile{Lines: []string{"- docs:\n"}}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("\nCould not gen structure file: ", r)
		}

		GenerateStructureFile(&structure, dirTree, &metaTree, 1)

		fmt.Printf("\nGenerated structure file content:\n\n")

		for _, line := range structure.Lines {
			fmt.Print(line)
		}
	}()
}
