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

func TestStatusUsesReadOnlyChecksAndPlainOutput(t *testing.T) {
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
