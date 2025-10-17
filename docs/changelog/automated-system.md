# Automated Changelog System

## Overview

The Rafay Terraform Provider uses an AI-powered automated changelog system that maintains the `CHANGELOG.md` file across master and release branches. This system ensures consistent, professional documentation of all changes following Terraform provider best practices.

## Key Features

- **AI-Powered Categorization**: Uses Claude AI to intelligently categorize and describe changes
- **Automatic Deprecation Detection**: Scans Go code for `Deprecated` and `DeprecationMessage` fields
- **Branch-Aware**: Handles both master branch (Unreleased) and release branches
- **Cherry-Pick Support**: Works seamlessly with the existing cherry-pick workflow
- **GitHub Release Integration**: Automatically generates GitHub Release Notes

## How It Works

### 1. PR Merge to Master Branch

When a PR is merged to the `master` branch:

1. **GitHub Action Triggers**: The `changelog-on-merge.yml` workflow activates
2. **Deprecation Scanning**: Go code changes are scanned for deprecation warnings
3. **Commit Analysis**: Claude AI analyzes commit messages and changes
4. **Categorization**: Changes are categorized into:
   - BREAKING CHANGES
   - FEATURES
   - ENHANCEMENTS
   - BUG FIXES
   - DEPRECATIONS
   - DOCUMENTATION
5. **CHANGELOG Update**: Entries are added to the "Unreleased" section
6. **Auto-Commit**: Changes are committed and pushed back to master

### 2. PR Merge to Release Branch

When a PR is cherry-picked and merged to a release branch (e.g., `v1.2.0`):

- Same process as master, but entries are added to the version section (e.g., `## 1.2.0`)
- Duplicate detection prevents the same PR from appearing multiple times

### 3. Branch Cut Process

When creating a new release branch using the `branch-cut.yaml` workflow:

1. New release branch is created from master
2. "Unreleased" section is renamed to the version number with date
3. The release branch now has the version section ready for additional cherry-picks

### 4. Release Process

When a tag is created on a release branch:

1. Release notes are extracted from the CHANGELOG.md version section
2. GoReleaser creates a GitHub Release with these notes
3. The full CHANGELOG.md is included in release artifacts

## System Components

### Files and Their Purpose

#### Scripts
- **`scripts/generate-changelog.py`** - AI-powered changelog generator
- **`scripts/scan-deprecations.go`** - Go AST parser that detects deprecation warnings
- **`scripts/extract-release-notes.sh`** - Extracts version-specific section from CHANGELOG
- **`scripts/update-unreleased.sh`** - Manages Unreleased section transitions
- **`scripts/requirements.txt`** - Python dependencies

#### GitHub Actions
- **`.github/workflows/changelog-on-merge.yml`** - Main automation workflow
- **`.github/workflows/release.yml`** - Release process with changelog integration
- **`.github/workflows/branch-cut.yaml`** - Branch cut with CHANGELOG handling

#### Configuration
- **`.github/changelog-config.json`** - AI model settings and categorization rules
- **`CHANGELOG.md`** - Main changelog file (root level)

#### Documentation
- **`docs/changelog/automated-system.md`** - This file
- **`docs/changelog/commit-guidelines.md`** - Best practices for commit messages
- **`docs/changelog/testing-guide.md`** - How to test the system

## CHANGELOG Structure

### Master Branch (Accumulating Changes)

```markdown
## Unreleased

### FEATURES
* **New Resource:** `rafay_environment_template` for managing environment templates ([#456](link))
* resource/rafay_eks_cluster: Add support for IPv6 networking ([#445](link))

### ENHANCEMENTS
* resource/rafay_aks_cluster: Improve node pool scaling performance ([#448](link))

### BUG FIXES
* resource/rafay_eks_cluster: Fix state drift when node groups modified externally ([#447](link))

### DEPRECATIONS
* resource/rafay_aks_cluster: Deprecate `project_id` argument ([#449](link))

## 1.1.51 (October 15, 2024)
...
```

### Release Branch (After Branch Cut)

```markdown
## 1.2.0 (January 15, 2025)

### FEATURES
* **New Resource:** `rafay_environment_template` for managing environment templates ([#456](link))
* resource/rafay_eks_cluster: Add support for IPv6 networking ([#445](link))

### ENHANCEMENTS
* resource/rafay_aks_cluster: Improve node pool scaling performance ([#448](link))
* resource/rafay_aks_cluster: Add subnet configuration validation ([#460](link))  # Cherry-picked PR

## 1.1.51 (October 15, 2024)
...
```

## AI Categorization Rules

The system uses Claude AI with specific rules to categorize changes:

### BREAKING CHANGES
**Only for user-facing breaking changes:**
- Removing or renaming resources
- Removing or renaming resource arguments/attributes
- Changing required vs optional status
- Changing default values affecting existing deployments

