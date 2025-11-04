#!/bin/bash
#
# Update the Unreleased section in CHANGELOG.md
# This script helps manage the transition from Unreleased to versioned releases
#
# Usage:
#   ./update-unreleased.sh rename <version>  # Rename Unreleased to version
#   ./update-unreleased.sh reset             # Create empty Unreleased section
#

set -e

COMMAND=$1
VERSION=$2
CHANGELOG_FILE="${3:-CHANGELOG.md}"

if [ -z "$COMMAND" ]; then
    echo "Usage: $0 <command> [version] [changelog-file]"
    echo ""
    echo "Commands:"
    echo "  rename <version>  - Rename Unreleased section to version number"
    echo "  reset             - Create new empty Unreleased section"
    echo ""
    echo "Examples:"
    echo "  $0 rename 1.2.0"
    echo "  $0 reset"
    exit 1
fi

if [ ! -f "$CHANGELOG_FILE" ]; then
    echo "Error: CHANGELOG file not found: $CHANGELOG_FILE"
    exit 1
fi

case "$COMMAND" in
    rename)
        if [ -z "$VERSION" ]; then
            echo "Error: Version required for rename command"
            echo "Usage: $0 rename <version>"
            exit 1
        fi
        
        # Remove 'v' prefix if present
        VERSION=$(echo "$VERSION" | sed 's/^v//')
        
        # Get current date
        DATE=$(date +"%B %d, %Y")
        
        # Replace "## Unreleased" with "## VERSION (DATE)"
        if grep -q "## Unreleased" "$CHANGELOG_FILE"; then
            if [[ "$OSTYPE" == "darwin"* ]]; then
                # macOS
                sed -i '' "s/## Unreleased/## ${VERSION} (${DATE})/" "$CHANGELOG_FILE"
            else
                # Linux
                sed -i "s/## Unreleased/## ${VERSION} (${DATE})/" "$CHANGELOG_FILE"
            fi
            echo "✓ Renamed Unreleased section to ${VERSION}"
        else
            echo "Warning: Unreleased section not found in $CHANGELOG_FILE"
            exit 1
        fi
        ;;
        
    reset)
        # Check if Unreleased section already exists
        if grep -q "## Unreleased" "$CHANGELOG_FILE"; then
            echo "Unreleased section already exists in $CHANGELOG_FILE"
            exit 0
        fi
        
        # Find the position after the header and before the first version
        # Insert new Unreleased section
        UNRELEASED_SECTION="## Unreleased\n\n### BREAKING CHANGES\n\n### FEATURES\n\n### ENHANCEMENTS\n\n### BUG FIXES\n\n### DEPRECATIONS\n\n### DOCUMENTATION\n\n"
        
        # Create temporary file with new Unreleased section
        awk -v section="$UNRELEASED_SECTION" '
            BEGIN { done=0 }
            /^## [0-9]/ && !done {
                printf "%s\n", section
                done=1
            }
            { print }
        ' "$CHANGELOG_FILE" > "$CHANGELOG_FILE.tmp"
        
        mv "$CHANGELOG_FILE.tmp" "$CHANGELOG_FILE"
        echo "✓ Created new Unreleased section"
        ;;
        
    *)
        echo "Error: Unknown command: $COMMAND"
        echo "Valid commands: rename, reset"
        exit 1
        ;;
esac

