resource "rafay_static_resource" "static-resource" {
  metadata {
    name    = var.name
    project = var.project
  }
  spec {
    variables {
      name       = "my-variable"
      value_type = "text"
      value      = "my-value"
      options {
        description = "this is a dummy variable"
        sensitive   = false
        required    = true
      }
    }
    variables {
      name       = "my-variable-two"
      value_type = "text"
      value      = "my-value-two"
      options {
        description = "this is another dummy variable"
        sensitive   = true
        required    = true
        override {
          type              = "restricted"
          restricted_values = ["value1", "value2"]
        }
      }
    }
  }
}