resource "rafay_aks_cluster" "cluster" {
  name          = "demo-terraform"
  projectname   = "dev"
  cloudprovider = "azure"
  cluster_config {
    resource_group_name = "dev-demo"
    location            = "centralindia"
    kubernetesversion   = "1.21.7"
    node_pools {
      name                 = "primary"
      count                = 1
      max_count            = 1
      max_pods             = 40
      min_count            = 1
      orchestrator_version = "1.21.7"
      vm_size              = "Standard_DS2_v2"
    }
  }
}
