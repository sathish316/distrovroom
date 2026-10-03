package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed setupconfig.sample.yml
var sampleConfig []byte

const defaultConfigPath = "~/.config/distrovroom/setupconfig.yml"

type setupConfig struct {
	Environment string           `yaml:"environment"`
	Categories  []configCategory `yaml:"categories"`
}

type configCategory struct {
	Name  string   `yaml:"name"`
	Items []string `yaml:"items"`
}

type selectedItem struct {
	category catalogCategory
	item     catalogItem
}

func activeConfigPath() (string, error) {
	path := configPath
	if path == "" {
		path = defaultConfigPath
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("find home directory: %w", err)
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

func decodeConfig(raw []byte) (setupConfig, error) {
	var config setupConfig
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	if err := decoder.Decode(&config); err != nil {
		// Older setup configs used a category-to-items mapping. Read that form so
		// existing apply configs remain usable after the CLI gains config editing.
		var legacy struct {
			Environment string              `yaml:"environment"`
			Categories  map[string][]string `yaml:"categories"`
		}
		legacyDecoder := yaml.NewDecoder(bytes.NewReader(raw))
		legacyDecoder.KnownFields(true)
		if legacyErr := legacyDecoder.Decode(&legacy); legacyErr != nil {
			return setupConfig{}, fmt.Errorf("parse setup config: %w", err)
		}
		config.Environment = legacy.Environment
		categories := make([]string, 0, len(legacy.Categories))
		for name := range legacy.Categories {
			categories = append(categories, name)
		}
		sort.Strings(categories)
		for _, name := range categories {
			config.Categories = append(config.Categories, configCategory{Name: name, Items: legacy.Categories[name]})
		}
	}
	config.Environment = strings.ToLower(strings.TrimSpace(config.Environment))
	return config, nil
}

func loadConfig() (setupConfig, error) {
	path, err := activeConfigPath()
	if err != nil {
		return setupConfig{}, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return setupConfig{}, fmt.Errorf("read setup config %q: %w (run `distro-vroom config init-from-sample` first)", path, err)
	}
	return decodeConfig(raw)
}

// resolveConfig validates references and returns items in config order.
func resolveConfig(config setupConfig, catalog catalogFile) ([]selectedItem, error) {
	selected := make([]selectedItem, 0)
	seenCategories := make(map[string]bool)
	for _, configuredCategory := range config.Categories {
		category, ok := findCategory(catalog, configuredCategory.Name)
		if !ok {
			return nil, fmt.Errorf("config category %q is not in the catalog", configuredCategory.Name)
		}
		categoryKey := canonicalizeName(category.Name)
		if seenCategories[categoryKey] {
			return nil, fmt.Errorf("config category %q is duplicated", category.Name)
		}
		seenCategories[categoryKey] = true
		seenItems := make(map[string]bool)
		for _, itemName := range configuredCategory.Items {
			item, ok := findItem(category, itemName)
			if !ok {
				return nil, fmt.Errorf("config item %q is not in catalog category %q", itemName, category.Name)
			}
			itemKey := canonicalizeName(item.Name)
			if seenItems[itemKey] {
				return nil, fmt.Errorf("config item %q is duplicated in category %q", item.Name, category.Name)
			}
			seenItems[itemKey] = true
			selected = append(selected, selectedItem{category: category, item: item})
		}
	}
	return selected, nil
}

func writeNewFile(path string, raw []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("create %q: %w", path, err)
	}
	defer file.Close()
	if _, err := file.Write(raw); err != nil {
		return fmt.Errorf("write %q: %w", path, err)
	}
	return nil
}

func saveConfig(config setupConfig) error {
	path, err := activeConfigPath()
	if err != nil {
		return err
	}
	raw, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("encode setup config: %w", err)
	}
	// Write beside the destination so rename replaces it atomically.
	file, err := os.CreateTemp(filepath.Dir(path), ".setupconfig-*")
	if err != nil {
		return fmt.Errorf("create temporary config: %w", err)
	}
	defer os.Remove(file.Name())
	if err := file.Chmod(0600); err != nil {
		file.Close()
		return err
	}
	if _, err := file.Write(raw); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(file.Name(), path); err != nil {
		return fmt.Errorf("save config %q: %w", path, err)
	}
	return nil
}
