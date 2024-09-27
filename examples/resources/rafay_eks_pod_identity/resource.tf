resource "rafay_eks_pod_identity" "pod_identity_1" {
  metadata {
    cluster_name = "eks_cluster_name"
    project_name = "defaultproject"
  }
  spec {
    service_account_name = "svc_one"
    namespace = "rafay-demo"
    create_service_account = true
    role_arn = "arn:aws:iam::679196758854:role/rafay-eks-full"
  }
}