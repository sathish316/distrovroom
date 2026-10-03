package main

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

func newCatalogCommandsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "commands [category] [item]",
		Short: "Show catalog commands for items selected in the setup config",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, catalog, selected, err := loadSelection()
			if err != nil {
				return err
			}
			if len(args) > 0 {
				if _, ok := findCategory(catalog, args[0]); !ok {
					return fmt.Errorf("category %q not found", args[0])
				}
			}
			if len(args) == 2 {
				category, _ := findCategory(catalog, args[0])
				if _, ok := findItem(category, args[1]); !ok {
					return fmt.Errorf("item %q not found in category %q", args[1], category.Name)
				}
			}
			shown := 0
			for _, selection := range selected {
				if len(args) > 0 && canonicalizeName(selection.category.Name) != canonicalizeName(args[0]) {
					continue
				}
				if len(args) == 2 {
					category, _ := findCategory(catalog, args[0])
					item, _ := findItem(category, args[1])
					if canonicalizeName(selection.item.Name) != canonicalizeName(item.Name) {
						continue
					}
				}
				if shown > 0 {
					if _, err := fmt.Fprintln(cmd.OutOrStdout()); err != nil {
						return err
					}
				}
				if err := printCatalogCommands(cmd, selection); err != nil {
					return err
				}
				shown++
			}
			if shown == 0 {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "No selected items found.")
				return err
			}
			return nil
		},
	}
}

func printCatalogCommands(cmd *cobra.Command, selection selectedItem) error {
	out := cmd.OutOrStdout()
	if _, err := fmt.Fprintf(out, "%s/%s\n", selection.category.Name, selection.item.Name); err != nil {
		return err
	}
	names := make([]string, 0, len(selection.item.Environments))
	for name := range selection.item.Environments {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		commands := selection.item.Environments[name]
		if _, err := fmt.Fprintf(out, "  %s:\n", name); err != nil {
			return err
		}
		for _, group := range []struct {
			name  string
			lines []string
		}{
			{"install", commands.Install}, {"upgrade", commands.Upgrade}, {"config", commands.Config}, {"test", commands.Test},
		} {
			for _, line := range group.lines {
				if _, err := fmt.Fprintf(out, "    %s: %s\n", group.name, line); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
