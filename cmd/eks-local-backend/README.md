# EKS Local Backend Harness

A lightweight JSON/YAML backend to run create/read/update cycles for
`rafay_eks_cluster` without hitting real EKS.

## Usage

From repo root:

```bash
go run ./cmd/eks-local-backend apply --file ./cluster.yaml
```

```bash
go run ./cmd/eks-local-backend read --name <cluster> --project <project>
```

## Input

The YAML file must contain two documents:

1. `Cluster`
2. `ClusterConfig`

These match the EKS cluster spec and config structures.

## Storage

Data is stored in `.eks-local-backend.json` at the repo root by default.

