package main

import (
	"os"

	"github.com/goccy/go-yaml"
)

func LoadConfigYaml(path string) Config {
	file, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		panic(err)
	}

	return config
}

func LoadMetaTree(path string) MetaNode {
	root := MetaNode{}

	file, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var icon []*MetaNode
	err = yaml.Unmarshal(file, &icon)
	if err != nil {
		panic(err)
	}

	root.Children = icon
	return root
}

type Config struct {
	Title          string
	Favicon        string
	BaseURL        string
	FontawesomeKit string
	Navbar         NavbarConfig
	Index          IndexConfig
	Langs          []string
}

type NavbarConfig struct {
	Logo   string
	Source LinkConfig
	Links  []LinkConfig
}

type LinkConfig struct {
	Name string
	Icon string
	Link string
}

type IndexConfig struct {
	Title       string
	Description string
	Icon        string
}

type MetaNode struct {
	Icon     string
	Name     string
	Path     string
	Children []*MetaNode
}
