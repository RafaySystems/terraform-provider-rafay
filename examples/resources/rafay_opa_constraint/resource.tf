#Basic example for opa constraint
resource "rafay_opa_constraint" "tfdemoopaconstraint1" {
  metadata {
    name    = "tfdemoopaconstraint1"
    project = "tfdemoproject1"
  }
  spec {
    template_name = "request-limit-ratio-template"
    artifact {
      type = "Yaml"
      artifact {
        paths {
          name = "file://artifacts/request-limit-ratio/request-limit-ratio-constraint.yaml"
        }
      }
    }
  }
}