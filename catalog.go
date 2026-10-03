package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed catalog.yml
var embeddedCatalog []byte

type catalogFile struct {
	Categories []catalogCategory `yaml:"categories"`
}

type catalogCategory struct {
	Name  string        `yaml:"name"`
	Items []catalogItem `yaml:"items"`
}

type catalogItem struct {
	Name         string                         `yaml:"name"`
	Aliases      []string                       `yaml:"aliases"`
	Description  string                         `yaml:"description"`
	Check        itemCheck                      `yaml:"check"`
	Environments map[string]environmentCommands `yaml:"environments"`
}

type itemCheck struct {
	Binary string `yaml:"binary"`
	SSHKey bool   `yaml:"ssh-key"`
}

type environmentCommands struct {
	Install []string `yaml:"install"`
	Upgrade []string `yaml:"upgrade"`
	Config  []string `yaml:"config"`
	Test    []string `yaml:"test"`
}

func loadCatalog() (catalogFile, error) {
	raw := embeddedCatalog
	if catalogPath != "" {
		contents, err := os.ReadFile(catalogPath)
		if err != nil {
			return catalogFile{}, fmt.Errorf("read catalog %q: %w", catalogPath, err)
		}
		raw = contents
	}

	var catalog catalogFile
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	if err := decoder.Decode(&catalog); err != nil {
		return catalogFile{}, fmt.Errorf("parse catalog: %w", err)
	}
	for _, category := range catalog.Categories {
		if strings.TrimSpace(category.Name) == "" {
			return catalogFile{}, fmt.Errorf("catalog contains a category with no name")
		}
		for _, item := range category.Items {
			if strings.TrimSpace(item.Name) == "" {
				return catalogFile{}, fmt.Errorf("category %q contains an item with no name", category.Name)
			}
		}
	}
	return catalog, nil
}

func findCategory(catalog catalogFile, name string) (catalogCategory, bool) {
	for _, category := range catalog.Categories {
		if canonicalizeName(category.Name) == canonicalizeName(name) {
			return category, true
		}
	}
	return catalogCategory{}, false
}

func findItem(category catalogCategory, name string) (catalogItem, bool) {
	query := canonicalizeName(name)
	for _, item := range category.Items {
		if canonicalizeName(item.Name) == query {
			return item, true
		}
	}
	for _, item := range category.Items {
		for _, alias := range item.Aliases {
			if canonicalizeName(alias) == query {
				return item, true
			}
		}
	}
	return catalogItem{}, false
}

// canonicalizeName lowercases names and turns whitespace runs into CLI-friendly hyphens.
func canonicalizeName(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(value))), "-")
}
