resource "rafay_eks_cluster" "cluster" {
  name         = "demo-terraform"
  projectname  = "dev"
  yamlfilepath = "<file-path/eks-cluster.yaml>"
}
