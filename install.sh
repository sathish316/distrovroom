#!/usr/bin/env bash
set -e

cd "$(dirname "$0")"
./build.sh
install -Dm755 ./distro-vroom "${HOME}/.local/bin/distro-vroom"
echo "Installed to ${HOME}/.local/bin/distro-vroom"
case ":${PATH}:" in
  *":${HOME}/.local/bin:"*) ;;
  *) echo "Note: ~/.local/bin is not on your PATH" ;;
esac
