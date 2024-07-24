resource "rafay_environment_template" "aws-et" {
  metadata {
    name    = "bptftestenvt"
    project = "defaultproject"
  }
  spec {
    version = "v2"
    resources {
      type = "dynamic"
      kind = "resourcetemplate"
      name = "bptftestrt"
      resource_options {
        version   = "v1"
      }
    }
  }
}