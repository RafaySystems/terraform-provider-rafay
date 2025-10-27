# Changelog Automation Testing Guide

## Overview

This guide helps you test and validate the automated changelog system before and after deployment.

## Prerequisites

Before testing, ensure you have:

1. **Python 3.9+** installed
2. **Go 1.19+** installed (for deprecation scanner)
3. **OpenAI API Key** (set as `OPENAI_API_KEY` environment variable)
4. **Git repository** with full history
5. **GitHub repository secrets** configured (for CI/CD testing)

## Local Testing

### 1. Test Python Dependencies

```bash
# Install dependencies
cd /path/to/terraform-provider-rafay
pip install -r scripts/requirements.txt

# Verify installation
python3 -c "import openai; print('✓ Dependencies installed')"
```

### 2. Test Deprecation Scanner

```bash
# Build the scanner
cd scripts
go build -o scan-deprecations scan-deprecations.go

# Test on actual code
./scan-deprecations -path ../rafay -output test-deprecations.json -verbose

# Verify output
cat test-deprecations.json | python3 -m json.tool
```

**Expected Output:**
```json
{
  "deprecations": [
    {
      "resource": "rafay_aks_cluster",
      "field": "project_id",
      "message": "Deprecated in v1.5.0...",
      "file": "rafay/resource_aks_cluster.go",
      "line": 145
    }
  ]
}
```

### 3. Test Changelog Generator (Dry Run)

```bash
# Set your API key
export OPENAI_API_KEY="your-key-here"

# Test with recent commits
python3 scripts/generate-changelog.py \
  --pr-number 999 \
  --pr-url "https://github.com/RafaySystems/terraform-provider-rafay/pull/999" \
  --base-ref HEAD~5 \
  --head-ref HEAD \
  --dry-run
```

**Expected Output:**
```
Fetching commits from HEAD~5...HEAD
Found 5 commit(s)
Loaded 0 deprecation(s)
Generating changelog with AI...

=== Generated Changelog Entries ===

### FEATURES
* resource/rafay_eks_cluster: Add support for custom IAM roles ([#999](link))

### ENHANCEMENTS
* provider: Improve error handling for API timeouts ([#999](link))
```

### 4. Test Actual CHANGELOG Update

```bash
# Create a test branch
git checkout -b test-changelog-system

# Run generator to update CHANGELOG.md
python3 scripts/generate-changelog.py \
  --pr-number 999 \
  --pr-url "https://github.com/RafaySystems/terraform-provider-rafay/pull/999" \
  --base-ref HEAD~5 \
  --head-ref HEAD \
  --target-section "Unreleased"

# Verify changes
git diff CHANGELOG.md

# Clean up
git checkout master
git branch -D test-changelog-system
```

### 5. Test Helper Scripts

#### Test extract-release-notes.sh

```bash
# Extract notes for existing version
bash scripts/extract-release-notes.sh 1.1.51

# Verify output contains expected sections
bash scripts/extract-release-notes.sh 1.1.51 | grep -E "^### (FEATURES|BUG FIXES)"
```

#### Test update-unreleased.sh

```bash
# Create test CHANGELOG
cp CHANGELOG.md CHANGELOG.test.md

# Test rename
bash scripts/update-unreleased.sh rename 9.9.9 CHANGELOG.test.md
grep "## 9.9.9" CHANGELOG.test.md && echo "✓ Rename works"

# Test reset
bash scripts/update-unreleased.sh reset CHANGELOG.test.md
grep "## Unreleased" CHANGELOG.test.md && echo "✓ Reset works"

# Clean up
rm CHANGELOG.test.md
```

## GitHub Actions Testing

### Test PR Merge Workflow (Recommended: Use Test Repository First)

1. **Create a test PR** in your repository or fork

