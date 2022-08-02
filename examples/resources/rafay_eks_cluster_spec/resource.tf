resource "rafay_eks_cluster_spec" "demo-terraform-eks" {
  name            = "demo-terraform-eks"
  projectname     = "terraform"
  yamlfilepath    = "/Users/testuser/terraform-provider-rafay/examples/resources/rafay_eks_cluster_spec/eks-cluster.yaml"
  yamlfileversion = "0"
  checkdiff       = true
}
