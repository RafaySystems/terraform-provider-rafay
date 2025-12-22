# Changelog Guidelines

## Overview

The Rafay Terraform Provider uses an **AI-powered changelog generation system** that generates and maintains the `CHANGELOG.md` file. This document provides a brief overview and links to detailed documentation.

**Note:** Due to branch protection rules, changelog generation is a manual process that must be run after a PR is created.

## How It Works

After you create a PR:
1. **Manual Script Execution**: Run the changelog generation script (see [Manual Generation](#manual-generation))
2. **AI Analysis**: OpenAI GPT analyzes your commits and code changes
3. **Smart Categorization**: Changes are categorized into BREAKING CHANGES, FEATURES, ENHANCEMENTS, BUG FIXES, DEPRECATIONS, and DOCUMENTATION
4. **Deprecation Detection**: Go code can be scanned for `Deprecated` and `DeprecationMessage` fields (optional)
5. **CHANGELOG Update**: Entries are written to `.changelog/{PR_NUMBER}.txt` and `CHANGELOG.md`
6. **Review and Commit**: Review the generated entries and commit them

## Branch Strategy

### Master Branch
- After PR merges to `master`, manually generate changelog entries for the "Unreleased" section
- Entries accumulate until a release branch is cut

### Release Branches  
- When a release branch is cut (e.g., `v1.2.0`), the "Unreleased" section becomes the version section
- After cherry-picking PRs to release branches, manually generate changelog entries for that version's section
- Use `--target-section` to specify the version number (e.g., `1.2.0`)
- Duplicate detection prevents the same PR from appearing twice

### Cherry-Picking Flow
1. PR merges to master → Manually generate changelog for "Unreleased" section
2. PR cherry-picked to release branch → Manually generate changelog for version section (e.g., "1.2.0")
3. System detects duplicate by PR number and skips if already present

## What You Need to Do

### ✅ Do This
- Write clear, descriptive commit messages
- Mention breaking changes explicitly in commit or PR description
- Use conventional commit format if possible (feat:, fix:, etc.)
- After PR creation, run the changelog generation script manually
- Review generated entries before committing
- Commit changelog changes

### ❌ Don't Do This
- Don't forget to generate changelog after PR creation
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

- `.github/changelog-config.json` - AI and categorization configuration
- `scripts/generate-changelog.py` - AI-powered changelog generator
- `scripts/scan-deprecations.go` - Go code deprecation scanner
- `.changelog/README.md` - Manual generation instructions

## Manual Generation

After a PR is merged, generate the changelog manually:

```bash
# Preview (dry run)
python3 scripts/generate-changelog.py \
  --pr-number 1131 \
  --pr-url https://github.com/RafaySystems/terraform-provider-rafay/pull/1131 \
  --base-ref origin/master \
  --head-ref HEAD \
  --dry-run

# Generate and write files
python3 scripts/generate-changelog.py \
  --pr-number 1131 \
  --pr-url https://github.com/RafaySystems/terraform-provider-rafay/pull/1131 \
  --base-ref origin/master \
  --head-ref HEAD
```

See [`.changelog/README.md`](../../.changelog/README.md) for detailed instructions.

## Troubleshooting

### Changelog Not Updated?

1. Verify you ran the changelog generation script
2. Check that `OPENAI_API_KEY` is set in your environment
3. Ensure the script completed successfully
4. Review the generated files before committing

### Incorrect Categorization?

1. Use clearer commit messages
2. Add explicit category hints in PR description
3. System learns from patterns - provide feedback

### Missing Deprecation?

1. Verify `Deprecated` or `DeprecationMessage` is in Go code
2. Run the deprecation scanner manually: `go build scripts/scan-deprecations.go && ./scan-deprecations -path ./rafay -verbose`
3. Use `--deprecations-file` when running the changelog generator
4. Ensure the Go file was actually changed in the PR

## Benefits

✅ **Consistent Quality** - AI ensures professional style  
✅ **Flexible Process** - Manual generation allows review before committing  
✅ **Deprecation Detection** - Can scan for deprecation warnings  
✅ **Branch-Aware** - Works across master and release branches  
✅ **Cherry-Pick Friendly** - Handles your existing workflow  

## Questions?

- 📚 Read the [detailed documentation](./automated-system.md)
- 🧪 Check the [testing guide](./testing-guide.md) to validate locally

---

**Last Updated**: December 2025  
**Version**: 1.0
