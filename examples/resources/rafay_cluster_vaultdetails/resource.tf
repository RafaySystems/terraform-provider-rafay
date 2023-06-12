resource "rafay_cluster_vaultdetails" "samplecluster" {
  cluster_name = "mini-vault-csi"
  project_name = "defaultproject"
}

output "kubernetes_host" {
  description = "kubernetes host"
  sensitive   = true
  value       = rafay_cluster_vaultdetails.samplecluster.kubernetes_host
}

output "kubernetes_ca_cert" {
  description = "kubernetes ca certificate"
  sensitive   = true
  value       = rafay_cluster_vaultdetails.samplecluster.kubernetes_ca_cert
}
