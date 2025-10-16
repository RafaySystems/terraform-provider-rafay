output "cluster_name" {
  value = rafay_eks_cluster.basic.cluster[0].metadata[0].name
  description = "The name of the EKS cluster"
}