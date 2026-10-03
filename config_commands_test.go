package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	previousCatalogPath, previousConfigPath := catalogPath, configPath
	defer func() {
		catalogPath, configPath = previousCatalogPath, previousConfigPath
	}()
	var out bytes.Buffer
	cmd := newRootCommand()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestDefaultConfigFileFlagUsesHomeConfigPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := newRootCommand()
	if got := root.PersistentFlags().Lookup("config-file").DefValue; got != defaultConfigPath {
		t.Fatalf("config-file default = %q, want %q", got, defaultConfigPath)
	}
	path, err := activeConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, ".config", "distrovroom", "config.yml"); path != want {
		t.Fatalf("active config path = %q, want %q", path, want)
	}
}

func TestSampleConfigAndSelectedCatalogCommands(t *testing.T) {
	catalog, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	config, err := decodeConfig(sampleConfig)
	if err != nil {
		t.Fatal(err)
	}
	selected, err := resolveConfig(config, catalog)
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 10 {
		t.Fatalf("got %d selected items, want 10", len(selected))
	}
	if len(config.Categories) != len(catalog.Categories) {
		t.Fatalf("sample has %d categories, catalog has %d", len(config.Categories), len(catalog.Categories))
	}
	for categoryIndex, category := range catalog.Categories {
		configured := config.Categories[categoryIndex]
		if configured.Name != category.Name || len(configured.Items) != len(category.Items) {
			t.Fatalf("sample category %d does not match catalog category %q", categoryIndex, category.Name)
		}
		for itemIndex, item := range category.Items {
			if configured.Items[itemIndex] != item.Name {
				t.Fatalf("sample item %d in %s = %q, want %q", itemIndex, category.Name, configured.Items[itemIndex], item.Name)
			}
		}
	}
	path := filepath.Join(t.TempDir(), "setupconfig.yml")
	if out, err := runCLI(t, "--config-file", path, "config", "init-from-sample"); err != nil {
		t.Fatalf("init: %v: %s", err, out)
	}
	out, err := runCLI(t, "--config-file", path, "catalog", "commands", "agents", "pi")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "curl -fsSL https://pi.dev/install.sh | sh") || strings.Contains(out, "@openai/codex") {
		t.Fatalf("unexpected selected commands: %s", out)
	}
}

func TestConfigureInitFromEmptyUsesActivePathAndPreservesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.yml")
	if _, err := runCLI(t, "--config-file", path, "configure", "init-from-empty"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, emptyConfig) {
		t.Fatal("created config differs from bundled minimal template")
	}
	config, err := decodeConfig(raw)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	selected, err := resolveConfig(config, catalog)
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 3 || selected[0].item.Name != "github-ssh-keys" || selected[1].item.Name != "github-cli" || selected[2].item.Name != "codex" {
		t.Fatalf("unexpected minimal selections: %#v", selected)
	}
	if _, err := runCLI(t, "--config-file", path, "configure", "init-from-sample"); err == nil {
		t.Fatal("expected existing config to be preserved")
	}
	again, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(again, raw) {
		t.Fatalf("existing config changed: %v", err)
	}
}