```bash
# Create feature branch
git checkout -b test-changelog-automation

# Make a trivial change
echo "# Test" >> test-file.md
git add test-file.md
git commit -m "feat: Test automated changelog generation

This is a test commit to verify the automated changelog system works correctly."

# Push and create PR
git push origin test-changelog-automation
```

2. **Merge the PR** through GitHub UI

3. **Verify the workflow**:
   - Go to Actions tab
   - Find "Update Changelog on PR Merge" workflow
   - Check the workflow ran successfully
   - Verify CHANGELOG.md was updated with a new commit

4. **Inspect the CHANGELOG**:
```bash
git pull origin master
git log --oneline CHANGELOG.md
git show HEAD:CHANGELOG.md | grep "test-changelog-automation"
```

### Test Branch Cut Workflow

1. **Manually trigger branch-cut workflow**:
   - Go to Actions → Release Branch Cut Workflow
   - Click "Run workflow"
   - Set source_branch: `master`
   - Set release_branch: `v9.9.9-test`
   - Click "Run workflow"

2. **Verify the results**:
```bash
# Check the new branch was created
git fetch --all
git checkout v9.9.9-test

# Verify CHANGELOG was updated
grep "## 9.9.9" CHANGELOG.md && echo "✓ Version section created"

# Clean up test branch
git push origin --delete v9.9.9-test
```

### Test Release Workflow

1. **Create a test tag** (on a test release branch):
```bash
git checkout v9.9.9-test
git tag -a v9.9.9-test -m "Test release"
git push origin v9.9.9-test
```

2. **Verify workflow execution**:
   - Go to Actions → release workflow
   - Check release notes extraction step
   - Verify GoReleaser used the extracted notes

3. **Clean up**:
```bash
git push origin --delete v9.9.9-test
git tag -d v9.9.9-test
git push origin :refs/tags/v9.9.9-test
```

## Validation Checklist

### Pre-Deployment

- [ ] Python dependencies install without errors
- [ ] Deprecation scanner compiles successfully
- [ ] Deprecation scanner detects known deprecations
- [ ] Changelog generator works in dry-run mode
- [ ] Changelog generator updates CHANGELOG.md correctly
- [ ] Helper scripts execute without errors
- [ ] OPENAI_API_KEY is set in GitHub secrets

### Post-Deployment

- [ ] PR merge triggers changelog workflow
- [ ] Workflow completes successfully
- [ ] CHANGELOG.md is updated with new entry
- [ ] Entry is in correct section (Unreleased for master)
- [ ] Entry includes PR number and link
- [ ] Deprecations are detected and included
- [ ] No duplicate entries appear
- [ ] Branch cut updates CHANGELOG correctly
- [ ] Release notes are extracted correctly

## Common Issues and Solutions

### Issue: "OPENAI_API_KEY not set"

**Solution:**
```bash
# For local testing
export OPENAI_API_KEY="sk-..."

# For GitHub Actions
# Add to repository secrets: Settings → Secrets → Actions → New secret
```

### Issue: "No commits found"

**Solution:**
- Check git ref syntax: `origin/master` not just `master`
- Ensure full git history: `git fetch --unshallow`
- Verify base and head refs are valid

### Issue: "Deprecation scanner finds nothing"

**Solution:**
1. Check Go code has proper `Deprecated` or `DeprecationMessage` fields
2. Verify files are in the scanned path
3. Run with `-verbose` flag for debugging:
```bash
./scripts/scan-deprecations -path ./rafay -verbose
```

### Issue: "AI generates poor categorization"

**Solution:**
1. Check commit messages - be more descriptive
2. Adjust prompt in `scripts/generate-changelog.py`
3. Modify categorization rules in `.github/changelog-config.json`
4. Consider using conventional commit format

### Issue: "Workflow fails with permission error"

**Solution:**
```yaml
# Ensure workflow has write permissions
permissions:
  contents: write
  pull-requests: read
```

### Issue: "Cherry-picked PR appears twice"

**Solution:**
- The system should detect duplicates by PR number
- Check if PR reference is included in entries
- Verify duplicate detection logic in `categorize_entries` function

