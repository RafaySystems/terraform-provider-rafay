output "cluster_name" {
  value = rafay_eks_cluster.ha_multi_az.cluster[0].metadata[0].name
  description = "The name of the EKS cluster"
}