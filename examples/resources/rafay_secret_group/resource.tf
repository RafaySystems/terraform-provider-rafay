resource "rafay_secret_group" "tfdemosecretgroup1" {
  metadata {
    name    = "tfdemosecretgroup1"
    project = "terraform"
  }
  spec {
    secret {
        name = "/Users/chaitanyaangadala/Downloads/v3exampleYaml/testsecretfile.yaml"
    }
  }
}