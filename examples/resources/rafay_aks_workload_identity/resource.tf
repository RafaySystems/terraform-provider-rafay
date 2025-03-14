# Workload Identity Resource with v1 cluster resource 
resource "rafay_aks_workload_identity" "demo-terraform" {
  metadata {
    cluster_name = "gautham-tf-wi-1"
    project      = "defaultproject"
  }

  spec {
    create_identity = true

    metadata {
      name           = "gautham-tf-wi-1-uai-1"
      location       = "centralindia"
      resource_group = "gautham-rg-ci"
      tags = {
        "owner"      = "gautham"
        "department" = "gautham"
      }
    }

    role_assignments {
      name  = "Key Vault Secrets User"
      scope = "/subscriptions/a2252eb2-7a25-432b-a5ec-e18eba6f26b1/resourceGroups/gautham-rg-ci/providers/Microsoft.KeyVault/vaults/gautham-keyvault"
    }

    service_accounts {
      create_account = true

      metadata {
        name      = "gautham-tf-wi-1-sa-10"
        namespace = "default"
        annotations = {
          "role" = "dev"
        }
        labels = {
          "owner"      = "gautham"
          "department" = "gautham"
        }
      }
    }

  }

  # Depends On is mandatory for the Workload Identity Resource to be created after the cluster resource is created.
  depends_on = [rafay_aks_cluster.demo-terraform-wi-cluster]
}

# Workload Identity resource for soft creation with pre created managed identity and pre created service account
resource "rafay_aks_workload_identity" "demo-terraform" {
  metadata {
    cluster_name = "gautham-tf-wi-1"
    project      = "defaultproject"
  }

  spec {
    create_identity = false

    metadata {
      name           = "gautham-tf-wi-1-uai-1"
      location       = "centralindia"
      resource_group = "gautham-rg-ci"
      client_id      = "a2252eb2-7a25-432b-a5ec-e18eba6f26b1"
      principal_id   = "a2252eb2-7a25-432b-a5ec-e18eba6f26b1"
      tags = {
        "owner"      = "gautham"
        "department" = "gautham"
      }
    }

    role_assignments {
      name  = "Key Vault Secrets User"
      scope = "/subscriptions/a2252eb2-7a25-432b-a5ec-e18eba6f26b1/resourceGroups/gautham-rg-ci/providers/Microsoft.KeyVault/vaults/gautham-keyvault"
    }

    service_accounts {
      create_account = false

      metadata {
        name      = "gautham-tf-wi-1-sa-10"
        namespace = "default"
        annotations = {
          "role" = "dev"
        }
        labels = {
          "owner"      = "gautham"
          "department" = "gautham"
        }
      }
    }

  }

  # Depends On is mandatory for the Workload Identity Resource to be created after the cluster resource is created.
  depends_on = [rafay_aks_cluster.demo-terraform-wi-cluster]
}

# Workload Identity Resource with v3 cluster resource 
resource "rafay_aks_workload_identity" "demo-terraform" {
  metadata {
    cluster_name = "gautham-tf-wi-1"
    project      = "defaultproject"
  }

  spec {
    create_identity = true

    metadata {
      name           = "gautham-tf-wi-1-uai-1"
      location       = "centralindia"
      resource_group = "gautham-rg-ci"
      tags = {
        "owner"      = "gautham"
        "department" = "gautham"
      }
    }

    role_assignments {
      name  = "Key Vault Secrets User"
      scope = "/subscriptions/a2252eb2-7a25-432b-a5ec-e18eba6f26b1/resourceGroups/gautham-rg-ci/providers/Microsoft.KeyVault/vaults/gautham-keyvault"
    }

    service_accounts {
      create_account = true

      metadata {
        name      = "gautham-tf-wi-1-sa-10"
        namespace = "default"
        annotations = {
          "role" = "dev"
        }
        labels = {
          "owner"      = "gautham"
          "department" = "gautham"
        }
      }
    }
  }

  # Depends On is mandatory for the Workload Identity Resource to be created after the cluster resource is created.
  depends_on = [rafay_aks_cluster_v3.demo-terraform-wi-cluster]
}


