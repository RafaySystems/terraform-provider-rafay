#Basic example for opa constraint
resource "rafay_opa_constraint_template" "opact" {
  metadata {
    name    = "ter-test"
    project = "defaultproject"
  }
  spec {
    artifact {
      artifact {
        paths {
          name = "file://k8sallowedrepos_temp.yaml"
        }
      }
      options {
        force = true
      }
      type = "Yaml"
    }
  }
}