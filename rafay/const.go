package rafay

// UserAgent for Terraform requests. It is currently used by v1
// clusters (AKS and EKS) to identiy source of the request.
const uaDef = "terraform"

// clusterSharingExt is used to indicate sharing of cluster is managed
// by dedicated resource either by `rafay_cluster_sharing_single` or
// `rafay_cluster_sharing`.
const clusterSharingExt = "true"

// clusterSharingExtKey is the key in edge.Settings to store cluster
// sharing external value. Its value can be "true" or "false".
const clusterSharingExtKey = "cluster_sharing_external"
