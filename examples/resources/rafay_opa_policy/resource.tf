#Basic example for opa policy
resource "rafay_opa_policy" "tftestopapolicy1" {
  metadata {
    name    = "tftestopapolicy1"
    project = "terraform"
  }
  spec {
    constraint_list {
      name = "se-linux"
    }
    installation_params {
      audit_interval              = 60
      audit_match_kind_only       = true
      constraint_violations_limit = 20
      audit_chunk_size = 20
      log_denies = true
      emit_audit_events = true
    }
    sharing {
      enabled = true
      projects {
        name = "defaultproject"
      }
    }
    version = "v0"
    sync_objects{
      version = "v1"
      kind = "ConfigMap"
      group = "tfuser"
    }
    excluded_namespaces {
      namespaces {
        name = "tfdemonamespace"
      }
    }
  }
}
