#Basic example for opa policy
resource "rafay_opa_policy" "test_123" {
  metadata {
    name    = "test_123"
    project = "kbr-test"
  }
  spec {
    constraint_list {
        # name = "one"
         #name = "host-network-ports"
        # name = "host-namespace"
        name = "seccomp"
        # name = "host-filesystem"
    }
    installation_params {
        audit_interval = 60
        audit_match_kind_only = true
        constraint_violations_limit = 20
    }
    sharing {
      enabled = false
    }
    version = "v1"
  }
}

