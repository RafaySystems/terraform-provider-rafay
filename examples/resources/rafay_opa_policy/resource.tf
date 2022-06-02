#Basic example for opa policy
resource "rafay_opa_policy" "tfdemoopapolicy1" {
  metadata {
    name    = "tfdemoopapolicy1"
    project = "tfdemoproject1"
  }
  spec {
    constraint_list {
      name = "one"
    }
    installation_params {
      audit_interval              = 60
      audit_match_kind_only       = true
      constraint_violations_limit = 20
    }
    sharing {
      enabled = false
    }
    version = "v23"
  }
}

