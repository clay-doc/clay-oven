package main

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

// Config holds the top-level clay.yaml configuration.
type Config struct {
	Title          string       `yaml:"title"`
	Favicon        string       `yaml:"favicon"`
	BaseURL        string       `yaml:"baseURL"`
	FontawesomeKit string       `yaml:"fontawesomeKit"`
	Navbar         NavbarConfig `yaml:"navbar"`
	Index          IndexConfig  `yaml:"index"`
	Langs          []string     `yaml:"langs"`
}

// NavbarConfig describes the navigation bar layout.
type NavbarConfig struct {
	Logo   string       `yaml:"logo"`
	Source LinkConfig   `yaml:"source"`
	Links  []LinkConfig `yaml:"links"`
}

// LinkConfig describes a single named link.
type LinkConfig struct {
	Name string `yaml:"name"`
	Icon string `yaml:"icon"`
	Link string `yaml:"link"`
}

// IndexConfig describes the landing page content.
type IndexConfig struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Icon        string `yaml:"icon"`
}

// MetaNode represents directory metadata (name, icon) loaded from dir-meta.yaml.
type MetaNode struct {
	Icon     string      `yaml:"icon"`
	Name     string      `yaml:"name"`
	Path     string      `yaml:"path"`
	Children []*MetaNode `yaml:"children"`
}

// LoadConfigYaml reads and parses the clay config file at the given path.
func LoadConfigYaml(path string) (Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("reading config %q: %w", path, err)
	}

	var config Config
	if err = yaml.Unmarshal(file, &config); err != nil {
		return Config{}, fmt.Errorf("parsing config %q: %w", path, err)
	}

	return config, nil
}

// LoadMetaTree reads and parses the directory metadata file at the given path.
func LoadMetaTree(path string) (MetaNode, error) {
	root := MetaNode{}

	file, err := os.ReadFile(path)
	if err != nil {
		return root, fmt.Errorf("reading meta file %q: %w", path, err)
	}

	var children []*MetaNode
	if err = yaml.Unmarshal(file, &children); err != nil {
		return root, fmt.Errorf("parsing meta file %q: %w", path, err)
	}

	root.Children = children
	return root, nil
}
