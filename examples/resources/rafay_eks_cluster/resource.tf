resource "rafay_eks_cluster" "cluster" {
  name             = "demo-terraform"
  projectname      = "dev"
  yamlfilepath     = "/Users/krishna/code/src/github.com/RafaySystems/terraform-provider-rafay/examples/resources/rafay_eks_cluster/eks-cluster.yaml"
}
