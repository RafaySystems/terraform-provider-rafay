# Changelog Generation System

## Overview

The Rafay Terraform Provider uses an AI-powered changelog generation system that maintains the `CHANGELOG.md` file across master and release branches. This system ensures consistent, professional documentation of all changes following Terraform provider best practices.

## Key Features

- **AI-Powered Categorization**: Uses OpenAI GPT models to intelligently categorize and describe changes
- **Deprecation Detection**: Scans Go code for `Deprecated` and `DeprecationMessage` fields
- **Branch-Aware**: Handles both master branch (Unreleased) and release branches

## How It Works

### 1. PR Merge to Master Branch

**Note:** Due to branch protection rules, changelog generation must be done manually prior to PR merge.

Before a PR is merged to the `master` branch:

1. **Manual Script Execution**: Run the changelog generation script (see [Manual Operations](#manual-operations))
2. **Deprecation Scanning**: Go code changes are scanned for deprecation warnings (optional, via `--deprecations-file`)
3. **Commit Analysis**: OpenAI GPT analyzes commit messages and changes
4. **Categorization**: Changes are categorized into:
   - BREAKING CHANGES
   - FEATURES
   - ENHANCEMENTS
   - BUG FIXES
   - DEPRECATIONS
   - DOCUMENTATION
5. **Fragment Creation**: Entries are written to `.changelog/{PR_NUMBER}.txt`
6. **CHANGELOG Update**: Entries are added to the "Unreleased" section in `CHANGELOG.md`
7. **Manual Commit**: Review, commit, and push the changes

### 2. PR Merge to Release Branch

When a PR is cherry-picked and merged to a release branch (e.g., `v1.2.0`):

- Same manual process as master, but specify `--target-section` with the version number (e.g., `1.2.0`)
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

### 5. Changelog Fragments Workflow

The system maintains individual changelog fragments in `.changelog/` for audit trail and traceability:

**On PR Merge (Manual Process):**
```
PR #1131 created
    ↓
Run changelog generation script manually
    ↓
Generate changelog entries
    ↓
Write to .changelog/1131.txt    ← Individual fragment
    ↓
Update CHANGELOG.md Unreleased  ← Consolidated changelog
    ↓
Review and commit both files
```

**Fragment File Example** (`.changelog/1131.txt`):
```markdown
### FEATURES

* **New Resource:** `rafay_eks_pod_identity`: Support for EKS Pod Identity ([#1131](URL))

### ENHANCEMENTS

* resource/rafay_eks_cluster_spec: Add Pod Identity configuration support ([#1131](URL))
```

**Benefits:**
- **Audit Trail**: Track exactly what changed in each PR
- **Easy Review**: Review individual PR changes before release
- **Revert Capability**: Easy to identify and revert specific PR's changes
- **Release Compilation**: Can rebuild CHANGELOG.md from fragments if needed

## System Components

### Files and Their Purpose

#### Scripts
- **`scripts/generate-changelog.py`** - AI-powered changelog generator
- **`scripts/scan-deprecations.go`** - Go AST parser that detects deprecation warnings
- **`scripts/extract-release-notes.sh`** - Extracts version-specific section from CHANGELOG
- **`scripts/update-unreleased.sh`** - Manages Unreleased section transitions
- **`scripts/requirements.txt`** - Python dependencies

#### GitHub Actions
- **`.github/workflows/release.yml`** - Release process with changelog integration
- **`.github/workflows/branch-cut.yaml`** - Branch cut with CHANGELOG handling

#### Configuration
- **`.github/changelog-config.json`** - AI model settings and categorization rules
- **`CHANGELOG.md`** - Main changelog file (root level)
- **`.changelog/`** - Directory containing individual PR changelog fragments
  - `.changelog/README.md` - Documentation for changelog fragments
  - `.changelog/{PR_NUMBER}.txt` - Individual PR changelog entries

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

The system uses OpenAI GPT with specific rules to categorize changes:

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

### Environment Variables

Required for local execution:
- `OPENAI_API_KEY` - For AI-powered changelog generation (set in your environment or `.env` file)

Optional (for GitHub Actions workflows):
- `GITHUB_TOKEN` - Automatically provided by GitHub Actions
- `JENKINS_PAT` - For branch cut workflow (if using Jenkins integration)
- `RCTL_GO_MODULES_TOKEN` - For accessing private Go modules

### Customization

Edit `.github/changelog-config.json` to customize:
- AI model version (e.g., gpt-4-turbo-preview, gpt-4, gpt-3.5-turbo)
- Maximum commits per PR
- Category names
- Keyword weights
- Skip patterns

## Troubleshooting

### Changelog Not Updated

**Check:**
1. Was the changelog generation script run manually after PR creation?
2. Verify `OPENAI_API_KEY` is set in your environment or `.env` file
3. Check that the script completed successfully
4. Ensure you're running the script from the correct branch with the created PR

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
2. Run the deprecation scanner manually: `go build scripts/scan-deprecations.go && ./scan-deprecations -path ./rafay -verbose`
3. Verify Go file was actually changed in the PR
4. If using `--deprecations-file`, ensure the file path is correct

## Manual Operations

### Generating Changelog After PR Creation

**Note:** This is the standard process due to branch protection rules. After a PR is created, follow these steps:

```bash
# Install dependencies
pip install -r scripts/requirements.txt

# Set API key
export OPENAI_API_KEY="your-key-here"

# Generate changelog (dry run - preview only, no files written)
python3 scripts/generate-changelog.py \
  --pr-number 123 \
  --pr-url "https://github.com/RafaySystems/terraform-provider-rafay/pull/123" \
  --base-ref origin/master \
  --head-ref HEAD \
  --dry-run

# Actually generate and write both .changelog/123.txt and update CHANGELOG.md
python3 scripts/generate-changelog.py \
  --pr-number 123 \
  --pr-url "https://github.com/RafaySystems/terraform-provider-rafay/pull/123" \
  --base-ref origin/master \
  --head-ref HEAD \
  --target-section "Unreleased"

# Generate without PR number (only updates CHANGELOG.md, no fragment file)
python3 scripts/generate-changelog.py \
  --base-ref HEAD~5 \
  --head-ref HEAD \
  --target-section "Unreleased"
```

**Note:** When `--pr-number` is provided, the script will:
1. Create/update `.changelog/{PR_NUMBER}.txt` with the categorized entries
2. Update the `CHANGELOG.md` Unreleased section with the same entries

When `--pr-number` is NOT provided, only `CHANGELOG.md` is updated.

**After running the script:**
1. Review the generated entries in `.changelog/{PR_NUMBER}.txt` and `CHANGELOG.md`
2. Commit the changes: `git add .changelog/ CHANGELOG.md`

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
2. **Flexible Process**: Manual generation allows review before committing
3. **No Label Requirements**: Works with any commit style
4. **Deprecation Tracking**: Can scan for deprecation warnings
5. **Branch-Aware**: Handles master and release branches with proper targeting
6. **Professional Quality**: Follows Terraform AWS provider standards

## Maintenance

### Updating AI Model

Edit `.github/changelog-config.json`:
```json
{
  "ai_model": "gpt-4-turbo-preview"  // Or gpt-4, gpt-4o, etc.
}
```

### Adjusting Categorization

Edit the categorization rules in `.github/changelog-config.json` or adjust the prompt in `scripts/generate-changelog.py`.

### Updating Python Dependencies

```bash
# Update requirements.txt
pip install --upgrade openai requests python-dotenv

# Regenerate requirements.txt
pip freeze | grep -E 'openai|requests|python-dotenv' > scripts/requirements.txt
```

## Support

For issues or questions:
- Check [Testing Guide](./testing-guide.md) for validation steps
- Review [Commit Guidelines](./commit-guidelines.md) for best practices
- Open an issue in the repository with `changelog` label

---

**Last Updated**: January 2025  
**System Version**: 1.0

