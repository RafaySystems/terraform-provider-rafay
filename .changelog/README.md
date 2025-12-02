# Changelog Fragments

This directory contains individual changelog fragments for each merged pull request. These fragments are manually generated using the changelog generation script and later compiled into the main `CHANGELOG.md` during releases.

## Workflow

### 1. PR Merge Process

**Note:** Due to branch protection rules, changelog generation is a manual process. After a PR is merged, follow these steps:

1. **Run the changelog generator script** (see [Manual Generation](#manual-generation) below)
2. The script will:
   - **Analyze commits** in the PR
   - **Generate changelog entries** using AI (categorized by type)
   - **Write a fragment file** to `.changelog/{PR_NUMBER}.txt`
   - **Update** the `Unreleased` section in `CHANGELOG.md`
3. **Review the generated entries** for accuracy


### 2. File Naming Convention

```
.changelog/
├── 1131.txt        # Changelog for PR #1131
├── 1132.txt        # Changelog for PR #1132
├── 1145.txt        # Changelog for PR #1145
└── README.md       # This file
```

Each file is named with the PR number: `{PR_NUMBER}.txt`

### 3. Fragment Format

Each fragment file contains categorized changelog entries:

```markdown
### FEATURES

* **New Resource:** `rafay_eks_cluster`: Support for managing EKS clusters ([#1131](https://github.com/RafaySystems/terraform-provider-rafay/pull/1131))

### ENHANCEMENTS

* resource/rafay_gke_cluster: Add label support for better resource organization ([#1131](https://github.com/RafaySystems/terraform-provider-rafay/pull/1131))

### BUG FIXES

* resource/rafay_aks_cluster: Fix NPE when node pool configuration is nil ([#1131](https://github.com/RafaySystems/terraform-provider-rafay/pull/1131))
```

## Categories

Changelog entries are categorized into:

- **BREAKING CHANGES** - Changes that break existing user configurations
- **FEATURES** - New resources, data sources, or major functionality
- **ENHANCEMENTS** - Improvements to existing functionality
- **BUG FIXES** - Corrections to incorrect behavior
- **DEPRECATIONS** - Advance notice of future breaking changes
- **DOCUMENTATION** - Significant documentation updates

## Manual Generation

After a PR is merged, generate the changelog entries manually using the script below.

### Prerequisites

- Python 3.9+ installed
- Dependencies installed: `pip install -r scripts/requirements.txt`
- `OPENAI_API_KEY` environment variable set (or in `.env` file)
- Go installed (for deprecation scanning)
- Repository checked out with the merged PR

### Steps

1. **Preview the changelog (dry run):**

```bash
python3 scripts/generate-changelog.py \
  --pr-number 1131 \
  --pr-url https://github.com/RafaySystems/terraform-provider-rafay/pull/1131 \
  --base-ref origin/master \
  --head-ref HEAD \
  --dry-run
```

2. **Generate and write the changelog files:**

```bash
python3 scripts/generate-changelog.py \
  --pr-number 1131 \
  --pr-url https://github.com/RafaySystems/terraform-provider-rafay/pull/1131 \
  --base-ref origin/master \
  --head-ref HEAD
```

This will:
1. Create/update `.changelog/1131.txt` (PR-specific fragment)
2. Update the `Unreleased` section in `CHANGELOG.md`

3. **Review the generated entries** for accuracy and completeness



### Command Line Arguments

- `--pr-number`: The pull request number (required)
- `--pr-url`: The full URL to the pull request (required)
- `--base-ref`: Base git reference (default: `origin/master`)
- `--head-ref`: Head git reference (default: `HEAD`)
- `--target-section`: Target changelog section (default: `Unreleased`)
- `--deprecations-file`: Path to deprecations JSON file (optional, for deprecation scanning)
- `--changelog-path`: Path to CHANGELOG.md (default: `CHANGELOG.md`)
- `--dry-run`: Preview output without writing files

## Release Process

During a release, fragments are compiled into the versioned section of `CHANGELOG.md`:

1. All fragments in `.changelog/` are reviewed
2. Entries are consolidated into a version section (e.g., `## 1.2.0 (Nov 15, 2025)`)
3. The `Unreleased` section is cleared

## Notes

- **Manual process required:** Changelog generation must be done manually after PR merge due to branch protection rules
- Fragment files should be committed to git
- Each PR should have its own fragment file
- Fragments are used for audit trail and release compilation
- AI-generated content should be reviewed for accuracy before committing
- The script requires `OPENAI_API_KEY` to be set in your environment or `.env` file

