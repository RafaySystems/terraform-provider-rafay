resource "rafay_cloud_credential" "tfcredential1" {
  name         = "tfcredential1"
  projectname  = "upgrade"
  description  = "description"
  type         = "cluster-provisioning"
  providertype = "AWS"
  awscredtype  = "accesskey"
  accesskey    = "fgbshhrgeineatyssde"
  secretkey    = "abcdfergsfeddadsf"
}


resource "rafay_cloud_credential" "tfcredential2" {
  name         = "tfcredential2"
  projectname  = "upgrade"
  description  = "description"
  type         = "cluster-provisioning"
  providertype = "AWS"
  awscredtype  = "rolearn"
  rolearn      = "tesgtarf"
  externalid   = "reinesg"
}

resource "rafay_cloud_credential" "tfcredential3" {
  name         = "tfcredential3"
  projectname  = "upgrade"
  description  = "descriptions"
  type         = "cluster-provisioning"
  providertype = "GCP"
  credfile     = "/Users/stephanbenny/Downloads/benny-rctl-test-10-848884a20733.json"
}

resource "rafay_cloud_credential" "tfcredential4" {
  name           = "tfcredential4"
  projectname    = "upgrade"
  description    = "description"
  type           = "cluster-provisioning"
  providertype   = "AZURE"
  clientid       = "azure-client-id"
  clientsecret   = "sbgffufwefnwfiefw"
  subscriptionid = "jsdgkdf"
  tenantid       = "dgmwkfj"
}