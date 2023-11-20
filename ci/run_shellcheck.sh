#!/bin/sh

set -e

# Find and run shellcheck on all shell scripts.  Previously the `file` command was used
# to identify shell commands but that checks for a #! which some of our shell code does not
# use so go purely on filename.
find . \( -path ./.git -o -path ./venv -o -path ./build -o -path ./src/control/vendor \
 -o -path ./install -o -path src/rdb/raft -type d  \) -prune -o -name "*.sh" -exec \
 shellcheck -f gcc \{\} \+
