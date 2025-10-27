# Deprecation Policy

## Overview

This document outlines the deprecation policy and semantic versioning for the Rafay Terraform Provider, establishing clear guidelines for managing breaking changes, version upgrades, and maintaining backward compatibility. This policy ensures predictable upgrade paths and maintains user trust during provider evolution.

## Semantic Versioning Strategy

The Rafay Terraform Provider follows strict Semantic Versioning (SemVer) using the format `MAJOR.MINOR.PATCH`:

### MAJOR Version Changes (Breaking Changes)

**When to increment:** Incompatible API changes that break existing configurations.

**Examples:**
- Removing resources (e.g., removing `rafay_legacy_cluster`)
- Removing or renaming resource fields (e.g., removing `project_id` attribute from `rafay_cluster`)
- Changing default behavior that breaks existing configurations
- Schema changes that require state migration
- Removing/renaming resource data sources/types
- Changing resource names (config refactoring)
- Changing required vs optional field status

**Version Example:** `1.1.51` → `2.0.0`

### MINOR Version Changes (Backward Compatible)

**When to increment:** Adding new functionality in a backward-compatible manner.

**Examples:**
- Adding new resources (e.g., adding `rafay_environment_template`)
- Adding new optional arguments to existing resources
- Adding data sources
- New resources that don't break existing configurations
- Adding new computed attributes

**Version Example:** `1.1.51` → `1.2.0`

### PATCH Version Changes (Bug Fixes)

**When to increment:** Backward-compatible bug fixes and improvements.

**Examples:**
- Fixing diff suppression bugs
- Correcting documentation
- Tightening validation without breaking existing usage
- Patching crashes without breaking existing usage
- Fixing `terraform import` issues
- Performance improvements without behavior changes

**Version Example:** `1.1.51` → `1.1.52`

## Deprecation Process

Following the [AWS provider deprecation model](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/version-2-upgrade) and [version 3 upgrade guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/version-3-upgrade), our deprecation process ensures users have adequate time to migrate.

### Deprecation Timeline

