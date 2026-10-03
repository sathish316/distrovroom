package main

import "github.com/spf13/cobra"

var catalogPath string

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "distro-vroom",
		Short:         "Browse software install commands across Linux distributions",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.PersistentFlags().StringVar(&catalogPath, "catalog", "", "path to a catalog YAML file")
	root.AddCommand(newCatalogCommand(), newSetupCommand())
	return root
}
