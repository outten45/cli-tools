# cli-tools

A collection of command line tools mostly in Go, but could also include other
scripts later.

# Gitlab

There is a glab command that queries a gitlab instance on various things. See
[glab README](cmd/glab/README.md) for more info.

# Development

Use go 1.5 with the GO15VENDOREXPERIMENT=1 flag.

    set -x GO15VENDOREXPERIMENT 1

Install [glide](https://github.com/Masterminds/glide) for package management.

    brew install glide

Use glide to pull dependencies

    glide install
