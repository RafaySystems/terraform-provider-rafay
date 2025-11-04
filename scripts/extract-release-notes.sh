#!/bin/bash
#
# Extract release notes from CHANGELOG.md for a specific version
# Usage: ./extract-release-notes.sh <version>
# Example: ./extract-release-notes.sh 1.2.0
#

set -e

VERSION=$1
CHANGELOG_FILE="${2:-CHANGELOG.md}"

if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version> [changelog-file]"
    echo "Example: $0 1.2.0"
    exit 1
fi

if [ ! -f "$CHANGELOG_FILE" ]; then
    echo "Error: CHANGELOG file not found: $CHANGELOG_FILE"
    exit 1
fi

# Remove 'v' prefix if present
VERSION=$(echo "$VERSION" | sed 's/^v//')

# Extract the section for this version
# This extracts everything between "## VERSION" and the next "##" heading
awk -v version="$VERSION" '
    /^## / {
        if (found) exit
        if ($0 ~ "## "version) {
            found=1
            next
        }
    }
    found { print }
' "$CHANGELOG_FILE"

# Check if we found the version
if ! grep -q "## $VERSION" "$CHANGELOG_FILE"; then
    echo "Warning: Version $VERSION not found in $CHANGELOG_FILE" >&2
    exit 1
fi

