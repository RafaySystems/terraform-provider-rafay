# YAML Upload Example
resource "rafay_addon" "tfdemoaddon1" {
  metadata {
    name    = "tfdemoaddon1"
    project = "terraform_project"
  }
  spec {
    namespace = "tfdemonamespace"
    version   = "v1.0"
    artifact {
      type = "Yaml"
      artifact {
        paths {
          name = "file://artifacts/tfdemoaddon1/busybox.yaml"
        }

      }
    }
    sharing {
      enabled = false
    }
  }
}


# Helm Chart Upload Example

resource "rafay_addon" "tfdemoaddon2" {
  metadata {
    name    = "tfdemoaddon2"
    project = "terraform_project"
  }
  spec {
    namespace = "tfdemonamespace"
    version   = "v1.0"
    artifact {
      type = "Helm"
      artifact {
        chart_path {
          name = "file://artifacts/tfdemoaddon4/apache-9.0.9.tgz"
        }
      }
      options {
          max_history = 10
          timeout = "5m0s"
      }
    }
    sharing {
      enabled = false
    }
  }
}