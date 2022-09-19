#Basic example for opa constraint
resource "rafay_opa_constraint" "tfdemoopaconstraint1" {
  metadata {
    name    = "tfdemoopaconstraint1"
    project = "tfdemoproject1"
  }
  spec {
    template_name = "one"
    artifact {
      type = "Yaml"
      artifact {
        paths {
          name = "file://artifacts/one/request-limit-ratio-constraint.yaml"
        }
      }
    }
  }
}