#Basic example for opa profile
resource "rafay_opa_installation_profile" "tftestopaprofileeleven" {
  metadata {
    name    = "tftestopaprofileeleven"
    project = "terraform"
  }
  spec {
    version = "v1"
    installation_params {
      audit_interval              = 60
      audit_match_kind_only       = true
      constraint_violations_limit = 20
      audit_chunk_size = 20
      log_denies = true
      emit_audit_events = true
    }  
    sync_objects{
      version = "v1"
      kind = "ConfigMap"
      group = "tfuser"
    }
    excluded_namespaces {
      namespaces {
        name = "testtwelve"
      }
    }
    sharing {
      enabled = true
      projects {
        name = "defaultproject"
      }
    }
  }
}