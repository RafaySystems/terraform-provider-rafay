#cluster override sharing example, share to specific projects
resource "rafay_cluster_override_sharing" "demo-terraform-specific" {
  clusteroverridename = "demo-terraform"
  project = "terraform"
  clusteroverridetype = "ClusterOverrideTypeAddon"
  sharing {
    all = false
    projects {
      name = "terraform-1"
    }
    projects {
      name = "terraform-2"
    }
    projects {
      name = "terraform-3"
    }
  }
}

#cluster override sharing example, share with ALL projects
resource "rafay_cluster_override_sharing" "demo-terraform-none" {
  clusteroverridename = "demo-terraform"
  project = "terraform"
  clusteroverridetype = "ClusterOverrideTypeAddon"
  sharing {
    all = false
  }
}

#cluster override sharing example, share with ALL projects
resource "rafay_cluster_override_sharing" "demo-terraform-all" {
  clusteroverridename = "demo-terraform"
  project = "terraform"
  clusteroverridetype = "ClusterOverrideTypeAddon"
  sharing {
    all = true
  }
}