Based on HashiCorp Terraform recommended best practices ([Framework Deprecations](https://developer.hashicorp.com/terraform/plugin/framework/deprecations)):

1. **Phase 1: Deprecated** (Version N)
   - **Status**: Feature is marked for removal in a future major release
   - Announce deprecations with at least one minor version lead time (minimum 3 months)
   - Add practitioner-focused deprecation warnings using `DeprecationMessage` field
   - Update documentation with deprecation notices and migration paths
   - Include in `CHANGELOG.md` under **DEPRECATIONS** section
   - Provide automated detection tools where possible
   - Feature remains fully functional with warnings

2. **Phase 2: Pending Removal** (Version N+1 to N+X)
   - **Status**: Feature behavior fundamentally altered; use strongly discouraged
   - Maintain backward compatibility for a grace period
   - Provide comprehensive migration examples and upgrade guides
   - Continue deprecation warnings with version-specific messaging
   - Offer automated migration tools and state migration utilities
   - May include runtime warnings or validation errors

3. **Phase 3: Removed** (Next Major Version)
   - **Status**: Feature no longer supported and has been removed
   - Implement breaking changes in major version releases only
   - Remove deprecated functionality with clear upgrade documentation
   - Provide automated state migration where technically feasible
   - Include detailed upgrade guides with before/after examples
   - Keep state upgraders for historical migrations

### Minimum Deprecation Periods

Following HashiCorp and AWS provider standards:

- **Resources:** Minimum 6 months or 2 minor versions, whichever is longer
- **Arguments/Attributes:** Minimum 3 months or 1 minor version, whichever is longer  
- **Data Sources:** Minimum 3 months or 1 minor version, whichever is longer
- **Provider Configuration:** Minimum 9 months or 3 minor versions, whichever is longer
- **Default Behavior Changes:** Minimum 6 months with opt-in mechanisms where possible

**Note:** These are minimum periods. Security vulnerabilities may require immediate breaking changes, bypassing normal deprecation timelines with expedited communication to users.

## Implementation Guide: Go Code to Terraform Resource Mapping

This section demonstrates how Go code deprecation warnings translate to user-facing Terraform warnings.

### Deprecation Message Best Practices

Following HashiCorp's practitioner-focused messaging guidelines:

**✅ GOOD Examples (Actionable and Clear):**
- `"Configure metadata.project instead. This attribute will be removed in the next major version of the provider."`
- `"Configure network_config block instead. This block will be removed in the next major version of the provider."`
- `"Use rafay_eks_cluster resource instead. This resource will be removed in v2.0.0."`

**❌ BAD Examples (Too Technical or Vague):**
- `"Attribute project_id is deprecated"` - Not actionable, doesn't tell users what to do
- `"This field is going away"` - No timeline or alternative provided
- `"Deprecated in favor of new implementation"` - Not specific enough
- `"Do not use this anymore"` - No migration path provided

**Message Format Guidelines:**

1. **Start with Action**: Begin with "Configure" or "Use" to tell users what to do
   - ✅ `"Configure metadata.project instead..."`
   - ❌ `"The attribute project_id is deprecated..."`

2. **Specify Alternative**: Clearly name the replacement feature
   - ✅ `"...instead. Use network_config block..."`
   - ❌ `"...instead. Use the new option..."`

3. **Include Timeline**: State when removal will occur
   - ✅ `"...will be removed in the next major version of the provider"`
   - ✅ `"...will be removed in v2.0.0"`
   - ❌ `"...will be removed soon"`

4. **Keep it Concise**: One or two sentences maximum
   - Users see these warnings during normal workflow
   - Long messages are often skipped

5. **Consistent Terminology**:
   - Use "attribute" for scalar values
   - Use "block" for nested configuration structures
   - Use "resource" for entire resource types
   - Use "data source" for data sources

## Version Support Policy

### Backward Compatibility Promise

Following the [HashiCorp Terraform AWS Provider model](https://hashicorp.github.io/terraform-provider-aws/faq/), our backward compatibility policy is:

**Once a major release is published, will new features and fixes be backported to previous versions?**

**Generally, no.** New features and fixes will only be added to the most recent major version. When a new major version is released, previous major versions become **static** and will not receive:
- ❌ New features (FEATURES in changelog)
- ❌ Enhancements (ENHANCEMENTS in changelog)  
- ❌ New resources or data sources
- ❌ Minor version updates
- ❌ Non-critical bug fixes

**Exception for Security Vulnerabilities:**  
Critical security vulnerabilities are reviewed on a case-by-case basis. Backporting security fixes to previous major versions may occur when it is the most reasonable course of action to protect users.

**Rationale:**  
Due to the high-touch nature of provider development and the extensive regression testing required to ensure stability, maintaining multiple active major versions is not sustainable. This approach allows the team to focus on delivering the best experience on the current major version.

**Recommendation:**  
Users should plan to upgrade to the latest major version to receive new features, enhancements, and bug fixes. It is generally recommended to pin the provider version in your configuration and test upgrades in non-production environments first.

### Support Windows
- **Current Major Version (N):** Full support including new features, enhancements, and bug fixes
- **Previous Major Version (N-1):** Static - no new features or enhancements; security updates only (case-by-case basis)
- **Older Major Versions (N-2 and below):** End-of-life, no support

### Support Timeline Details

**Minor Versions (within same major version):**
- Latest minor version receives all updates
- Previous minor versions: Users should upgrade to latest minor within the same major version
- No backporting to older minor versions within a major version

**Major Version Lifecycle:**
- **Active Development:** Current major version (N) receives all updates
- **Security-Only Period:** Previous major version (N-1) may receive critical security fixes for 6 months after new major release
- **End-of-Life:** After 6 months, previous major versions receive no updates

### Migration Assistance
- Automated migration tools for common scenarios
- Step-by-step migration guides in dedicated upgrade documentation
- Professional services for complex migrations
- Community support during major version transitions

### Example Version Lifecycle

**Scenario:** Provider is at v1.8.9, and v2.0.0 is about to be released

**Before v2.0.0 release:**
```
v1.8.9 (Current) → Receives all updates:
  ✅ New features
  ✅ Enhancements
  ✅ Bug fixes
  ✅ Security patches
```

**After v2.0.0 release (Day 1):**
```
v2.0.0 (Current) → Receives all updates:
  ✅ New features
  ✅ Enhancements  
  ✅ Bug fixes
  ✅ Security patches

v1.8.9 (Previous) → STATIC (security-only period):
  ❌ New features
  ❌ Enhancements
  ❌ Bug fixes
  ⚠️ Security patches (case-by-case, 6 months only)
```

**After v2.0.0 release (6+ months):**
```
v2.0.0+ (Current) → Receives all updates:
  ✅ New features
  ✅ Enhancements
  ✅ Bug fixes
  ✅ Security patches

v1.x (Previous) → END OF LIFE:
  ❌ No updates of any kind
  ❌ No security patches
  ⚠️ Users must upgrade to v2.x
```

**Key Takeaway:** Plan major version upgrades proactively. Previous major versions become static immediately upon new major release.

## Communication Strategy

### CHANGELOG.md Integration

Based on AWS provider [changelog patterns](https://github.com/hashicorp/terraform-provider-aws/releases), automated changelog generation via GitHub Actions with clear categorization:

- **BREAKING CHANGES:** Schema changes, resource removals, behavior changes requiring user action
- **FEATURES:** New resources, data sources, and major functionality additions
- **ENHANCEMENTS:** Improvements to existing resources, performance optimizations, new optional arguments
- **BUG FIXES:** Issue resolutions, crash fixes, and patches
- **DEPRECATIONS:** Advance notice of upcoming changes with removal timelines
- **DOCUMENTATION:** Documentation updates, example improvements, and clarifications

### AWS Provider Changelog Format Example

Following the standardized format from AWS provider releases:

```markdown
## 2.0.0 (January 15, 2025)

BREAKING CHANGES:

* provider: Remove deprecated `project_id` argument from all cluster resources. Use `metadata.project` instead ([#123](https://github.com/RafaySystems/terraform-provider-rafay/issues/123))
* resource/rafay_eks_cluster: Change default `capacity_type` from `ON_DEMAND` to `MIXED` for cost optimization ([#124](https://github.com/RafaySystems/terraform-provider-rafay/issues/124))
* resource/rafay_aks_cluster: Remove deprecated `tags` argument. Use `resource_tags` block instead ([#125](https://github.com/RafaySystems/terraform-provider-rafay/issues/125))

FEATURES:

* **New Resource:** `rafay_environment_template` ([#126](https://github.com/RafaySystems/terraform-provider-rafay/issues/126))
* **New Data Source:** `rafay_cost_profiles` ([#127](https://github.com/RafaySystems/terraform-provider-rafay/issues/127))
* resource/rafay_eks_cluster: Add `auto_create_node_group` argument for intelligent defaults ([#128](https://github.com/RafaySystems/terraform-provider-rafay/issues/128))

ENHANCEMENTS:

* resource/rafay_aks_cluster: Add `resource_tags` block with propagation support ([#129](https://github.com/RafaySystems/terraform-provider-rafay/issues/129))
* data-source/rafay_eks_clusters: Add enhanced filtering with `label_selector` and version ranges ([#130](https://github.com/RafaySystems/terraform-provider-rafay/issues/130))
* provider: Add structured `api_credentials`, `endpoints`, and `retry_config` blocks ([#131](https://github.com/RafaySystems/terraform-provider-rafay/issues/131))

BUG FIXES:

* resource/rafay_eks_cluster: Fix state inconsistency when node groups are modified outside Terraform ([#132](https://github.com/RafaySystems/terraform-provider-rafay/issues/132))
* resource/rafay_aks_cluster: Prevent `terraform import` failures for clusters with custom node pools ([#133](https://github.com/RafaySystems/terraform-provider-rafay/issues/133))

DEPRECATIONS:

* data-source/rafay_clusters: Deprecate in favor of type-specific data sources `rafay_eks_clusters` and `rafay_aks_clusters`. Will be removed in v3.0.0 ([#134](https://github.com/RafaySystems/terraform-provider-rafay/issues/134))
```

### User Notification Process

Following HashiCorp and AWS provider communication patterns with multi-channel approach:

1. **In-Code Deprecation Warnings** (Primary Channel)
   
   Terraform Plugin Framework automatically generates warnings during `terraform plan` and `terraform apply`:
   ```
   │ Warning: Attribute Deprecated
   │ 
   │   with rafay_eks_cluster.example,
   │   on main.tf line 5, in resource "rafay_eks_cluster" "example":
   │    5:   project_id = "my-project"
   │ 
   │ Configure metadata.project instead. This attribute will be removed in the
   │ next major version of the provider.
   ```

2. **Provider Log Messages** (Development/Debug)
   ```
   [WARN] rafay_cluster resource is deprecated and will be removed in version 2.0.0. Use rafay_eks_cluster instead. See upgrade guide: https://registry.terraform.io/providers/RafaySystems/rafay/latest/docs/guides/version-2-upgrade
   
   [WARN] rafay_aks_cluster.example: Argument "project_id" is deprecated and will be removed in version 2.0.0. Use "metadata.project" instead.
   
   [WARN] rafay_aks_cluster.example: Argument "tags" is deprecated and will be removed in version 2.0.0. Use "resource_tags" block instead for enhanced tagging capabilities.
   ```

3. **Documentation Updates** (Reference Material)
   - **Deprecation notices** prominently displayed at the top of resource documentation
   - **Warning banners** on affected pages with timeline and migration path
   - **Migration examples** with side-by-side comparisons (before/after)
   - **Clear timelines** with specific version numbers
   - **Links to upgrade guides** and automated migration tools
   - **Search optimization** to help users find deprecation information

4. **CHANGELOG.md** (Version History)
   - **DEPRECATIONS section** in every release that introduces deprecations
   - **BREAKING CHANGES section** in major versions
   - Consistent formatting following AWS provider patterns
   - Direct links to migration guides and related issues

5. **GitHub Releases** (Release Announcements)
   - Detailed upgrade notes with each release following AWS provider format
   - Prominent deprecation warnings in release notes
   - Direct links to migration guides and documentation
   - Highlight breaking changes with visual markers
   - Provide downloadable migration scripts where applicable

6. **Upgrade Guides** (Detailed Migration Instructions)
   - Dedicated upgrade guide for each major version (e.g., `version-2-upgrade.md`)
   - Step-by-step migration instructions
   - Before/after code examples
   - Common pitfalls and troubleshooting
   - Automated migration tool documentation

**Multi-Channel Strategy Benefits:**
- **Proactive Warnings**: Users see warnings in their workflow (`terraform plan`)
- **Reference Documentation**: Users can research deprecations before encountering them
- **Version Control**: Changelog provides historical context
- **Search Discoverability**: Multiple channels improve search engine visibility

## State Migration Support

### Automatic State Migration

For compatible schema changes:
- Built-in state migration during `terraform plan`
- Validation and integrity checks
- Automatic backup creation before migration
- Rollback procedures for failed migrations

### Manual Migration Procedures

For complex changes requiring user intervention, following AWS provider migration patterns:

1. **Resource Renames**
   ```bash
   # Example: Migrating from rafay_cluster to rafay_eks_cluster
   
   # Step 1: Remove the old resource from state (without destroying)
   terraform state rm rafay_cluster.example
   
   # Step 2: Import the existing cluster with new resource type
   terraform import rafay_eks_cluster.example cluster-id
   
   # Step 3: Update configuration file to use new resource
   # (See deprecation examples above for configuration changes)
   
   # Step 4: Verify the plan shows no changes
   terraform plan
   ```

2. **Argument Restructuring**
   ```bash
   # Example: Migrating from project_id to metadata.project
   
   # Step 1: Update configuration file
   # Change from:
   #   project_id = "my-project"
   # To:
   #   metadata {
   #     project = "my-project"
   #   }
   
   # Step 2: Run terraform plan to validate changes
   terraform plan
   
   # Step 3: Apply changes (should show in-place update)
   terraform apply
   ```

3. **Provider Configuration Migration**
   ```bash
   # Example: Migrating to structured provider configuration
   
   # Step 1: Update provider block configuration
   # (See provider configuration examples above)
   
   # Step 2: Re-initialize Terraform
   terraform init -upgrade
   
   # Step 3: Verify connectivity with new configuration
   terraform plan
   ```

4. **Complex Data Structure Migration**
   ```bash
   # Example: Migrating from simple tags to resource_tags block
   
   # Step 1: Export current state for backup
   terraform show -json > backup-state.json
   
   # Step 2: Update configuration to use new structure
   # (See argument deprecation examples above)
   
   # Step 3: Plan and apply changes
   terraform plan
   terraform apply
   
   # Step 4: Verify tags are correctly applied to cloud resources
   # Check via cloud provider console or CLI
   ```

### Automated Migration Tools

Following AWS provider automation patterns:

```bash
# Rafay Provider Migration CLI Tool (planned for v2.0 release)

# Check for deprecated usage in current configuration
rafay-migrate scan --path ./terraform/

# Generate migration plan
rafay-migrate plan --from-version 1.x --to-version 2.0

# Apply automated migrations where possible
rafay-migrate apply --backup-state

# Validate migration results
rafay-migrate validate --post-migration-check
```

## Upgrade Documentation Structure

```
docs/
├── guides/
│   ├── version-2-upgrade.md      # Major version upgrade guide
│   ├── version-1.5-upgrade.md    # Minor version with deprecations
│   └── migration-examples/       # Specific migration scenarios
│       ├── cluster-migration.md
│       ├── rbac-migration.md
│       └── state-migration.md
├── UPGRADE.md                    # Current upgrade notes
└── CHANGELOG.md                  # Automated changelog
```

## Implementation Guidelines

### For Developers

1. **Before Deprecating:**
   - Ensure replacement functionality exists
   - Create comprehensive migration documentation
   - Implement deprecation warnings
   - Update tests to cover both old and new approaches

2. **During Deprecation Period:**
   - Maintain backward compatibility
   - Provide migration assistance
   - Update documentation regularly

3. **Before Removal:**
   - Confirm adequate notice period has passed
   - Verify migration paths are well-documented
   - Ensure automated tools are available
   - Coordinate with release management

### For Users

1. **Stay Informed:**
   - Subscribe to release notifications
   - Review changelog for each update
   - Monitor provider logs for warnings

2. **Plan Migrations:**
   - Test migrations in non-production environments
   - Review upgrade guides before applying updates
   - Schedule migration windows appropriately
   - Maintain backup configurations

## Exceptions and Special Cases

### Emergency Deprecations and Security Vulnerabilities

Following HashiCorp best practices for handling security-critical situations:

**Security Vulnerabilities:**
- **Immediate Action Required**: Security vulnerabilities may require immediate breaking changes, bypassing normal deprecation timelines
- **No Minimum Notice Period**: When security is at risk, changes may be implemented immediately without the standard grace period
- **Enhanced Communication**: 
  - Immediate security advisory with CVE details (if applicable)
  - Clear explanation of the vulnerability and impact
  - Urgent upgrade recommendations
  - Direct notification to known users via multiple channels
- **Expedited Documentation**: Fast-tracked documentation updates and migration guides
- **Backport Policy**: Critical security fixes may be backported to previous major version (N-1) on a case-by-case basis, reviewed within 6 months of major release

**Critical Bug Emergency Deprecations:**
For non-security critical bugs requiring rapid action:
- **Minimum 30-day notice** for emergency deprecations
- Immediate patch release with fixes
- Accelerated migration assistance with direct support
- Clear communication about urgency and impact
- Temporary workarounds provided where possible

**Communication Priority Order:**
1. Security advisories (for vulnerabilities)
2. GitHub releases with emergency tag
3. Provider changelog with prominent warnings
4. Documentation homepage notices
5. Community channels (forums, slack, etc.)
6. Direct email to registered users (if available)

## Compliance and Monitoring

### Deprecation Tracking

- Automated tracking of deprecated features
- Regular reviews of deprecation timelines
- Metrics on migration adoption rates
- User feedback collection and analysis

### Quality Assurance

- All deprecations must include migration paths
- Automated testing of upgrade scenarios
- Documentation review requirements

## Summary: HashiCorp Deprecation Best Practices

This policy aligns with HashiCorp's core deprecation principles:

### Key Principles

1. **Practitioner-Focused Communication**
   - Use clear, actionable language in deprecation messages
   - Tell users what to do, not just what's deprecated
   - Include timelines in all deprecation notices

2. **Phased Deprecation Approach**
   - **Phase 1 (Deprecated)**: Feature marked for removal, fully functional with warnings
   - **Phase 2 (Pending Removal)**: Behavior altered, use strongly discouraged
   - **Phase 3 (Removed)**: Feature removed in major version with migration support

3. **Multi-Channel Communication**
   - In-code warnings (primary channel for practitioners)
   - Documentation updates (reference and research)
   - Changelog entries (version history)
   - GitHub releases (announcements)
   - Upgrade guides (detailed migration instructions)

4. **Adequate Notice Periods**
   - Minimum 3 months for attributes and data sources
   - Minimum 6 months for resources
   - Minimum 9 months for provider configuration
   - Security exceptions may bypass timelines

5. **Backward Compatibility**
   - Deprecated features remain fully functional during grace period
   - Breaking changes only in major versions
   - State migration support for smooth upgrades
   - **Previous major versions become STATIC** - no new features or enhancements after new major release

6. **Clear Migration Paths**
   - Always provide alternatives before deprecation
   - Include side-by-side examples
   - Offer automated migration tools where possible
   - Maintain historical state upgraders

7. **Version Support Model** (Following [AWS Provider approach](https://hashicorp.github.io/terraform-provider-aws/faq/))
   - Only current major version receives features, enhancements, and bug fixes
   - Previous major versions: security patches only (case-by-case, 6 months max)
   - Older major versions: end-of-life, no support
   - Users should plan proactive upgrades to latest major version

### Plugin Framework Specifics

- Use `DeprecationMessage` field on both attributes and blocks
- Follow format: "Configure {alternative} instead. This {type} will be removed in the next major version of the provider."
- Update both `DeprecationMessage` and `Description` fields
- Terraform automatically generates warnings during plan/apply

### SDKv2 Specifics

- Use `Deprecated` field for schema attributes
- Use `DeprecationMessage` field for entire resources/data sources
- Custom validation functions for value-specific deprecations
- Implement `StateUpgraders` for schema version migrations

## Contact and Support

For questions about deprecations or migration assistance:

- **Documentation:** [Provider Documentation](https://registry.terraform.io/providers/RafaySystems/rafay/latest/docs)
- **GitHub Issues:** [Rafay TF Provider](https://github.com/RafaySystems/terraform-provider-rafay/issues)
- **Professional Services:** [Book a Demo](https://rafay.co/)

---

**Document Version:** 1.1  
**Last Updated:** October 2025  
**Next Review:** January 2026
**Change Summary:** Updated to align with HashiCorp Terraform Plugin Framework deprecation best practices

This deprecation policy ensures predictable, user-friendly evolution of the Rafay Terraform Provider while maintaining stability and trust in production environments. It follows HashiCorp's recommended best practices for practitioner-focused communication, phased deprecation timelines, and multi-channel notification strategies.
