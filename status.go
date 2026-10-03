package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type installationStatus string

const (
	installed    installationStatus = "installed"
	notInstalled installationStatus = "not-installed"
	unknown      installationStatus = "unknown"
)

func checkInstalled(item catalogItem) installationStatus {
	hasTest := false
	for _, commands := range item.Environments {
		if len(commands.Test) == 0 {
			continue
		}
		hasTest = true
		passed := true
		for _, testCommand := range commands.Test {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			command := exec.CommandContext(ctx, "/bin/sh", "-c", testCommand)
			command.Stdout = io.Discard
			command.Stderr = io.Discard
			if err := command.Run(); err != nil {
				passed = false
			}
			cancel()
			if !passed {
				break
			}
		}
		if passed {
			return installed
		}
	}
	if !hasTest {
		return unknown
	}
	return notInstalled
}

func newStatusCommand() *cobra.Command {
	var colorMode string
	status := &cobra.Command{Use: "status", Short: "Show installation status for selected items"}
	status.PersistentFlags().StringVar(&colorMode, "color", "auto", "color output: auto, always, or never")
	run := func(cmd *cobra.Command, args []string) error {
		if colorMode != "auto" && colorMode != "always" && colorMode != "never" {
			return fmt.Errorf("invalid --color value %q", colorMode)
		}
		config, catalog, selected, err := loadSelection()
		if err != nil {
			return err
		}
		if len(args) == 1 {
			if _, ok := findCategory(catalog, args[0]); !ok {
				return fmt.Errorf("category %q not found", args[0])
			}
		}
		useColor := shouldColor(cmd.OutOrStdout(), colorMode)
		out := cmd.OutOrStdout()
		for _, category := range config.Categories {
			if len(args) == 1 && canonicalizeName(category.Name) != canonicalizeName(args[0]) {
				continue
			}
			type itemStatus struct {
				name  string
				state installationStatus
			}
			var items []itemStatus
			count := 0
			for _, selection := range selected {
				if canonicalizeName(selection.category.Name) == canonicalizeName(category.Name) {
					state := checkInstalled(selection.item)
					items = append(items, itemStatus{name: selection.item.Name, state: state})
					if state == installed {
						count++
					}
				}
			}
			if _, err := fmt.Fprintf(out, "%s (%d/%d installed)\n", category.Name, count, len(items)); err != nil {
				return err
			}
			for _, item := range items {
				if _, err := fmt.Fprintf(out, "  %s  %s\n", colorize(string(item.state), item.state, useColor), item.name); err != nil {
					return err
				}
			}
		}
		return nil
	}
	status.RunE = run
	status.Args = cobra.NoArgs
	status.AddCommand(&cobra.Command{Use: "list [category]", Short: "Show all or one category", Args: cobra.MaximumNArgs(1), RunE: run})
	return status
}

func shouldColor(out io.Writer, mode string) bool {
	if mode == "always" {
		return true
	}
	if mode == "never" || os.Getenv("NO_COLOR") != "" || strings.EqualFold(os.Getenv("TERM"), "dumb") {
		return false
	}
	file, ok := out.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func colorize(label string, state installationStatus, enabled bool) string {
	if !enabled {
		return label
	}
	code := "33"
	if state == installed {
		code = "32"
	}
	if state == notInstalled {
		code = "31"
	}
	return "\x1b[" + code + "m" + label + "\x1b[0m"
}
