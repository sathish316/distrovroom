# DistroVroom

Distro Vroom is a utility to install your favourite apps, utils, programming environments in any Linux distro that you want to test-drive using a single config file in your dotfiles.

It was born from a need to switch my most frequently used apps, utils, env from Ubuntu Linux to Omarchy and then to other Arch based distros that i wanted to test-drive.

It's originally built for linux distros, but also happens to work for mac and windows.

# Config manual

Create the sample config with `distro-vroom config init-from-sample`, or create a minimal `setupconfig.empty.yml` in the current directory with `distro-vroom config init-empty`. The sample selects SSH keys, GitHub CLI, and all seven CLI agents. Config files contain only selected category and item names; install commands stay in the catalog.

Use `config list`, `config add <category> <item>`, and `config remove <category> <item>` to manage selections. Use `catalog commands` to see the catalog commands for those selections and `status` to check local installation state. See the [getting started guide](docs/getting-started.md) for examples.

## Setup

- ssh keys for github
- 

## Editors

- Zed
- Neovim
- Emacs

## Editor config and plugins

- Neovim
- Emacs

## Terminals and Terminal utils

- tmux
- ohmyzsh
- 

## Terminals and Terminal utils config

- tmux
- ohmyzsh

## Programming environments

- golang
- java
- 

## AI Agents

- Codex
- Pi

## Agent orchestration

- herdr
- cmux

## Apps - Productivity

- Todoist
- 

## CLIs - Productivity

## Apps - Programming

## CLIs - Programming

- gh

## Other config

- gitconfig

# Getting started

See the [getting started guide](docs/getting-started.md) for build, catalog, and custom-catalog instructions.
