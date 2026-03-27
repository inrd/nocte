#!/bin/sh

set -eu

demo_root="/tmp/nocte-vhs-demo"
demo_home="$demo_root/home"
notes_dir="$demo_home/nocte"
fixtures_dir="$(dirname "$0")/fixtures/notes"

rm -rf "$demo_root"
mkdir -p "$notes_dir"
cp "$fixtures_dir"/*.md "$notes_dir"/
