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
	if want := filepath.Join(home, ".config", "distrovroom", "setupconfig.yml"); path != want {
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
	if len(selected) != 9 {
		t.Fatalf("got %d selected items, want 9", len(selected))
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
	if err != nil || !strings.Contains(out, "'EMAIL_ADDRESS'") {
		t.Fatalf("missing email placeholder: %v: %s", err, out)
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

func TestInitEmptyHasOnlyBasicItems(t *testing.T) {
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	if _, err := runCLI(t, "config", "init-empty"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("setupconfig.empty.yml")
	if err != nil {
		t.Fatal(err)
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
	if len(selected) != 3 {
		t.Fatalf("basic config has %d items, want 3", len(selected))
	}
	for _, name := range []string{"ssh-keys", "github-cli", "codex"} {
		if !strings.Contains(string(raw), name) {
			t.Fatalf("basic config missing %s", name)
		}
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
