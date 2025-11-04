# Commit Message Guidelines

## Overview

While our AI-powered changelog system can work with any commit style, following these guidelines helps ensure accurate categorization and professional changelog entries.

## General Principles

1. **Write for users, not developers**: Describe what changed from a user perspective
2. **Be specific**: "Add IPv6 support to EKS clusters" not "Add feature"
3. **Use present tense**: "Add" not "Added", "Fix" not "Fixed"
4. **Keep it concise**: First line under 72 characters
5. **Add context in body**: Use commit body for detailed explanations

## Commit Message Format

### Recommended Format

```
<type>: <short description>

<detailed explanation>
<why this change was needed>
<what impact it has>
```

### Examples

#### Good Commit Messages

```
feat: Add support for IPv6 networking in EKS clusters

Adds new `ipv6_enabled` argument to `rafay_eks_cluster` resource, allowing users
to enable IPv6 for their cluster networking. This includes support for dual-stack
configurations.

Closes #445
```

```
fix: Prevent state drift when node groups modified externally

Previously, external modifications to EKS node groups would cause persistent
state drift. This fix properly handles node group updates from AWS console
or CLI.

Fixes #447
```

```
deprecate: Remove project_id in favor of metadata.project

The flat `project_id` argument is deprecated in favor of the structured
`metadata.project` field for consistency with Kubernetes conventions.

Migration path: Update configurations to use metadata { project = "..." }
instead of project_id = "...".

Deprecated in v1.5.0, will be removed in v2.0.0.
```

#### Avoid These Patterns

❌ **Too vague**
```
fix: update code
feat: add new feature
chore: improvements
```

❌ **Technical jargon without context**
```
refactor: extract parseNodeGroups method
fix: nil pointer dereference in line 234
```

❌ **Focusing on implementation rather than impact**
```
Add new struct field for IPv6 configuration
Update Go version in mod file
```

## Commit Type Prefixes

While not required, these prefixes help with automatic categorization:

### Feature Additions
- `feat:` - New resources, data sources, or major functionality
- `add:` - Adding new capabilities to existing resources

**Examples:**
```
feat: Add `rafay_environment_template` resource
add: Support custom IAM roles in EKS clusters
```

### Bug Fixes
- `fix:` - Fixing incorrect behavior
- `patch:` - Small corrections

**Examples:**
```
fix: Correct timeout handling for large cluster operations
patch: Fix typo in error message
```

### Deprecations
- `deprecate:` - Deprecating existing functionality
- `breaking:` - Breaking changes (use sparingly)

**Examples:**
```
deprecate: Mark simple tags map as deprecated
breaking: Remove deprecated rafay_cluster resource
```

### Enhancements
- `enhance:` - Improvements to existing features
- `improve:` - Performance or UX improvements
- `update:` - Updating dependencies or configurations

**Examples:**
```
enhance: Add validation for subnet configurations
improve: Optimize cluster status polling
update: Support latest AWS EKS API version
```

### Documentation
- `docs:` - Documentation changes
- `example:` - Example code updates

**Examples:**
```
docs: Add guide for multi-region deployments
example: Update EKS cluster example with IPv6
```

### Internal Changes (Usually Skipped in Changelog)
- `refactor:` - Code restructuring without behavior change
- `test:` - Test additions or modifications
- `chore:` - Maintenance tasks
- `ci:` - CI/CD changes

**Examples:**
```
refactor: Extract shared validation logic
test: Add unit tests for state migration
chore: Update Go dependencies
ci: Add automated changelog workflow
```

## Special Cases

### Breaking Changes

Always explicitly state breaking changes in both the commit message and PR description:

```
breaking: Change default capacity_type from ON_DEMAND to MIXED

BREAKING CHANGE: The default `capacity_type` for EKS node groups has changed
from "ON_DEMAND" to "MIXED" for better cost optimization.

Users who want to preserve the old behavior should explicitly set:
capacity_type = "ON_DEMAND"

This change applies to new clusters created in v2.0.0 and later.
```

### Deprecations in Code

When adding deprecation warnings to Go code, mention it in the commit:

```
deprecate: Add deprecation warning for project_id argument

Adds `Deprecated` field to project_id schema definition with migration
guidance. The field will be removed in v2.0.0.

Users should migrate to:
metadata {
  project = "project-name"
}
```

### Multiple Related Changes

If a PR contains multiple related changes, use one commit per logical change, or describe all changes in the commit body:

```
feat: Add comprehensive IAM configuration for EKS

- Add support for OIDC identity providers
- Add pod identity association configuration
- Add IAM service account management
- Add well-known IAM policies (EBS CSI, VPC CNI, etc.)

This enables users to configure cluster IAM without manual AWS console work.

Closes #430, #432, #435
```

### Cherry-Picked Commits

When cherry-picking to a release branch, the commit message is automatically reused. Ensure the original commit message is clear and complete.

## PR Descriptions

Your PR description complements commit messages. Include:

1. **Why**: Business or technical motivation
2. **What**: High-level description of changes
3. **Impact**: Who is affected and how
4. **Testing**: How you verified the changes
5. **Breaking Changes**: Explicit callout if any
6. **Deprecations**: Migration path if applicable

### Example PR Description

```markdown
## Summary
Adds support for IPv6 networking in EKS clusters, addressing user requests for dual-stack configurations.

## Changes
- Add `ipv6_enabled` boolean argument to rafay_eks_cluster
- Add `ipv6_cidr_block` computed attribute
- Update cluster creation logic to handle IPv6 configuration
- Add validation for IPv6 subnet configurations

## Impact
- **Users**: Can now enable IPv6 on new EKS clusters
- **Backwards Compatible**: Existing clusters unaffected (defaults to IPv4)
- **Terraform State**: No migration needed

## Testing
- [x] Unit tests for IPv6 validation logic
- [x] Acceptance tests with IPv6-enabled cluster
- [x] Manual testing with dual-stack configuration
- [x] Verified state import/export works correctly

## Documentation
- [x] Updated resource documentation
- [x] Added example configuration
- [x] Updated migration guide

Closes #445
```

## Skipping Changelog

For internal changes that shouldn't appear in the changelog, add the `skip-changelog` label to the PR or include `[skip-changelog]` in the commit message:

```
chore: update internal test fixtures [skip-changelog]

Reorganizes test fixture files for better maintainability.
No user-facing changes.
```

## Best Practices Summary

### DO
✅ Write clear, user-focused descriptions  
✅ Explain the "why" not just the "what"  
✅ Use present tense  
✅ Include issue references  
✅ Mention deprecations explicitly  
✅ Callout breaking changes clearly  

### DON'T
❌ Use vague descriptions like "fix bug" or "update code"  
❌ Focus on implementation details users don't care about  
❌ Forget to mention breaking changes or deprecations  
❌ Write overly technical commit messages  
❌ Include internal refactoring in user-facing changelog  

## Questions?

If you're unsure about commit message format:
1. Check recent commits in the repository for examples
2. Review generated CHANGELOG.md and .changelog/{PR_number}.txt entries to see what works well