## Performance Testing

### Test with Large Number of Commits

```bash
# Test with 100 commits
python3 scripts/generate-changelog.py \
  --pr-number 999 \
  --pr-url "https://github.com/RafaySystems/terraform-provider-rafay/pull/999" \
  --base-ref HEAD~100 \
  --head-ref HEAD \
  --dry-run
```

**Expected**: Should complete within 30-60 seconds

### Test AI Model Response Time

```bash
# Time the changelog generation
time python3 scripts/generate-changelog.py \
  --pr-number 999 \
  --pr-url "..." \
  --base-ref HEAD~10 \
  --head-ref HEAD \
  --dry-run
```

**Expected**: 5-15 seconds for typical PRs

## Integration Testing Scenarios

### Scenario 1: Simple Bug Fix PR

```bash
# Commit
git commit -m "fix: Correct timeout handling in EKS operations"

# Expected CHANGELOG entry
### BUG FIXES
* resource/rafay_eks_cluster: Correct timeout handling ([#XXX](link))
```

### Scenario 2: New Feature PR

```bash
# Commits
git commit -m "feat: Add IPv6 support to EKS clusters"
git commit -m "docs: Add IPv6 configuration example"

# Expected CHANGELOG entry
### FEATURES
* resource/rafay_eks_cluster: Add IPv6 networking support ([#XXX](link))
```

### Scenario 3: Breaking Change PR

```bash
# Commit
git commit -m "breaking: Remove deprecated rafay_cluster resource

BREAKING CHANGE: The rafay_cluster resource has been removed.
Use cloud-specific resources instead."

# Expected CHANGELOG entry
### BREAKING CHANGES
* Remove deprecated rafay_cluster resource. Use cloud-specific resources (rafay_eks_cluster, rafay_aks_cluster, rafay_gke_cluster) instead ([#XXX](link))
```

### Scenario 4: Deprecation PR

```bash
# Commit with code changes adding Deprecated field
git commit -m "deprecate: Mark project_id as deprecated

Adds deprecation warning for project_id argument.
Users should migrate to metadata.project."

# Expected CHANGELOG entry
### DEPRECATIONS
* resource/rafay_aks_cluster: Deprecate `project_id` argument in favor of `metadata.project` ([#XXX](link))
```

## Monitoring and Maintenance

### Regular Checks

**Weekly:**
- [ ] Review recent CHANGELOG entries for quality
- [ ] Check GitHub Actions success rate
- [ ] Verify no duplicate entries

**Monthly:**
- [ ] Review AI categorization accuracy
- [ ] Check for missed deprecations
- [ ] Validate changelog follows style guide

**Quarterly:**
- [ ] Update Python dependencies
- [ ] Review and update categorization rules
- [ ] Consider AI model upgrades

### Metrics to Track

- Workflow success rate (target: >95%)
- AI categorization accuracy (manual review sample)
- Time to generate changelog (target: <30s)
- User feedback on changelog quality

## Rollback Plan

If the automated system has issues:

1. **Disable the workflow**:
```bash
# Rename workflow file temporarily
mv .github/workflows/changelog-on-merge.yml \
   .github/workflows/changelog-on-merge.yml.disabled
```

2. **Manual changelog updates**:
```bash
# Edit CHANGELOG.md manually
# Follow the same format and categories
# Commit with [skip ci] to avoid triggering workflows
```

3. **Fix and re-enable**:
- Debug the issue locally
- Test thoroughly
- Re-enable workflow

## Support and Questions

- **Documentation**: Read [automated-system.md](./automated-system.md)
- **Commit Guidelines**: See [commit-guidelines.md](./commit-guidelines.md)
- **Issues**: Open GitHub issue with `changelog` label
- **Testing**: Use test repository or feature branch first

---

**Remember**: Test thoroughly in a non-production environment before deploying to the main repository!

