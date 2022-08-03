resource "rafay_secret_provider" "tfdemosecretprovider" {
  metadata {
    name    = "test"
    project = "defaultproject"
  }
  spec {
    parameters = {
      "objects": {
        "jmesPath": {
          "objectAlias" : "apiq"
          "path" : "apiq"
        },
        "objectName": "testq",
        "objectType": "secretsmanager"
      }
    }
    provider = "AWS"
  }
}
