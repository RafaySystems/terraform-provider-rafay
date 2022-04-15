resource "rafay_secretsealer" "tfdemosecretsealer1" {
  metadata {
    name    = "tfdemosecretsealer1"
    project = "tfdemoproject1"
  }
  spec {
    type = ""
    sharing {
      enabled = true
      projects {
        name = "demo"
      }
    }
    version = ""
  }
}