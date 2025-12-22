# Changelog Generator Testing Guide

## Overview

This guide helps you test and validate the changelog generator script. The script outputs to CLI/stdout only and does not write any files.

## Prerequisites

Before testing, ensure you have:

1. **Python 3.9+** installed
2. **Go 1.19+** installed (optional, for deprecation scanner)
3. **OpenAI API Key** (set as `OPENAI_API_KEY` environment variable)
4. **Git repository** with full history
5. **Python dependencies** installed

## Setup

### 1. Install Python Dependencies

```bash
# Navigate to project root
cd /path/to/terraform-provider-rafay

# Install dependencies
pip install -r scripts/requirements.txt

# Verify installation
python3 -c "import openai; print('✓ Dependencies installed')"
```

### 2. Set OpenAI API Key

```bash
# Option 1: Environment variable
export OPENAI_API_KEY="sk-your-api-key-here"

# Option 2: Create .env file in project root
echo "OPENAI_API_KEY=sk-your-api-key-here" > .env

# Verify it's set
python3 -c "import os; print('✓ API key set' if os.getenv('OPENAI_API_KEY') else '✗ API key not set')"
```

## Testing the Changelog Generator

### Test 1: Help Output

Verify the script runs and shows usage information:

```bash
python3 scripts/generate-changelog.py --help
```

**Expected Output:**
```
usage: generate-changelog.py [-h] [--pr-number PR_NUMBER] [--pr-url PR_URL]
                             [--base-ref BASE_REF] [--head-ref HEAD_REF]
                             [--deprecations-file DEPRECATIONS_FILE]
                             [--config CONFIG]

Generate changelog entries using AI (CLI output only - no files written)
...
```

### Test 2: Generate Changelog for Recent Commits

Test with recent commits in your current branch:

```bash
python3 scripts/generate-changelog.py \
  --base-ref origin/master \
  --head-ref HEAD
```

**Expected Output:**
```
================================================================================
CHANGELOG GENERATOR - CLI MODE (NO FILES MODIFIED)
================================================================================
PR Number: N/A
PR URL: N/A
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

* **New Resource:** `rafay_eks_cluster`: Support for Amazon EKS clusters

### ENHANCEMENTS

* provider: Improve error handling for API timeouts

================================================================================
END OF CHANGELOG ENTRIES
================================================================================

Note: No files were modified. Copy the above entries as needed.
```

### Test 3: Generate with PR Reference

Test with PR number and URL (for reference in output):

```bash
python3 scripts/generate-changelog.py \
  --pr-number 1234 \
  --pr-url "https://github.com/RafaySystems/terraform-provider-rafay/pull/1234" \
  --base-ref origin/master \
  --head-ref HEAD
```

**Expected:**
- Output includes PR number and URL in the header
- PR references are included in generated entries: `([#1234](url))`

### Test 4: Generate with Specific Commit Range

Test with a specific range of commits:

```bash
# Test with last 5 commits
python3 scripts/generate-changelog.py \
  --base-ref HEAD~5 \
  --head-ref HEAD

# Test with specific commit hashes
python3 scripts/generate-changelog.py \
  --base-ref abc1234 \
  --head-ref def5678
```

### Test 5: Generate with Deprecations

First, generate a deprecations file (optional):

```bash
# Build and run deprecation scanner
cd scripts
go build -o scan-deprecations scan-deprecations.go
./scan-deprecations -path ../rafay -output deprecations.json -verbose
cd ..
```

Then generate changelog with deprecations:

```bash
python3 scripts/generate-changelog.py \
  --base-ref origin/master \
  --head-ref HEAD \
  --deprecations-file scripts/deprecations.json
```

**Expected:**
- DEPRECATIONS section includes detected deprecations from code

## Testing the Deprecation Scanner (Optional)

### Test 1: Build the Scanner

```bash
cd scripts
go build -o scan-deprecations scan-deprecations.go
echo "✓ Scanner built successfully"
```

### Test 2: Scan for Deprecations

```bash
# Scan with verbose output
./scan-deprecations -path ../rafay -output test-deprecations.json -verbose

# View results
cat test-deprecations.json | python3 -m json.tool
```

**Expected Output:**
```json
{
  "deprecations": [
    {
      "resource": "rafay_aks_cluster",
      "field": "project_id",
      "message": "Deprecated: Use metadata.project instead",
      "file": "rafay/resource_aks_cluster.go",
      "line": 145
    }
  ]
}
```

### Test 3: Use Deprecations in Changelog

