resource "rafay_eks_cluster" "cluster" {
  name         = "demo-terraform2"
  projectname  = "dev3"
  yamlfilepath = "<file-path/eks-cluster.yaml>"
}
