resource "rafay_cloud_credential" "credential" {
  name         = "testinggcp2"
  projectname  = "dev3"
  description  = "description"
  providertype = "GCP"
  rolearn      = "xxxxxxx"
  credtype     = "access"
  externalid   = "yyyyyy"
  type         = "cluster-provisioning"
  accesskey    = "12334234"
  secretkey    = "sdfkls"
  credfile     = "/Users/krishna/Downloads/Neridiosys-shreekrishna@rafay.co.json"
}
