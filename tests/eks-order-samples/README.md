# EKS Nodegroup Ordering Tests

This directory contains sample Terraform configs and a local-only test runner
for validating ordering stability of `node_groups` and `managed_nodegroups`.

## What Was Fixed

We canonicalize ordering for list fields inside nodegroups to prevent diffs
and churn when users insert items in the middle of a list. The fix covers:

- `node_groups` and `managed_nodegroups` list ordering (sorted by `name`)
- Nested list fields (sorted, with state alignment where possible):
  - `availability_zones`
  - `subnets`
  - `instance_types`
  - `asg_suspend_processes`
  - `classic_load_balancer_names`
  - `target_group_arns`
  - `security_groups.attach_ids`
- `taints` are ordered by `key`, `effect`, `value`

`pre_bootstrap_commands` is intentionally NOT sorted.

## Samples

- `managed-reorder.tf`: reorder `managed_nodegroups`, insert at start/middle/end
- `node-reorder.tf`: reorder `node_groups`, insert at start/middle/end
- `nested-reorder.tf`: reorder taints, attach IDs, instance types, suspend processes, LB names, target group ARNs

These are templates you can modify for your own scenarios.

## Local Runner (No Real EKS)

The local test runner uses the harness in `cmd/eks-local-backend` to
apply/read YAML without hitting real EKS.

### Run

From repo root:

```bash
./tests/eks-order-samples/run-local-order-tests.sh
```

Output shows expected vs actual ordering for each scenario.

If you hit repeated protobuf warnings, you can silence them with:

```bash
export GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore
```

## Files

- `run-local-order-tests.sh`: multi-scenario local runner
- `eks-order-check.go`: helper that prints ordering for validation
