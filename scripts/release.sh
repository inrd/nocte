#!/bin/sh

set -eu

usage() {
	echo "usage: scripts/release.sh X.Y.Z [--push]" >&2
	exit 1
}

[ $# -ge 1 ] || usage

version="$1"
push=0

if ! printf '%s\n' "$version" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$'; then
	echo "release: version must be in X.Y.Z format" >&2
	exit 1
fi

shift

while [ $# -gt 0 ]; do
	case "$1" in
		--push)
			push=1
			;;
		*)
			usage
			;;
	esac
	shift
done

repo_root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
main_file="$repo_root/cmd/nocte/main.go"

cd "$repo_root"

if [ -n "$(git status --porcelain)" ]; then
	echo "release: working tree must be clean" >&2
	exit 1
fi

current_version=$(sed -n 's/^const version = "\(.*\)"$/\1/p' "$main_file")

if [ -z "$current_version" ]; then
	echo "release: could not find current version in cmd/nocte/main.go" >&2
	exit 1
fi

if [ "$current_version" = "$version" ]; then
	echo "release: version is already $version" >&2
	exit 1
fi

if git rev-parse -q --verify "refs/tags/v$version" >/dev/null 2>&1; then
	echo "release: tag v$version already exists" >&2
	exit 1
fi

tmp_file=$(mktemp "$repo_root/.release-main.go.XXXXXX")
trap 'rm -f "$tmp_file"' EXIT INT TERM HUP

awk -v version="$version" '
BEGIN {
	replaced = 0
}
/^const version = "/ {
	print "const version = \"" version "\""
	replaced = 1
	next
}
{
	print
}
END {
	if (replaced == 0) {
		exit 1
	}
}
' "$main_file" >"$tmp_file"

mv "$tmp_file" "$main_file"
trap - EXIT INT TERM HUP

make test

git add "$main_file"
git commit -m "Bump version to $version"
git tag "v$version"

if [ "$push" -eq 1 ]; then
	current_branch=$(git branch --show-current)
	if [ -z "$current_branch" ]; then
		echo "release: could not determine current branch for push" >&2
		exit 1
	fi

	git push origin "$current_branch"
	git push origin "v$version"
fi

echo "release: created commit and tag v$version"
