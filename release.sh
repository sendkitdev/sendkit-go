#!/bin/bash
set -e

CURRENT=$(git tag --sort=-v:refname | head -1)

if [ -z "$CURRENT" ]; then
  echo "No existing tags found. Cannot auto-increment."
  exit 1
fi

MAJOR=$(echo "$CURRENT" | cut -d. -f1)
MINOR=$(echo "$CURRENT" | cut -d. -f2)
PATCH=$(echo "$CURRENT" | cut -d. -f3)
VERSION="$MAJOR.$MINOR.$((PATCH + 1))"

echo "Current version: $CURRENT"
echo "New version: $VERSION"

if [ -n "$(git status --porcelain)" ]; then
  echo "Error: uncommitted changes. Commit or stash them first."
  exit 1
fi

git tag "$VERSION"
git push origin "$VERSION"

echo "Released $VERSION successfully!"
echo "The Go module proxy will pick it up automatically."
