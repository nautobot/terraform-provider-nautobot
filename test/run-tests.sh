#!/bin/sh
set -eu

go mod download

exec make testacc-run TEST=./internal/provider/...