```bash
# Generate changelog with deprecations
cd ..
python3 scripts/generate-changelog.py \
  --base-ref origin/master \
  --head-ref HEAD \
  --deprecations-file scripts/test-deprecations.json
```

## Manual Workflow Testing

Test the complete manual workflow from generation to commit:

### Step 1: Create Test Branch

```bash
# Create test branch
git checkout -b test-changelog-cli
```

### Step 2: Generate Changelog Entries

```bash
# Generate entries
python3 scripts/generate-changelog.py \
  --pr-number 9999 \
  --pr-url "https://github.com/RafaySystems/terraform-provider-rafay/pull/9999" \
  --base-ref origin/master \
  --head-ref HEAD
```

### Step 3: Manually Copy Output

1. Copy the generated entries from the terminal (between the separator lines)
2. Open CHANGELOG.md in your editor
3. Paste the entries under the `## Unreleased` section

### Step 4: Verify Changes

```bash
# Check what changed
git diff CHANGELOG.md

# View the updated section
git diff CHANGELOG.md | grep "^+\*"
```

### Step 5: Clean Up

```bash
# Discard test changes
git checkout CHANGELOG.md

# Return to master
git checkout master
git branch -D test-changelog-cli
```

## Validation Tests

### Test Output Format

Verify the CLI output has the correct format:

```bash
python3 scripts/generate-changelog.py --base-ref HEAD~3 --head-ref HEAD > /tmp/changelog-output.txt

# Check for required sections
grep -q "CHANGELOG GENERATOR - CLI MODE" /tmp/changelog-output.txt && echo "✓ Header present"
grep -q "GENERATED CHANGELOG ENTRIES" /tmp/changelog-output.txt && echo "✓ Content section present"
grep -q "Note: No files were modified" /tmp/changelog-output.txt && echo "✓ Warning present"

# Verify no files were created
test ! -f .changelog/9999.txt && echo "✓ No fragment files created"
```

### Test No File Writes

Verify the script doesn't modify any files:

```bash
# Get current git status
git status --porcelain > /tmp/before.txt

# Run generator
python3 scripts/generate-changelog.py \
  --base-ref HEAD~5 \
  --head-ref HEAD \
  --pr-number 9999

# Check git status again
git status --porcelain > /tmp/after.txt

# Compare (should be identical)
diff /tmp/before.txt /tmp/after.txt && echo "✓ No files modified"
```

### Test with Different Commit Types

Create test commits with different prefixes:

```bash
# Create test branch
git checkout -b test-commit-types

# Test different commit types
echo "test1" > test1.md && git add test1.md && git commit -m "feat: Add new feature"
echo "test2" > test2.md && git add test2.md && git commit -m "fix: Fix critical bug"
echo "test3" > test3.md && git add test3.md && git commit -m "docs: Update documentation"
echo "test4" > test4.md && git add test4.md && git commit -m "deprecate: Mark old field as deprecated"

# Generate changelog
python3 scripts/generate-changelog.py \
  --base-ref master \
  --head-ref HEAD

# Clean up
git checkout master
git branch -D test-commit-types
```

**Expected:**
- `feat:` commits → FEATURES section
- `fix:` commits → BUG FIXES section
- `docs:` commits → DOCUMENTATION section
- `deprecate:` commits → DEPRECATIONS section

## Performance Testing

### Test with Large Commit Range

```bash
# Test with 50 commits
time python3 scripts/generate-changelog.py \
  --base-ref HEAD~50 \
  --head-ref HEAD

# Test with 100 commits
time python3 scripts/generate-changelog.py \
  --base-ref HEAD~100 \
  --head-ref HEAD
```

**Expected:**
- 50 commits: ~10-20 seconds
- 100 commits: ~15-30 seconds
- No errors or timeouts

### Test AI Response Time

```bash
# Time a typical changelog generation
time python3 scripts/generate-changelog.py \
  --base-ref HEAD~10 \
  --head-ref HEAD
```

**Expected:** 5-15 seconds for 10 commits

## Error Handling Tests

### Test Missing API Key

```bash
# Unset API key temporarily
unset OPENAI_API_KEY

# Try to run generator
python3 scripts/generate-changelog.py --base-ref HEAD~1 --head-ref HEAD

# Should see error message
# Expected: "Error: OPENAI_API_KEY environment variable not set"

# Restore API key
export OPENAI_API_KEY="your-key"
```

### Test Invalid Git References

