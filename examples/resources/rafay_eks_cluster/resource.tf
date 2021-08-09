resource "rafay_eks_cluster" "cluster" {
  name         = "demo-terraform1"
  projectname  = "dev3"
  yamlfilepath = "<file-path/eks-cluster.yaml>"
  waitflag     = "1"
}
