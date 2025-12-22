# Changelog Scripts

This directory contains scripts for generating and managing changelog entries for the Rafay Terraform Provider.

## Scripts Overview

### 1. `generate-changelog.py` (Primary Script)

**Purpose:** Generate AI-powered changelog entries from git commits

**Important:** This script outputs to CLI/stdout ONLY and does NOT write to any files.

**Usage:**
```bash
python3 scripts/generate-changelog.py \
  --pr-number 1131 \
  --pr-url https://github.com/RafaySystems/terraform-provider-rafay/pull/1131 \
  --base-ref origin/master \
  --head-ref HEAD
```

**Arguments:**
- `--pr-number` - Pull request number (optional, for reference in output)
- `--pr-url` - Pull request URL (optional, for reference in output)  
- `--base-ref` - Base git reference (default: `origin/master`)
- `--head-ref` - Head git reference (default: `HEAD`)
- `--deprecations-file` - Path to deprecations JSON file (optional)
- `--config` - Config file path (default: `.github/changelog-config.json`)

**Prerequisites:**
- Python 3.9+
- Install dependencies: `pip install -r scripts/requirements.txt`
- Set environment variable: `OPENAI_API_KEY`

**Output Format:**
```
================================================================================
CHANGELOG GENERATOR - CLI MODE (NO FILES MODIFIED)
================================================================================
PR Number: 1131
PR URL: https://github.com/.../pull/1131
Base Ref: origin/master
Head Ref: HEAD
================================================================================

Fetching commits from origin/master...HEAD
Found 5 commit(s)

Generating changelog with AI...

================================================================================
GENERATED CHANGELOG ENTRIES
================================================================================

### FEATURES

* **New Resource:** `rafay_eks_cluster`: Support for managing EKS clusters ([#1131](...))

### ENHANCEMENTS

* resource/rafay_gke_cluster: Add label support ([#1131](...))

================================================================================
END OF CHANGELOG ENTRIES
================================================================================

Note: No files were modified. Copy the above entries as needed.
```

**Workflow:**
1. Run the script to generate changelog entries
2. Manually copy the output
3. Update `CHANGELOG.md` Unreleased section with the entries
4. Commit the changes to git

---

### 2. `scan-deprecations.go`

**Purpose:** Scan codebase for deprecated fields and resources

**Usage:**
```bash
go run scripts/scan-deprecations.go > scripts/deprecations.json
```

**Output:** JSON file containing detected deprecations that can be passed to `generate-changelog.py`

---

## Typical Workflow

### After Merging a PR

1. **Generate changelog entries:**
   ```bash
   python3 scripts/generate-changelog.py \
     --pr-number 1234 \
     --pr-url https://github.com/RafaySystems/terraform-provider-rafay/pull/1234 \
     --base-ref origin/master \
     --head-ref HEAD
   ```

2. **Copy the CLI output** (everything between the separator lines)

3. **Update CHANGELOG.md:**
   ```bash
   # Edit CHANGELOG.md and paste entries under ## Unreleased section
   vim CHANGELOG.md
   ```

4. **Commit the changes:**
   ```bash
   git add CHANGELOG.md
   git commit -m "docs: Add changelog for PR #1234"
   git push
   ```

---

## Configuration

### `.github/changelog-config.json`

Configuration file for the changelog generator:

- `ai_model` - OpenAI model to use (default: `gpt-4o-mini`)
- `max_commits_per_pr` - Maximum commits to analyze (default: 300)
- `categories` - Changelog categories (BREAKING CHANGES, FEATURES, etc.)
- `*_keywords` - Keywords for categorizing changes

---

## Environment Variables

- `OPENAI_API_KEY` - Required for `generate-changelog.py`
  - Set in environment or create `.env` file in project root
  - Get API key from: https://platform.openai.com/api-keys

---

## File Structure

```
scripts/
├── generate-changelog.py          # Main changelog generator (CLI output only)
├── scan-deprecations.go           # Scan for deprecations
├── requirements.txt               # Python dependencies
└── README.md                      # This file

.github/
└── changelog-config.json          # Changelog generator configuration

CHANGELOG.md                       # Main changelog file (manually updated)
```

---

## Important Notes

1. **No Automatic File Writes**: The main changelog generator (`generate-changelog.py`) outputs to CLI only and does NOT modify any files automatically. This is by design to ensure full control over what gets committed.

2. **Manual Process**: You must manually copy the generated output and update CHANGELOG.md.

3. **Review Required**: Always review AI-generated content for accuracy before committing.

4. **API Key Security**: Never commit your `OPENAI_API_KEY` to the repository. Use environment variables or `.env` file (which is gitignored).

---

## Troubleshooting

### "OPENAI_API_KEY not set"
Set the environment variable:
```bash
export OPENAI_API_KEY='your-api-key-here'
```

Or create a `.env` file in the project root:
```
OPENAI_API_KEY=your-api-key-here
```

### "No commits found"
Check that your git references are correct:
```bash
git log origin/master..HEAD  # Should show commits
```

### Python dependencies missing
Install requirements:
```bash
pip install -r scripts/requirements.txt
```

---

## See Also

- [Changelog Guidelines](../docs/changelog/README.md)
- [Commit Guidelines](../docs/changelog/commit-guidelines.md)
