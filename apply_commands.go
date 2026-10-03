package main

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

type plannedCommand struct {
	category string
	item     string
	action   string
	commands []string
}

func newApplyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "apply",
		Short: "Run the commands selected in your config",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runConfiguredCommands(cmd, "")
		},
	}
}

func newSetupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "setup <item>",
		Short: "Set up tools",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfiguredCommands(cmd, args[0])
		},
	}
}

func runConfiguredCommands(cmd *cobra.Command, setupItem string) error {
	catalog, err := loadCatalog()
	if err != nil {
		return err
	}
	config, err := loadSetupConfig(configPath)
	if err != nil {
		return err
	}
	// Validate every selection before running any shell commands.
	plan, err := planConfiguredCommands(catalog, config)
	if err != nil {
		return err
	}

	if setupItem != "" {
		category, ok := findCategory(catalog, "setup")
		if !ok {
			return fmt.Errorf("setup category is not in the catalog")
		}
		item, ok := findItem(category, setupItem)
		if !ok {
			return fmt.Errorf("setup item %q is not in the catalog", setupItem)
		}
		selected := make([]plannedCommand, 0, 1)
		for _, step := range plan {
			if step.category == category.Name && step.item == item.Name && step.action == "config" {
				selected = append(selected, step)
			}
		}
		if len(selected) == 0 {
			return fmt.Errorf("setup item %q is not selected with action config in your setup config", item.Name)
		}
		plan = selected
	}

	for _, step := range plan {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Running %s/%s %s commands for %s\n", step.category, step.item, step.action, config.Environment); err != nil {
			return err
		}
		for _, command := range step.commands {
			process := exec.Command("sh", "-c", command)
			process.Stdin = cmd.InOrStdin()
			process.Stdout = cmd.OutOrStdout()
			process.Stderr = cmd.ErrOrStderr()
			if err := process.Run(); err != nil {
				return fmt.Errorf("run %s/%s %s commands for %s: %w", step.category, step.item, step.action, config.Environment, err)
			}
		}
	}
	return nil
}

func planConfiguredCommands(catalog catalogFile, config setupConfig) ([]plannedCommand, error) {
	plan := make([]plannedCommand, 0, len(config.Commands))
	for _, selection := range config.Commands {
		category, ok := findCategory(catalog, selection.Category)
		if !ok {
			return nil, fmt.Errorf("category %q is not in the catalog", selection.Category)
		}
		item, ok := findItem(category, selection.Item)
		if !ok {
			return nil, fmt.Errorf("item %q is not in catalog category %q", selection.Item, category.Name)
		}
		environment, ok := item.Environments[config.Environment]
		if !ok {
			return nil, fmt.Errorf("item %q in category %q does not support environment %q", item.Name, category.Name, config.Environment)
		}
		commands := commandsForAction(environment, selection.Action)
		if len(commands) == 0 {
			return nil, fmt.Errorf("item %q in category %q has no %s commands for %s", item.Name, category.Name, selection.Action, config.Environment)
		}
		plan = append(plan, plannedCommand{
			category: category.Name,
			item:     item.Name,
			action:   selection.Action,
			commands: commands,
		})
	}
	return plan, nil
}

func commandsForAction(commands environmentCommands, action string) []string {
	switch action {
	case "install":
		return commands.Install
	case "upgrade":
		return commands.Upgrade
	case "config":
		return commands.Config
	case "test":
		return commands.Test
	default:
		return nil
	}
}