func TestCatalogCommandEmailTemplate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "setupconfig.yml")
	if err := os.WriteFile(path, []byte("categories:\n  - name: setup\n    items: [ssh-keys]\n"), 0600); err != nil {
		t.Fatal(err)
	}
	out, err := runCLI(t, "--config-file", path, "catalog", "commands", "setup", "ssh-keys", "--email", "you@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "ssh-keygen -t ed25519 -C 'you@example.com'") || strings.Contains(out, "{{") {
		t.Fatalf("email was not rendered: %s", out)
	}
	out, err = runCLI(t, "--config-file", path, "catalog", "commands", "setup", "ssh-keys")
	if err == nil {
		t.Fatalf("missing email should be rejected: %v: %s", err, out)
	}
	if _, err := runCLI(t, "--config-file", path, "catalog", "commands", "setup", "ssh-keys", "--email", "invalid"); err == nil {
		t.Fatal("expected invalid email to be rejected")
	}
}

func TestConfigAddRemoveAndAliasDeduplication(t *testing.T) {
	path := filepath.Join(t.TempDir(), "setupconfig.yml")
	if err := os.WriteFile(path, []byte("categories: []\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := runCLI(t, "--config-file", path, "config", "add", "cli-programming", "gh"); err != nil {
		t.Fatal(err)
	}
	if _, err := runCLI(t, "--config-file", path, "config", "add", "cli-programming", "github-cli"); err == nil {
		t.Fatal("expected duplicate alias to be rejected")
	}
	out, err := runCLI(t, "--config-file", path, "config", "list")
	if err != nil || !strings.Contains(out, "github-cli") {
		t.Fatalf("list: %v: %s", err, out)
	}
	if _, err := runCLI(t, "--config-file", path, "config", "remove", "cli-programming", "gh"); err != nil {
		t.Fatal(err)
	}
	out, err = runCLI(t, "--config-file", path, "config", "list")
	if err != nil || strings.Contains(out, "github-cli") {
		t.Fatalf("remove: %v: %s", err, out)
	}
}

func TestStatusUsesCatalogTestsAndPlainOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "setupconfig.yml")
	if err := os.WriteFile(path, []byte("categories:\n  - name: agents\n    items: [codex, pi]\n"), 0600); err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "codex"), []byte("#!/bin/sh\nexit 0\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	out, err := runCLI(t, "--config-file", path, "status", "list", "agents", "--color", "never")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "agents (1/2 installed)") || !strings.Contains(out, "installed  codex") || !strings.Contains(out, "not-installed  pi") || strings.Contains(out, "\x1b[") {
		t.Fatalf("status: %s", out)
	}
}

func TestStatusUsesCatalogTestCommands(t *testing.T) {
	item := catalogItem{Environments: map[string]environmentCommands{
		"default": {Test: []string{"exit 0"}},
	}}
	if got := checkInstalled(item); got != installed {
		t.Fatalf("passing test = %s, want installed", got)
	}
	item.Environments["default"] = environmentCommands{Test: []string{"exit 1"}}
	if got := checkInstalled(item); got != notInstalled {
		t.Fatalf("failing test = %s, want not-installed", got)
	}
	item.Environments["default"] = environmentCommands{}
	if got := checkInstalled(item); got != unknown {
		t.Fatalf("no test = %s, want unknown", got)
	}
}

func TestApplyUsesSelectedItemsAndDefaultEnvironment(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "setupconfig.yml")
	catalogFile := filepath.Join(dir, "catalog.yml")
	outputFile := filepath.Join(dir, "installed.txt")
	if err := os.WriteFile(configFile, []byte("environment: arch\ncategories:\n  - name: agents\n    items: [demo]\n"), 0600); err != nil {
		t.Fatal(err)
	}
	catalog := "categories:\n  - name: agents\n    items:\n      - name: demo\n        environments:\n          default:\n            install:\n              - printf %s {{.email}} > " + shellQuote(outputFile) + "\n"
	if err := os.WriteFile(catalogFile, []byte(catalog), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := runCLI(t, "--catalog", catalogFile, "--config-file", configFile, "apply", "agents", "demo", "--email", "you@example.com"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "you@example.com" {
		t.Fatalf("applied output = %q", raw)
	}
}

func TestLegacyCategoryMapConfigStillLoads(t *testing.T) {
	previousCatalogPath := catalogPath
	catalogPath = ""
	defer func() { catalogPath = previousCatalogPath }()
	config, err := decodeConfig([]byte("environment: arch\ncategories:\n  setup: [github-ssh-keys]\n  cli-programming: [github-cli]\n"))
	if err != nil {
		t.Fatal(err)
	}
	if config.Environment != "arch" || len(config.Categories) != 2 {
		t.Fatalf("legacy config was not converted: %#v", config)
	}
	catalog, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	selected, err := resolveConfig(config, catalog)
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 2 {
		t.Fatalf("legacy config selected %d items, want 2", len(selected))
	}
}
