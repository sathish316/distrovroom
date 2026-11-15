package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
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
		newShowCommand(),
		newSearchCommand(),
		newCatalogCommandsCommand(),
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

func newShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show-item <category> <item>",
		Short: "Show item details and supported environments",
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
	distance int
}

func searchCatalog(catalog catalogFile, query string) []catalogMatch {
	items := make([]catalogMatch, 0)
	searchTerms := make([]string, 0)
	termItemIndexes := make([]int, 0)
	for _, category := range catalog.Categories {
		for _, item := range category.Items {
			itemIndex := len(items)
			items = append(items, catalogMatch{category: category, item: item})
			for _, name := range append([]string{item.Name}, item.Aliases...) {
				searchTerms = append(searchTerms, canonicalizeName(name))
				termItemIndexes = append(termItemIndexes, itemIndex)
			}
		}
	}

	bestDistances := make(map[int]int)
	for _, match := range fuzzy.RankFind(canonicalizeName(query), searchTerms) {
		itemIndex := termItemIndexes[match.OriginalIndex]
		if distance, ok := bestDistances[itemIndex]; !ok || match.Distance < distance {
			bestDistances[itemIndex] = match.Distance
		}
	}

	matches := make([]catalogMatch, 0, len(bestDistances))
	for itemIndex, match := range items {
		if distance, ok := bestDistances[itemIndex]; ok {
			match.distance = distance
			matches = append(matches, match)
		}
	}
	// Stable sorting preserves catalog order when distance, item name, and category are tied.
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].distance != matches[j].distance {
			return matches[i].distance < matches[j].distance
		}
		if matches[i].item.Name != matches[j].item.Name {
			return matches[i].item.Name < matches[j].item.Name
		}
		return matches[i].category.Name < matches[j].category.Name
	})
	return matches
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
	environments := make([]string, 0, len(item.Environments))
	for environment := range item.Environments {
		environments = append(environments, environment)
	}
	if len(environments) == 0 {
		environments = append(environments, "none")
	}
	if _, err := fmt.Fprintf(out, "Supported environments: %s\n", strings.Join(environments, ", ")); err != nil {
		return err
	}
	return nil
}
