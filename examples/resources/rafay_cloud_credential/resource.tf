resource "rafay_cloud_credential" "credential" {
  name        = "testingaws2"
  projectname = "dev3"
  description = "description"
  rolearn     = "xxxxxxx"
  credtype    = 1
  externalid  = "yyyyyy"
}
