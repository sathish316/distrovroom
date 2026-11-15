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

