resource "rafay_cloud_credential" "tfcredential1" {
  name         = "tfcredential1"
  project      = "terraform"
  description  = "description"
  type         = "cluster-provisioning"
  providertype = "AWS"
  awscredtype  = "accesskey"
  accesskey    = "aws-accesskey"
  secretkey    = "aws-secretkey"
}


resource "rafay_cloud_credential" "tfcredential2" {
  name         = "tfcredential2"
  project      = "terraform"
  description  = "description"
  type         = "cluster-provisioning"
  providertype = "AWS"
  awscredtype  = "rolearn"
  rolearn      = "arn:aws:iam::<AWS_ACCOUNT_ID>:role/<role-name>"
  externalid   = "aws-externalid"
}

resource "rafay_cloud_credential" "tfcredential3" {
  name         = "tfcredential3"
  project      = "terraform"
  description  = "descriptions"
  type         = "cluster-provisioning"
  providertype = "GCP"
  credfile     = "/Users/user1/gcpcredentials.json"
}

resource "rafay_cloud_credential" "tfcredential4" {
  name           = "tfcredential4"
  project        = "terraform"
  description    = "description"
  type           = "cluster-provisioning"
  providertype   = "AZURE"
  clientid       = "azure-client-id"
  clientsecret   = "azure-clientsecret"
  subscriptionid = "azure-subscriptionid"
  tenantid       = "azure-tenantid"
}

resource "rafay_cloud_credential" "tfcredential5" {
  name           = "tfcredential5"
  project        = "terraform"
  type           = "data-backup"
  providertype   = "AWS"
  awscredtype    = "accesskey"
  accesskey      = "aws-accesskey"
  secretkey      = "aws-secretkey"
}