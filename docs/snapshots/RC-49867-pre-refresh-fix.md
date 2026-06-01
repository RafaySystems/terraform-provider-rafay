# Snapshot: before EKS refresh fix (2026-05-29)

Revert point before fixing post-refresh plan drift (`+ ng-2` after refresh) and reorder validation errors during refresh.

## Git references

- **Branch (code at parent commit):** `snapshot/RC-49867-pre-refresh-fix-20260529`
- **Stash (was applied back):** `snapshot-before-eks-refresh-fix-20260529-220337` (dropped after `git stash pop`)

## What the refresh fix changes (after this snapshot)

1. **Read/flatten list mode** — preserve state-only node groups missing from API response; rebuild `*_map` from merged list.
2. **ModifyPlan reorder-only** — keep state list order in plan (no positional diff when blocks are reordered in HCL).

## Revert working tree to pre-fix implementation

```bash
git checkout snapshot/RC-49867-pre-refresh-fix-20260529 -- \
  internal/provider/eks_cluster_resource.go \
  internal/resource_eks_cluster/eks_cluster_resource_expand.go \
  internal/resource_eks_cluster/eks_cluster_resource_flatten.go \
  docs/guides/eks-node-group-migration.md

git clean -fd internal/resource_eks_cluster/eks_cluster_resource_nodegroup_*.go \
  internal/resource_eks_cluster/eks_cluster_resource_schema_patch.go
```

Or reset all EKS node-group work to branch tip:

```bash
git checkout snapshot/RC-49867-pre-refresh-fix-20260529
```

## Rebuild after revert

```bash
make build
```
