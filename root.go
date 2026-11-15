package main

import "github.com/spf13/cobra"

var catalogPath string
var configPath string

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "distro-vroom",
		Short:         "Go from 0-100 on a new Linux distro by installing your favourite apps, CLIs, tools, and setup commands easily",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.PersistentFlags().StringVar(&catalogPath, "catalog", "", "path to a catalog YAML file")
	root.PersistentFlags().StringVar(&configPath, "config-file", "", "path to the setup config YAML file")
	root.AddCommand(newCatalogCommand(), newConfigCommand(), newStatusCommand())
	return root
}
