resource "rafay_secret_provider" "tfdemosecretprovider_gitrepo" {
  metadata {
    name    = "test"
    project = "defaultproject"
  }
  spec {
    artifact {
      artifact {
        paths {
          name = "aws-csi1-bp-change/projects/defaultproject/secretproviderclasses/artifacts/two/aws-sample.yaml"
        }
        repository = "github-test"
        revision = "main"
      }
      options {}
      type = "Yaml"
    }
    provider = "AWS"
  }
}
