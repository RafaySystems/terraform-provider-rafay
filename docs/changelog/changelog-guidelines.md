# Changelog Guidelines

## Overview

The Rafay Terraform Provider uses an **automated AI-powered changelog system** that generates and maintains the `CHANGELOG.md` file. This document provides a brief overview and links to detailed documentation.

## How It Works

When you merge a PR:
1. **Automatic Trigger**: GitHub Actions workflow activates
2. **AI Analysis**: OpenAI GPT analyzes your commits and code changes
3. **Smart Categorization**: Changes are categorized into BREAKING CHANGES, FEATURES, ENHANCEMENTS, BUG FIXES, DEPRECATIONS, and DOCUMENTATION
4. **Deprecation Detection**: Go code is scanned for `Deprecated` and `DeprecationMessage` fields
5. **CHANGELOG Update**: Entries are automatically added to CHANGELOG.md
6. **Branch-Aware**: Works correctly on both master (Unreleased) and release branches

## Branch Strategy

### Master Branch
- All PR merges to `master` add entries to the "Unreleased" section
- Entries accumulate until a release branch is cut

### Release Branches  
- When a release branch is cut (e.g., `v1.2.0`), the "Unreleased" section becomes the version section
- Cherry-picked PRs to release branches are added to that version's section
- Duplicate detection prevents the same PR from appearing twice

### Cherry-Picking Flow
1. PR merges to master → Added to "Unreleased"
2. PR cherry-picked to release branch → Added to version section (e.g., "1.2.0")
3. System detects duplicate by PR number and skips if already present

## What You Need to Do

### ✅ Do This
- Write clear, descriptive commit messages
- Mention breaking changes explicitly in commit or PR description
- Use conventional commit format if possible (feat:, fix:, etc.)
- Add `skip-changelog` label for internal changes (tests, refactoring)

### ❌ Don't Do This
- Don't manually edit CHANGELOG.md (it won't be overwritten, but automation is better)
- Don't worry about categorization (AI handles it)
- Don't stress about perfect wording (AI makes it user-friendly)

## Quick Examples

**Good Commit Messages:**
```
feat: Add IPv6 support to EKS clusters

fix: Correct timeout handling in cluster operations

deprecate: Mark project_id as deprecated, use metadata.project instead

breaking: Remove deprecated rafay_cluster resource
```

**Result in CHANGELOG:**
```markdown
## Unreleased

### FEATURES
* resource/rafay_eks_cluster: Add IPv6 networking support ([#445](link))

### BUG FIXES
* resource/rafay_eks_cluster: Correct timeout handling ([#447](link))

### DEPRECATIONS
* resource/rafay_aks_cluster: Deprecate `project_id` argument in favor of `metadata.project` ([#449](link))
```

## Documentation Links

### 📖 Detailed Documentation

- **[Automated System Overview](./automated-system.md)** - Complete technical documentation of the automation system
- **[Commit Guidelines](./commit-guidelines.md)** - Best practices for writing commits
- **[Testing Guide](./testing-guide.md)** - How to test and validate the system

### 🔧 Configuration Files

- `.github/workflows/changelog-on-merge.yml` - Main automation workflow
- `.github/changelog-config.json` - AI and categorization configuration
- `scripts/generate-changelog.py` - AI-powered changelog generator
- `scripts/scan-deprecations.go` - Go code deprecation scanner

## Troubleshooting

### Changelog Not Updated?

1. Check GitHub Actions tab for workflow status
2. Verify PR was merged (not just closed)
3. Check if `skip-changelog` label was added

### Incorrect Categorization?

1. Use clearer commit messages
2. Add explicit category hints in PR description
3. System learns from patterns - provide feedback

### Missing Deprecation?

1. Verify `Deprecated` or `DeprecationMessage` is in Go code
2. Check the deprecation scanner output in Actions logs
3. Ensure the Go file was actually changed in the PR

## Benefits

✅ **No Manual Work** - Automatic on every PR merge  
✅ **Consistent Quality** - AI ensures professional style  
✅ **Never Miss Deprecations** - Automatic code scanning  
✅ **Branch Sync** - Works across master and release branches  
✅ **Cherry-Pick Friendly** - Handles your existing workflow  

## Questions?

- 📚 Read the [detailed documentation](./automated-system.md)
- 🧪 Check the [testing guide](./testing-guide.md) to validate locally

---

**System Status**: ✅ Active  
**Last Updated**: October 2025  
**Version**: 1.0
