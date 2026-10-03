# DistroVroom

Distro Vroom is a utility to install your favourite apps, utils, programming environments in any Linux distro that you want to test-drive using a single config file in your dotfiles.

It was born from a need to switch my most frequently used apps, utils, env from Ubuntu Linux to Omarchy and then to other Arch based distros that i wanted to test-drive.

It's originally built for linux distros, but also happens to work for mac and windows.

# Config manual

Run `./distro-vroom configure init-from-sample` to create `~/.config/distrovroom/config.yml` with every catalog item, or `./distro-vroom configure init-from-empty` for GitHub SSH keys, GitHub CLI, and Codex. Both commands refuse to overwrite an existing file. Use `--config-file path/to/config.yml` for another location, then set `environment` to `arch` or `debian`.

Use `configure list`, `configure add <category> <item>`, and `configure remove <category> <item>` to manage selections. Run `catalog commands` to see commands, `status` to check installations, or `apply <category> [item]` to install selected items.

## Setup

- GitHub SSH keys

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
