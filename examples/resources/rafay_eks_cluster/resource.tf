resource "rafay_eks_cluster" "cluster" {
  name         = "demo-terraform"
  projectname  = "dev"
  yamlfilepath = "<filepath/eks-cluster.yaml>"
  waitflag     = "1"
}
