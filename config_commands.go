package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newConfigCommand() *cobra.Command {
	config := &cobra.Command{Use: "configure", Aliases: []string{"config"}, Short: "Initialize and edit selected setup items"}
	config.AddCommand(
		&cobra.Command{
			Use: "init-from-empty", Short: "Copy the bundled minimal config to the active config path", Args: cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				path, err := activeConfigPath()
				if err != nil {
					return err
				}
				if err := writeNewFile(path, emptyConfig); err != nil {
					return err
				}
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", path)
				return err
			},
		},
		&cobra.Command{
			Use: "init-from-sample", Short: "Copy the bundled sample to the active config path", Args: cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				path, err := activeConfigPath()
				if err != nil {
					return err
				}
				if err := writeNewFile(path, sampleConfig); err != nil {
					return err
				}
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", path)
				return err
			},
		},
		&cobra.Command{
			Use: "list [category]", Short: "List selected items in all or one category", Args: cobra.MaximumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				config, catalog, selected, err := loadSelection()
				if err != nil {
					return err
				}
				if len(args) == 1 {
					if _, ok := findCategory(catalog, args[0]); !ok {
						return fmt.Errorf("category %q not found", args[0])
					}
				}
				out := cmd.OutOrStdout()
				for _, entry := range config.Categories {
					if len(args) == 1 && canonicalizeName(entry.Name) != canonicalizeName(args[0]) {
						continue
					}
					if _, err := fmt.Fprintln(out, entry.Name); err != nil {
						return err
					}
					for _, item := range selected {
						if canonicalizeName(item.category.Name) == canonicalizeName(entry.Name) {
							if _, err := fmt.Fprintf(out, "  %s\n", item.item.Name); err != nil {
								return err
							}
						}
					}
				}
				return nil
			},
		},
		&cobra.Command{
			Use: "add <category> <item>", Short: "Select an item from the catalog", Args: cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				config, catalog, _, err := loadSelection()
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
				for index := range config.Categories {
					if canonicalizeName(config.Categories[index].Name) != canonicalizeName(category.Name) {
						continue
					}
					for _, existing := range config.Categories[index].Items {
						existingItem, _ := findItem(category, existing)
						if canonicalizeName(existingItem.Name) == canonicalizeName(item.Name) {
							return fmt.Errorf("%s/%s is already selected", category.Name, item.Name)
						}
					}
					config.Categories[index].Items = append(config.Categories[index].Items, item.Name)
					if err := saveConfig(config); err != nil {
						return err
					}
					_, err = fmt.Fprintf(cmd.OutOrStdout(), "Added %s/%s\n", category.Name, item.Name)
					return err
				}
				config.Categories = append(config.Categories, configCategory{Name: category.Name, Items: []string{item.Name}})
				if err := saveConfig(config); err != nil {
					return err
				}
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Added %s/%s\n", category.Name, item.Name)
				return err
			},
		},
		&cobra.Command{
			Use: "remove <category> <item>", Short: "Remove a selected item", Args: cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				config, catalog, _, err := loadSelection()
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
				for i := range config.Categories {
					if canonicalizeName(config.Categories[i].Name) != canonicalizeName(category.Name) {
						continue
					}
					for j, existing := range config.Categories[i].Items {
						existingItem, _ := findItem(category, existing)
						if canonicalizeName(existingItem.Name) != canonicalizeName(item.Name) {
							continue
						}
						config.Categories[i].Items = append(config.Categories[i].Items[:j], config.Categories[i].Items[j+1:]...)
						if len(config.Categories[i].Items) == 0 {
							config.Categories = append(config.Categories[:i], config.Categories[i+1:]...)
						}
						if err := saveConfig(config); err != nil {
							return err
						}
						_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed %s/%s\n", category.Name, item.Name)
						return err
					}
				}
				return fmt.Errorf("%s/%s is not selected", category.Name, item.Name)
			},
		},
	)
	return config
}

func loadSelection() (setupConfig, catalogFile, []selectedItem, error) {
	catalog, err := loadCatalog()
	if err != nil {
		return setupConfig{}, catalogFile{}, nil, err
	}
	config, err := loadConfig()
	if err != nil {
		return setupConfig{}, catalogFile{}, nil, err
	}
	selected, err := resolveConfig(config, catalog)
	if err != nil {
		return setupConfig{}, catalogFile{}, nil, err
	}
	return config, catalog, selected, nil
}
