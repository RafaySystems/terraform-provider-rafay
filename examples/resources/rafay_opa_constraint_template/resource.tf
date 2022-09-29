#Basic example for opa constraint
resource "rafay_opa_constraint_template" "opact" {
  metadata {
    name    = "test_template"
    project = "defaultproject"
  }
  spec {
    artifact {
      artifact {
        paths {
          name = "file://artifacts/testk8srequiredlabels/k8srequiredlabels.yaml"
        }
      }
      options {
        force = true
      }
      type = "Yaml"
    }
  }
}
