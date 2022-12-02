
# Username/Password example
resource "rafay_container_registry" "tfdemocontainerregistry" {
  metadata {
    name    = "cr-tf"
    project = "defaultproject"
  }

  spec {
    provider = "Custom"
    credentials {//only part that can be modified 
      password = "password_token"
      username = "sou"
    }
    endpoint = "myregistry.example.com:5000"
  }
}


