PRs to be merged only after checking all these items:
- [ ] terraform schema update
- [ ] terraform docs update
- [ ] backward compatibility test

## Changelog (Automated)

This PR will be automatically included in the CHANGELOG upon merge. The changelog entry is generated using AI analysis of your commits.

**To help with accurate categorization:**
- Use clear, descriptive commit messages
- Explicitly mention if this introduces breaking changes
- Note any deprecations in commit messages or PR description
- Add `skip-changelog` label if this PR should not appear in the changelog (internal changes, test-only changes)

**Deprecations** in Go code (`Deprecated:` or `DeprecationMessage:`) will be automatically detected and included.
