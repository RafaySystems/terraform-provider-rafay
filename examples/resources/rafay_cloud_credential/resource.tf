resource "rafay_cloud_credential" "credential" {
  name        = "dev1"
  description = "dev1-description"
  roleARN     = "xxxxxxx"
  credType    = "cluster-provisioning"
  externalId  = "yyyyyy"
}
