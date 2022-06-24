#Create a secret sealer, sharing across project disabled
resource "rafay_secretsealer" "tfdemosecretsealer1" {
  metadata {
    name    = "tfdemosecretsealer1"
    project = "tfdemoproject1"
  }
  spec {
    type = ""
    sharing {
      enabled = false
    }
    version = ""
  }
}

resource "rafay_secretsealer" "tfdemosecretsealer2" {
  metadata {
    name    = "tfdemosecretsealer2"
    project = "tfdemoproject2"
  }
  spec {
    type = ""
    sharing {
      enabled = true
      projects {
        name = "defaultproject"
      }
    }
    version = ""
  }
}