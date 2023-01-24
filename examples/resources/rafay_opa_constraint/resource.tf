#Basic example for opa constraint
resource "rafay_opa_constraint" "tfdemoopaconstraint1" {
  metadata {
    name    = "tfdemoopaconstraint1"
    project     = "tfdemoproject1"
    labels = {
      "rafay.dev/opa" = "constraint"
    }
  }
  spec {
    artifact {
      type = "Yaml"
      artifact {
        paths {
          name = "file://artifacts/request-limit-ratio/request-limit-ratio.yaml"
        }
      }
      options {
      	force =  true
    	disable_open_api_validation = true
      }
    }
    template_name = "k8srequiredlabels"
    version = "v1"
    published =  true
  }
}