**NOT for:** removing comments, renaming variables, refactoring code, internal changes

### FEATURES
- New resources (`rafay_*`)
- New data sources
- New optional arguments adding capabilities
- Major new functionality

### ENHANCEMENTS
- Performance improvements
- Better error messages
- Additional validation
- Support for new cloud provider features

### BUG FIXES
- Fixes for crashes, errors, or incorrect results
- State management corrections
- Import/export fixes

### DEPRECATIONS
- Deprecated resources, arguments, or values
- Automatically detected from Go code
- Should include migration path

### DOCUMENTATION
- Only significant changes (new guides, major rewrites)
- Minor typo fixes are skipped

## Configuration

### Environment Variables (GitHub Secrets)

Required:
- `CLAUDE_API_KEY` or `ANTHROPIC_API_KEY` - For AI-powered changelog generation
- `GITHUB_TOKEN` - Automatically provided by GitHub Actions
- `JENKINS_PAT` - For branch cut workflow (if using Jenkins integration)
- `RCTL_GO_MODULES_TOKEN` - For accessing private Go modules

### Customization

Edit `.github/changelog-config.json` to customize:
- AI model version
- Maximum commits per PR
- Category names
- Keyword weights
- Skip patterns

## Troubleshooting

### Changelog Not Updated

**Check:**
1. Was the PR actually merged (not just closed)?
2. Check the GitHub Actions run for errors
3. Verify `CLAUDE_API_KEY` is set in repository secrets
4. Check if PR has `skip-changelog` label

### Incorrect Categorization

**Solutions:**
1. Use clearer commit messages with specific keywords
2. Add PR description with explicit categorization hints
3. Adjust categorization rules in `.github/changelog-config.json`
4. Manually edit CHANGELOG.md if needed (will not be overwritten)

### Duplicate Entries

The system detects duplicates by PR number. If you see duplicates:
1. Check if the PR number is included in entries
2. Verify the duplicate detection logic in `generate-changelog.py`

### Missing Deprecations

**Check:**
1. Is the `Deprecated` or `DeprecationMessage` field correctly formatted in Go code?
2. Check the deprecation scanner output in GitHub Actions logs
3. Verify Go file was actually changed in the PR

## Manual Operations

### Manually Trigger Changelog Generation

You can test the changelog generator locally:

```bash
# Install dependencies
pip install -r scripts/requirements.txt

# Set API key
export CLAUDE_API_KEY="your-key-here"

# Generate changelog (dry run)
python3 scripts/generate-changelog.py \
  --pr-number 123 \
  --pr-url "https://github.com/RafaySystems/terraform-provider-rafay/pull/123" \
  --base-ref origin/master \
  --head-ref HEAD \
  --dry-run

# Actually update CHANGELOG.md
python3 scripts/generate-changelog.py \
  --pr-number 123 \
  --pr-url "https://github.com/RafaySystems/terraform-provider-rafay/pull/123" \
  --base-ref origin/master \
  --head-ref HEAD \
  --target-section "Unreleased"
```

### Manually Update Unreleased Section

```bash
# Rename Unreleased to version (for release branches)
bash scripts/update-unreleased.sh rename 1.2.0

# Create new Unreleased section (after branch cut on master)
bash scripts/update-unreleased.sh reset
```

### Extract Release Notes

```bash
# Extract notes for a specific version
bash scripts/extract-release-notes.sh 1.2.0 > release-notes.md
```

## Benefits

1. **Consistency**: AI ensures uniform style and quality
2. **No Manual Work**: Automatic on every PR merge
3. **No Label Requirements**: Works with any commit style
4. **Deprecation Tracking**: Never miss a deprecation warning
5. **Branch Sync**: Handles master and release branches automatically
6. **Professional Quality**: Follows Terraform AWS provider standards

## Maintenance

### Updating AI Model

Edit `.github/changelog-config.json`:
```json
{
  "ai_model": "claude-3-5-sonnet-20250201"  // Update to newer model
}
```

### Adjusting Categorization

Edit the categorization rules in `.github/changelog-config.json` or adjust the prompt in `scripts/generate-changelog.py`.

### Updating Python Dependencies

```bash
# Update requirements.txt
pip install --upgrade anthropic requests python-dotenv

# Regenerate requirements.txt
pip freeze | grep -E 'anthropic|requests|python-dotenv' > scripts/requirements.txt
```

## Support

For issues or questions:
- Check [Testing Guide](./testing-guide.md) for validation steps
- Review [Commit Guidelines](./commit-guidelines.md) for best practices
- Open an issue in the repository with `changelog` label

---

**Last Updated**: January 2025  
**System Version**: 1.0

