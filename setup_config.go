package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type setupConfig struct {
	Environment string              `yaml:"environment"`
	Commands    []configuredCommand `yaml:"commands"`
}

type configuredCommand struct {
	Category string `yaml:"category"`
	Item     string `yaml:"item"`
	Action   string `yaml:"action"`
}

func loadSetupConfig(requestedPath string) (setupConfig, error) {
	if requestedPath == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return setupConfig{}, fmt.Errorf("find config directory: %w", err)
		}
		requestedPath = filepath.Join(configDir, "distrovroom", "mysetupconfig.yml")
	}

	contents, err := os.ReadFile(requestedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return setupConfig{}, fmt.Errorf("setup config %q was not found; copy setupconfig.sample.yml to that path and edit it", requestedPath)
		}
		return setupConfig{}, fmt.Errorf("read setup config %q: %w", requestedPath, err)
	}

	var config setupConfig
	decoder := yaml.NewDecoder(bytes.NewReader(contents))
	decoder.KnownFields(true)
	if err := decoder.Decode(&config); err != nil {
		return setupConfig{}, fmt.Errorf("parse setup config %q: %w", requestedPath, err)
	}
	config.Environment = strings.ToLower(strings.TrimSpace(config.Environment))
	if config.Environment == "" {
		return setupConfig{}, fmt.Errorf("setup config %q must set an environment", requestedPath)
	}
	if len(config.Commands) == 0 {
		return setupConfig{}, fmt.Errorf("setup config %q must select at least one catalog command", requestedPath)
	}

	for index := range config.Commands {
		selection := &config.Commands[index]
		selection.Category = strings.TrimSpace(selection.Category)
		selection.Item = strings.TrimSpace(selection.Item)
		selection.Action = strings.ToLower(strings.TrimSpace(selection.Action))
		if selection.Category == "" || selection.Item == "" {
			return setupConfig{}, fmt.Errorf("setup config command %d must set both category and item", index+1)
		}
		switch selection.Action {
		case "install", "upgrade", "config", "test":
		default:
			return setupConfig{}, fmt.Errorf("setup config command %d has unsupported action %q", index+1, selection.Action)
		}
	}
	return config, nil
}
