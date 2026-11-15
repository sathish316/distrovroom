# Setup commands

DistroVroom reads a private setup config to choose catalog commands for one Linux environment. Each entry selects a catalog category, item, and action.

## Select commands

Copy the sample config to the default location and edit it:

```sh
mkdir -p ~/.config/distrovroom
cp setupconfig.sample.yml ~/.config/distrovroom/mysetupconfig.yml
```

The sample selects GitHub SSH key setup for Arch Linux:

```yaml
environment: arch # or debian
commands:
  - category: setup
    item: github-ssh-keys
    action: config
```

The category and item identify the command and subcommand in the catalog. Add or remove entries to choose the commands for your environment. Actions are `install`, `upgrade`, `config`, and optional `test`. Pass `--config /path/to/config.yml` to use a different file.

## Apply the config

Run all selected catalog commands in order, or run one configured setup item:

```sh
distro-vroom apply
distro-vroom setup github-ssh-keys
```

Catalog entries contain shell commands. Review them before applying a config. Interactive prompts and command output are passed through to the terminal.

## Add the GitHub SSH key

The `setup/github-ssh-keys` config action reuses `~/.ssh/id_ed25519` when present, creates a key only if it is missing, checks that the public key matches the private key, and adds the key to a responding `ssh-agent`. It starts an agent if the configured socket is stale or absent and prints the export commands needed to reuse that agent in the current shell. It then prints the public key.

Copy the displayed key to [GitHub SSH key settings](https://github.com/settings/ssh/new). Set the config `environment` to `arch` or `debian`; install OpenSSH first if needed.
