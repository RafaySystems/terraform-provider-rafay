resource "rafay_eks_cluster" "cluster" {
  name             = "demo-terraform6"
  projectname      = "dev3"
  yamlfilepath     = "/Users/krishna/code/src/github.com/RafaySystems/terraform-provider-rafay/examples/resources/rafay_eks_cluster/eks-cluster.yaml"
  blueprintflag    = "1"
  blueprintname    = ""
  blueprintversion = ""
  alertflag        = ""
}
