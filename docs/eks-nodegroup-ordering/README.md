# EKS Node Group Ordering Stability

This note documents the ordering-stability fix for `node_groups` and `managed_nodegroups` in `rafay_eks_cluster`, plus the local test harness and plan-only tests.

## What Changed

- Plan-time canonicalization (`CustomizeDiff`) now sorts:
  - `node_groups` and `managed_nodegroups` by `name`.
  - Nested list fields inside each nodegroup (except `pre_bootstrap_commands`), including:
    - `availability_zones`, `subnets`, `instance_types`, `asg_suspend_processes`
    - `classic_load_balancer_names`, `target_group_arns`
    - `security_groups.attach_ids`
    - `ssh.source_security_group_ids`
    - `iam.attach_policy_arns`
    - `instances_distribution.instance_types`
    - `asg_metrics_collection.metrics`
    - `taints` (ordered by `key`, then `effect`, then `value`)
- State-time canonicalization (flatten) mirrors the above to keep state stable across refreshes.

## Local Harness

The local harness lets you exercise Create/Read/Update without real backend calls.

- Path: `cmd/eks-local-backend/main.go`
- It stores state in a local JSON file and reads it back for `read`.

### Run Local Order Tests

```
tests/eks-order-samples/run-local-order-tests.sh
```

This script:
- Applies a set of scenarios using the local harness.
- Reads back and prints ordered `managed_nodegroups` and `node_groups`.
- Includes reorder tests for `taints`, `security_groups.attach_ids`, and `instance_types`.

If protobuf warnings appear during `go run`, set:

```
export GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn
```

The script already sets this env var.

## Plan-Only Tests

Plan-only tests live in:

- `rafay/tests/resource_eks_cluster_plan_test.go`

Run with:

```
go test ./rafay/tests -tags=planonly -run PlanOnly
```

## Sample Terraform Configs

Sample configs for ordering scenarios:

- `tests/eks-order-samples/managed-reorder.tf`
- `tests/eks-order-samples/node-reorder.tf`
- `tests/eks-order-samples/nested-reorder.tf`
