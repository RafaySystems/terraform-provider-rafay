resource "rafay_secret_provider" "tfdemosecretprovider" {
  metadata {
    name    = "secret"
    project = "terraform"
  }
  spec {
    artifact {
      artifact {
        paths {
          name = "projects/terraform/secretproviderclasses/artifacts/aws-sample.yaml"
        }
        repository = "github-test"
        revision   = "main"
      }
      options {}
      type = "Yaml"
    }
    provider = "AWS"
  }
}
