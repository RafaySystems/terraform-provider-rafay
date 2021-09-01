resource "rafay_cloud_credential" "credential" {
  name         = "testinggcp2"
  projectname  = "dev3"
  description  = "description"
  type         = "cluster-provisioning"
  providertype = "GCP"
  credtype     = "access"
  rolearn      = ""
  externalid   = ""
  accesskey    = ""
  secretkey    = ""
  credfile     = "/Users/krishna/Downloads/Neridiosys-shreekrishna@rafay.co.json"
}
