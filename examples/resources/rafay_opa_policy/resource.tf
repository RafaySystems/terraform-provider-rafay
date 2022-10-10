#Basic example for opa policy
resource "rafay_opa_policy" "tftestopapolicy1" {
  metadata {
    name    = "tftestopapolicy1"
    project = "terraform"
  }
  spec {
    constraint_list {
      name = "tfdemoopaconstraint1"
      version = "v1"
    }
    sharing {
      enabled = false
    }
    version = "v0"
  }
}
