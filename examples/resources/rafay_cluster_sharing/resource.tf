#cluster sharing example, share to specific projects
resource "rafay_cluster_sharing" "demo-terraform-specific" {
  clustername = "demo-terraform"
  project = "terraform"
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


#cluster sharing example, share to ALL projects
resource "rafay_cluster_sharing" "demo-terraform-all" {
  clustername = "demo-terraform"
  project = "terraform"
  sharing {
    all = true
  }
}