```bash
# Test with non-existent ref
python3 scripts/generate-changelog.py \
  --base-ref invalid-ref-12345 \
  --head-ref HEAD

# Expected: Error message about invalid git reference
```

### Test No Commits Found

```bash
# Test with same base and head
python3 scripts/generate-changelog.py \
  --base-ref HEAD \
  --head-ref HEAD

# Expected: "No commits found."
```

## Integration Test Scenarios

### Scenario 1: Simple Bug Fix

```bash
# Create test commit
git checkout -b test-bugfix
echo "fix" > test.md
git add test.md
git commit -m "fix: Correct timeout handling in EKS operations

Fixed an issue where EKS cluster operations would timeout prematurely."

# Generate changelog
python3 scripts/generate-changelog.py \
  --base-ref master \
  --head-ref HEAD

# Expected in BUG FIXES:
# * resource/rafay_eks_cluster: Correct timeout handling
```

### Scenario 2: New Feature

```bash
# Create test commit
git checkout -b test-feature
echo "feat" > test.md
git add test.md
git commit -m "feat: Add IPv6 support to EKS clusters

Adds configuration options for IPv6 networking in EKS clusters."

# Generate changelog
python3 scripts/generate-changelog.py \
  --base-ref master \
  --head-ref HEAD

# Expected in FEATURES:
# * resource/rafay_eks_cluster: Add IPv6 networking support
```

### Scenario 3: Breaking Change

```bash
# Create test commit
git checkout -b test-breaking
echo "break" > test.md
git add test.md
git commit -m "breaking: Remove deprecated rafay_cluster resource

BREAKING CHANGE: The rafay_cluster resource has been removed.
Use cloud-specific resources instead."

# Generate changelog
python3 scripts/generate-changelog.py \
  --base-ref master \
  --head-ref HEAD

# Expected in BREAKING CHANGES:
# * Remove deprecated rafay_cluster resource
```

## Validation Checklist

Before using the changelog generator in production:

- [ ] Python dependencies install successfully
- [ ] OPENAI_API_KEY is set and valid
- [ ] Script runs without errors
- [ ] Output format is correct (headers, separators, categories)
- [ ] No files are created or modified
- [ ] PR references appear in output when provided
- [ ] Different commit types are categorized correctly
- [ ] Deprecation scanner works (if using)
- [ ] Performance is acceptable for typical commit ranges
- [ ] Error messages are clear and helpful

## Common Issues and Solutions

### Issue: "OPENAI_API_KEY not set"

**Solution:**
```bash
# Set in current session
export OPENAI_API_KEY="sk-..."

# Or create .env file
echo "OPENAI_API_KEY=sk-..." > .env
```

### Issue: "No commits found"

**Solution:**
- Check git ref syntax: use `origin/master` not just `master`
- Ensure refs are valid: `git log origin/master..HEAD`
- Verify you have commits in the range

### Issue: "Module 'openai' not found"

**Solution:**
```bash
# Reinstall dependencies
pip install -r scripts/requirements.txt

# Or install directly
pip install openai python-dotenv
```

### Issue: AI generates poor categorization

**Solution:**
1. Use clearer commit messages with prefixes (`feat:`, `fix:`, etc.)
2. Adjust configuration in `.github/changelog-config.json`
3. Review and manually edit the output before pasting to CHANGELOG.md

### Issue: "API rate limit exceeded"

**Solution:**
- Wait a few minutes before retrying
- Check your OpenAI API usage at platform.openai.com
- Consider upgrading your OpenAI API plan

## Best Practices

1. **Test on a branch** - Always test on a non-master branch first
2. **Review output** - Always review AI-generated content before committing
3. **Use meaningful commits** - Better commit messages = better changelog entries
4. **Regular testing** - Test the script regularly to ensure it still works
5. **Keep dependencies updated** - Update Python packages periodically

## Troubleshooting Commands

```bash
# Check Python version
python3 --version  # Should be 3.9+

# Check if dependencies are installed
pip list | grep openai
pip list | grep python-dotenv

# Test OpenAI API connection
python3 -c "from openai import OpenAI; client = OpenAI(); print('✓ API connection works')"

# Check git repository status
git status
git log --oneline -10

# View recent commits in range
git log origin/master..HEAD --oneline
```

## Support

- **Script Documentation**: See `scripts/README.md`
- **Commit Guidelines**: See `docs/changelog/commit-guidelines.md`
- **Issues**: Open GitHub issue with `changelog` label

---

**Remember**: The script only outputs to CLI. You must manually copy and paste the entries into CHANGELOG.md.
