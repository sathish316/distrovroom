# Getting started

DistroVroom is a Go CLI for selecting tools in a setup config, browsing their install guidance, and checking which selected tools are present. Install, upgrade, and configuration commands are printed for review. Status runs only the catalog test commands.

## Build

Install Go 1.22 or later, then clone and build the repository:

```sh
git clone https://github.com/sathish316/distrovroom.git
cd distrovroom
./build.sh
```

The executable is `./distro-vroom`. You can also run `go build -o ./distro-vroom .`.

## Create a setup config

```sh
./distro-vroom config init-from-sample
```

This creates `~/.config/distrovroom/setupconfig.yml` on Linux (or the directory selected by `XDG_CONFIG_HOME`). It copies the bundled sample and refuses to overwrite an existing file. The sample selects SSH keys, GitHub CLI, and seven CLI coding agents.

For a small starting point:

```sh
./distro-vroom config init-empty
```

This creates `setupconfig.empty.yml` in the current directory with only SSH keys, GitHub CLI, and Codex. Use it directly with `--config-file setupconfig.empty.yml`, or move it to your active config path. It also refuses to overwrite an existing file.

The config lists category and item names only. Installation, upgrade, configuration, and test commands live in the catalog once and are resolved from those names:

```yaml
categories:
  - name: agents
    items:
      - codex
      - pi
```

## Manage selected items

```sh
./distro-vroom config list
./distro-vroom config list agents
./distro-vroom config add agents claude-code
./distro-vroom config remove agents claude-code
```

Item aliases work with `add` and `remove`. For example, `gh` resolves to `github-cli`. An unrecognized category or item is rejected, and duplicate selections are not added.

## Browse the catalog and selected commands

```sh
./distro-vroom catalog list-categories
./distro-vroom catalog browse-category agents
./distro-vroom catalog show-item agents codex
./distro-vroom catalog search cursor
./distro-vroom catalog commands
./distro-vroom catalog commands agents
./distro-vroom catalog commands agents codex
./distro-vroom catalog commands setup ssh-keys --email you@example.com
```

The browse and search commands show all available catalog entries so you can discover items to add. `catalog commands` reads the setup config and shows commands only for selected entries. Commands are grouped by environment; `default` applies across supported systems. The SSH key command uses a Go template and fills its email placeholder from `--email`. Without that flag, the output shows `EMAIL_ADDRESS` as a placeholder. The commands are guidance to review before running in a shell.

## Check status

```sh
./distro-vroom status
./distro-vroom status list agents
./distro-vroom status list agents --color=never
```

Status checks only selected items by running their catalog `test` commands and checking the exit code. The bundled tests use `command -v` for CLIs and look for an SSH public key in `~/.ssh`; they do not contact services. Colors appear on a terminal by default: green means installed, red means not installed, and yellow means the catalog has no test command. Use `--color=always` or `--color=never` to override.

## Custom files

Use `--config-file` to select another setup config and `--catalog` to load a custom catalog:

```sh
./distro-vroom --config-file ./setupconfig.empty.yml config list
./distro-vroom --catalog ./my-catalog.yml catalog list-categories
```

A catalog contains categories, items, and commands by environment. The `test` commands determine status: an exit code of zero means installed, a nonzero code means not installed, and no test means unknown. Custom catalog test commands are executed by `status`, so review them before using a custom catalog.
