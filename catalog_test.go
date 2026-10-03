package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindItemPrefersCanonicalNameOverEarlierAlias(t *testing.T) {
	category := catalogCategory{
		Name: "cli-programming",
		Items: []catalogItem{
			{Name: "alpha", Aliases: []string{"beta"}},
			{Name: "beta"},
		},
	}

	item, ok := findItem(category, "BETA")
	if !ok {
		t.Fatal("findItem did not find the canonical item")
	}
	if item.Name != "beta" {
		t.Fatalf("findItem returned %q, want canonical item %q", item.Name, "beta")
	}
}

func TestCanonicalizeNameMatchesHyphenatedCLIName(t *testing.T) {
	if got, want := canonicalizeName("  GitHub   CLI "), "github-cli"; got != want {
		t.Fatalf("canonicalizeName() = %q, want %q", got, want)
	}
}

func TestJavaScriptShellSetupIsIdempotentAndStatusUsesNvm(t *testing.T) {
	catalog, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	category, ok := findCategory(catalog, "programming-env")
	if !ok {
		t.Fatal("programming-env category missing")
	}
	item, ok := findItem(category, "javascript")
	if !ok {
		t.Fatal("javascript item missing")
	}
	commands := item.Environments["default"]
	if len(commands.Install) != 3 || len(commands.Test) != 1 {
		t.Fatalf("unexpected JavaScript commands: %#v", commands)
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	for range 2 {
		output, err := exec.Command("sh", "-c", commands.Install[1]).CombinedOutput()
		if err != nil {
			t.Fatalf("configure bashrc: %v: %s", err, output)
		}
	}
	raw, err := os.ReadFile(filepath.Join(home, ".bashrc"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(raw), "export NVM_DIR=") != 1 || strings.Count(string(raw), "$NVM_DIR/nvm.sh") != 2 {
		t.Fatalf("nvm shell setup was duplicated or incomplete: %s", raw)
	}

	nvmDir := filepath.Join(home, ".nvm")
	if err := os.MkdirAll(nvmDir, 0700); err != nil {
		t.Fatal(err)
	}
	nvmScript := `nvm() { [ "$NVM_DIR" = "$HOME/.nvm" ] && [ "$1" = use ] && [ "$2" = 24 ]; }`
	if err := os.WriteFile(filepath.Join(nvmDir, "nvm.sh"), []byte(nvmScript), 0600); err != nil {
		t.Fatal(err)
	}
	binDir := filepath.Join(home, "bin")
	if err := os.MkdirAll(binDir, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"node", "npm"} {
		if err := os.WriteFile(filepath.Join(binDir, name), []byte("#!/bin/sh\nexit 0\n"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))
	if got := checkInstalled(item); got != installed {
		t.Fatalf("JavaScript status = %s, want installed", got)
	}
}

func TestMuseSparkInstallerRunsWithBash(t *testing.T) {
	catalog, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	category, ok := findCategory(catalog, "agents")
	if !ok {
		t.Fatal("agents category missing")
	}
	item, ok := findItem(category, "muse-spark")
	if !ok {
		t.Fatal("muse-spark item missing")
	}
	install := item.Environments["default"].Install
	if len(install) != 1 {
		t.Fatalf("unexpected Muse Spark install commands: %v", install)
	}

	binDir := t.TempDir()
	marker := filepath.Join(binDir, "installed")
	fakeCurl := "#!/bin/sh\nprintf '%s\\n' '[[ -n \"$MUSE_TEST_MARKER\" ]] || exit 1' 'touch \"$MUSE_TEST_MARKER\"'\n"
	if err := os.WriteFile(filepath.Join(binDir, "curl"), []byte(fakeCurl), 0700); err != nil {
		t.Fatal(err)
	}
	process := exec.Command("sh", "-c", install[0])
	process.Env = append(os.Environ(), "PATH="+binDir+":"+os.Getenv("PATH"), "MUSE_TEST_MARKER="+marker)
	if output, err := process.CombinedOutput(); err != nil {
		t.Fatalf("run Muse Spark installer: %v: %s", err, output)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("Bash installer did not run: %v", err)
	}
}
