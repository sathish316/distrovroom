#!/usr/bin/env bash
set -e

go mod tidy
go build -o ./distro-vroom .
