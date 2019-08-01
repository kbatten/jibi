#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
orgdir="$workspace/src/github.com/kbatten"
if [ ! -L "$orgdir/jibi" ]; then
    mkdir -p "$orgdir"
    cd "$orgdir"
    ln -s ../../../../../. jibi
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

# Run the command inside the workspace.
cd "$orgdir/jibi"
PWD="$orgdir/jibi"

# Launch the arguments with the configured environment.
exec "$@"
