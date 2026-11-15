# Getting started

DistroVroom browses a catalog and runs install commands for items selected in your setup config.

## Build

Install Go 1.22 or later, then clone and build the repository:

```sh
git clone https://github.com/sathish316/distrovroom.git
cd distrovroom
./build.sh
```

The build script downloads the Go modules, tidies the module files, and writes the executable as `./distro-vroom`. To build without the script, run `go build -o ./distro-vroom .`.

## Create a setup config

Choose a starter file:

```sh
./distro-vroom configure init-from-sample
# Or start with SSH keys, GitHub CLI, and Codex:
./distro-vroom configure init-from-empty
```

The default path is `~/.config/distrovroom/config.yml`. The commands do not overwrite an existing file. Pass `--config-file path/to/config.yml` to create it elsewhere. The sample selects every catalog item; the minimal starter selects GitHub SSH keys, GitHub CLI, and Codex. Edit the environment and selections for your machine before running `apply`. Installing Codex requires npm.

## Browse the catalog

Run these commands from the repository directory:

```sh
./distro-vroom catalog list-categories
./distro-vroom catalog browse-category cli-programming
./distro-vroom catalog show-item cli-programming github-cli
./distro-vroom catalog search gh
```

Category and item lookups ignore letter case. Item names use lowercase hyphens; aliases such as `gh` are also accepted by `show-item`. The bundled catalog includes an `agents` category, with shared commands under the `default` environment.

## Use a custom catalog

The executable uses its embedded catalog by default, wherever you run it from. Pass `--catalog` to load a custom YAML file explicitly:

```sh
./distro-vroom --catalog ./my-catalog.yml catalog list-categories
```

A catalog has categories, items, and per-environment commands. The optional `test` list can hold commands for checking an installed item:

```yaml
categories:
  - name: cli-programming
    items:
      - name: github-cli
        aliases: [gh, "github cli"]
        environments:
          debian:
            install: ["sudo apt install gh"]
            upgrade: ["sudo apt install --only-upgrade gh"]
            config: ["gh auth login"]
            test: ["gh --version"]
```

The `install`, `upgrade`, `config`, and `test` command lists are catalog data. `apply` runs the install commands for the selected item or category.
