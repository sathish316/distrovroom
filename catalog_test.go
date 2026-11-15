package main

import "testing"

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
