resource "rafay_aks_cluster" "cluster" {
  name        = "demo-terraform"
  projectname = "dev"
  blueprint   = "default-aks"
  #blueprintversion = ""
  cloudprovider = "azure"
  cluster_config {
    resource_group_name = "dev-demo"
    #identity_type          = "SystemAssigned"
    location               = "centralindia"
    enable_private_cluster = true
    dnsprefix              = "demo-test-dns"
    kubernetesversion      = "1.21.7"
    #loadbalancer_sku       = "standard"
    network_plugin = "kubenet"
    #network_policy         = "calico"
    #sku_name               = "Basic"
    #sku_tier               = "Free"
    #tags = {
    #   cluster-id     = "Terraform Default Tags"
    #   owner          = "Alt Owner"
    #}
    #type                   = "Microsoft.ContainerService/managedClusters"
    node_pools {
      location             = "centralindia"
      name                 = "primary"
      count                = 1
      enable_autoscaling   = true
      max_count            = 2
      max_pods             = 40
      min_count            = 1
      mode                 = "System"
      orchestrator_version = "1.21.7"
      os_type              = "Linux"
      vm_size              = "Standard_DS2_v2"
      type                 = "Microsoft.ContainerService/managedClusters/agentPools"
      #availability_zones   = ["1", "2"]
      #apiversion           = ""
      #property_type        = "VirtualMachineScaleSets"
      #node_labels = {
      #     cluster-label = "Terraform-Label"
      #     owner  = "Alt Owner"
      #}
    }
  }
}
