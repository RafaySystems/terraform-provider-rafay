resource "rafay_eks_cluster_spec" "demo-terraform-eks1" {
  name            = "demo-terraform-eks1"
  projectname     = "dev"
  yamlfilepath    = "/Users/stephanbenny/code/src/github.com/RafaySystems/terraform-provider-rafay/examples/resources/rafay_eks_cluster_spec/eks-cluster.yaml"
  yamlfileversion = ""
  checkdiff       = true
}
