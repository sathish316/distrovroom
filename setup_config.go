package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultSetupConfigRelativePath = ".config/distrovroom/setupconfig.yml"

type legacySetupConfig struct {
	Environment string              `yaml:"environment"`
	Categories  map[string][]string `yaml:"categories"`
}

func loadSetupConfig(requestedPath string) (legacySetupConfig, error) {
	if requestedPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return legacySetupConfig{}, fmt.Errorf("find home directory: %w", err)
		}
		requestedPath = filepath.Join(homeDir, defaultSetupConfigRelativePath)
	}

	contents, err := os.ReadFile(requestedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return legacySetupConfig{}, fmt.Errorf("setup config %q was not found; copy setupconfig.sample.yml to that path and edit it", requestedPath)
		}
		return legacySetupConfig{}, fmt.Errorf("read setup config %q: %w", requestedPath, err)
	}

	var config legacySetupConfig
	decoder := yaml.NewDecoder(bytes.NewReader(contents))
	decoder.KnownFields(true)
	if err := decoder.Decode(&config); err != nil {
		return legacySetupConfig{}, fmt.Errorf("parse setup config %q: %w", requestedPath, err)
	}
	config.Environment = strings.ToLower(strings.TrimSpace(config.Environment))
	if config.Environment == "" {
		return legacySetupConfig{}, fmt.Errorf("setup config %q must set an environment", requestedPath)
	}
	if len(config.Categories) == 0 {
		return legacySetupConfig{}, fmt.Errorf("setup config %q must select at least one category", requestedPath)
	}
	for category, items := range config.Categories {
		if strings.TrimSpace(category) == "" {
			return legacySetupConfig{}, fmt.Errorf("setup config %q has an empty category name", requestedPath)
		}
		if len(items) == 0 {
			return legacySetupConfig{}, fmt.Errorf("setup config category %q must select at least one item", category)
		}
		for index, item := range items {
			if strings.TrimSpace(item) == "" {
				return legacySetupConfig{}, fmt.Errorf("setup config category %q has an empty item at position %d", category, index+1)
			}
		}
	}
	return config, nil
}

func configuredItems(config legacySetupConfig, categoryName string) ([]string, bool) {
	for name, items := range config.Categories {
		if strings.EqualFold(strings.TrimSpace(name), strings.TrimSpace(categoryName)) {
			return items, true
		}
	}
	return nil, false
}
