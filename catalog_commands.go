package main

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
)

func newCatalogCommand() *cobra.Command {
	catalog := &cobra.Command{
		Use:   "catalog",
		Short: "Browse and search the software catalog",
	}
	catalog.AddCommand(
		newListCategoriesCommand(),
		newBrowseCategoryCommand(),
		newShowItemCommand(),
		newSearchCommand(),
	)
	return catalog
}

func newListCategoriesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-categories",
		Short: "List catalog categories",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			catalog, err := loadCatalog()
			if err != nil {
				return err
			}
			for _, category := range catalog.Categories {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), category.Name); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func newBrowseCategoryCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "browse-category <category>",
		Short: "Show the items in a catalog category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			catalog, err := loadCatalog()
			if err != nil {
				return err
			}
			category, ok := findCategory(catalog, args[0])
			if !ok {
				return fmt.Errorf("category %q not found", args[0])
			}

			out := cmd.OutOrStdout()
			if _, err := fmt.Fprintf(out, "%s\n", category.Name); err != nil {
				return err
			}
			for _, item := range category.Items {
				if _, err := fmt.Fprintf(out, "  %s", item.Name); err != nil {
					return err
				}
				if item.Description != "" {
					if _, err := fmt.Fprintf(out, " — %s", item.Description); err != nil {
						return err
					}
				}
				if _, err := fmt.Fprintln(out); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func newShowItemCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show-item <category> <item>",
		Short: "Show item details and commands for each environment",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			catalog, err := loadCatalog()
			if err != nil {
				return err
			}
			category, ok := findCategory(catalog, args[0])
			if !ok {
				return fmt.Errorf("category %q not found", args[0])
			}
			item, ok := findItem(category, args[1])
			if !ok {
				return fmt.Errorf("item %q not found in category %q", args[1], category.Name)
			}
			return printItemDetails(cmd.OutOrStdout(), category, item)
		},
	}
}

func newSearchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Find the five closest item names in the catalog",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.TrimSpace(strings.Join(args, " "))
			if query == "" {
				return fmt.Errorf("search query cannot be empty")
			}
			catalog, err := loadCatalog()
			if err != nil {
				return err
			}
			matches := searchCatalog(catalog, query)
			if len(matches) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No catalog items found.")
				return nil
			}
			limit := min(5, len(matches))
			out := cmd.OutOrStdout()
			if _, err := fmt.Fprintln(out, "ITEM\tCATEGORY"); err != nil {
				return err
			}
			for _, match := range matches[:limit] {
				if _, err := fmt.Fprintf(out, "%s\t%s\n", match.item.Name, match.category.Name); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

type catalogMatch struct {
	category catalogCategory
	item     catalogItem
	score    float64
}

func searchCatalog(catalog catalogFile, query string) []catalogMatch {
	matches := make([]catalogMatch, 0)
	for _, category := range catalog.Categories {
		for _, item := range category.Items {
			score := fuzzyScore(query, item.Name)
			for _, alias := range item.Aliases {
				if aliasScore := fuzzyScore(query, alias); aliasScore > score {
					score = aliasScore
				}
			}
			matches = append(matches, catalogMatch{category: category, item: item, score: score})
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		if matches[i].item.Name != matches[j].item.Name {
			return matches[i].item.Name < matches[j].item.Name
		}
		return matches[i].category.Name < matches[j].category.Name
	})
	return matches
}

func fuzzyScore(query, candidate string) float64 {
	query = normalizedText(query)
	candidate = normalizedText(candidate)
	if query == "" || candidate == "" {
		return 0
	}
	if query == candidate {
		return 1
	}

	queryLength := utf8.RuneCountInString(query)
	candidateLength := utf8.RuneCountInString(candidate)
	longest := max(queryLength, candidateLength)
	distance := levenshteinDistance(query, candidate)
	return 1 - float64(distance)/float64(longest)
}

func levenshteinDistance(left, right string) int {
	leftRunes := []rune(left)
	rightRunes := []rune(right)
	previous := make([]int, len(rightRunes)+1)
	current := make([]int, len(rightRunes)+1)
	for column := range previous {
		previous[column] = column
	}
	for row, leftRune := range leftRunes {
		current[0] = row + 1
		for column, rightRune := range rightRunes {
			cost := 0
			if leftRune != rightRune {
				cost = 1
			}
			current[column+1] = min(
				current[column]+1,
				previous[column+1]+1,
				previous[column]+cost,
			)
		}
		previous, current = current, previous
	}
	return previous[len(rightRunes)]
}

func printItemDetails(out io.Writer, category catalogCategory, item catalogItem) error {
	if _, err := fmt.Fprintf(out, "Item: %s\nCategory: %s\n", item.Name, category.Name); err != nil {
		return err
	}
	if item.Description != "" {
		if _, err := fmt.Fprintf(out, "Description: %s\n", item.Description); err != nil {
			return err
		}
	}
	if len(item.Aliases) > 0 {
		if _, err := fmt.Fprintf(out, "Aliases: %s\n", strings.Join(item.Aliases, ", ")); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(out, "Environments:"); err != nil {
		return err
	}
	environments := make([]string, 0, len(item.Environments))
	for environment := range item.Environments {
		environments = append(environments, environment)
	}
	sort.Strings(environments)
	for _, environment := range environments {
		if _, err := fmt.Fprintf(out, "  %s:\n", environment); err != nil {
			return err
		}
		commands := item.Environments[environment]
		for _, commandGroup := range []struct {
			name     string
			commands []string
		}{
			{name: "install", commands: commands.Install},
			{name: "upgrade", commands: commands.Upgrade},
			{name: "config", commands: commands.Config},
		} {
			if _, err := fmt.Fprintf(out, "    %s:\n", commandGroup.name); err != nil {
				return err
			}
			if len(commandGroup.commands) == 0 {
				if _, err := fmt.Fprintln(out, "      (not specified)"); err != nil {
					return err
				}
				continue
			}
			for _, command := range commandGroup.commands {
				if _, err := fmt.Fprintf(out, "      - %s\n", command); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
