package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newApplyCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "apply <category> [item]",
		Short:   "Install one configured item or every item in a category",
		Example: "  distro-vroom apply setup github-ssh-keys --email you@example.com\n  distro-vroom apply cli-programming",
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return applyConfiguredItems(cmd, args)
		},
	}
	command.Flags().String("email", "", "email for catalog commands that request it")
	return command
}

func applyConfiguredItems(cmd *cobra.Command, args []string) error {
	config, catalog, selected, err := loadSelection()
	if err != nil {
		return err
	}
	if config.Environment == "" {
		return fmt.Errorf("setup config must set environment to arch or debian before applying")
	}
	category, ok := findCategory(catalog, args[0])
	if !ok {
		return fmt.Errorf("category %q is not in the catalog", args[0])
	}
	var selectedItems []catalogItem
	for _, selection := range selected {
		if canonicalizeName(selection.category.Name) == canonicalizeName(category.Name) {
			selectedItems = append(selectedItems, selection.item)
		}
	}
	if len(selectedItems) == 0 {
		return fmt.Errorf("category %q is not selected in your setup config", category.Name)
	}

	if len(args) == 2 {
		item, ok := findItem(category, args[1])
		if !ok {
			return fmt.Errorf("item %q is not in catalog category %q", args[1], category.Name)
		}
		configured := false
		for _, selectedItem := range selectedItems {
			if selectedItem.Name == item.Name {
				configured = true
				break
			}
		}
		if !configured {
			return fmt.Errorf("item %q is not selected in setup config category %q", item.Name, category.Name)
		}
		return applyItem(cmd, config.Environment, category, item)
	}

	for _, item := range selectedItems {
		if err := applyItem(cmd, config.Environment, category, item); err != nil {
			return err
		}
	}
	return nil
}

func applyItem(cmd *cobra.Command, environmentName string, category catalogCategory, item catalogItem) error {
	environment, ok := item.Environments[environmentName]
	if !ok {
		environment, ok = item.Environments["default"]
		if !ok {
			return fmt.Errorf("item %q in category %q does not support environment %q", item.Name, category.Name, environmentName)
		}
	}
	if len(environment.Install) == 0 {
		return fmt.Errorf("item %q in category %q has no install commands for %s", item.Name, category.Name, environmentName)
	}
	parameters := commandParameters(cmd)
	renderedCommands := make([]string, 0, len(environment.Install))
	for _, command := range environment.Install {
		rendered, err := renderCatalogCommand(command, parameters)
		if err != nil {
			return fmt.Errorf("render install command for %s/%s: %w", category.Name, item.Name, err)
		}
		renderedCommands = append(renderedCommands, rendered)
	}
	for _, command := range renderedCommands {
		if err := echoCommand(cmd, command); err != nil {
			return err
		}
		process := exec.Command("sh", "-c", command)
		process.Stdin = cmd.InOrStdin()
		process.Stdout = cmd.OutOrStdout()
		process.Stderr = cmd.ErrOrStderr()
		if err := process.Run(); err != nil {
			return fmt.Errorf("install %s/%s for %s: %w", category.Name, item.Name, environmentName, err)
		}
	}
	return nil
}

func echoCommand(cmd *cobra.Command, command string) error {
	out := cmd.OutOrStdout()
	if file, ok := out.(*os.File); ok && os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "" && os.Getenv("TERM") != "dumb" {
		if info, err := file.Stat(); err == nil && info.Mode()&os.ModeCharDevice != 0 {
			_, err = fmt.Fprintf(out, "\033[36m$ %s\033[0m\n", command)
			return err
		}
	}
	_, err := fmt.Fprintf(out, "$ %s\n", command)
	return err
}

// Pass every supplied Cobra flag to catalog templates as a shell-safe value.
func commandParameters(cmd *cobra.Command) map[string]string {
	parameters := make(map[string]string)
	collect := func(flag *pflag.Flag) {
		if value := flag.Value.String(); value != "" {
			parameters[strings.ReplaceAll(flag.Name, "-", "_")] = shellQuote(value)
		}
	}
	cmd.Flags().Visit(collect)
	cmd.InheritedFlags().Visit(collect)
	return parameters
}

func renderCatalogCommand(command string, parameters map[string]string) (string, error) {
	tmpl, err := template.New("catalog command").Option("missingkey=error").Parse(command)
	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	if err := tmpl.Execute(&output, parameters); err != nil {
		return "", err
	}
	return output.String(), nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
