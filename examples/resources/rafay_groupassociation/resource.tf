resource "rafay_groupassociation" "groupassociation" {
  group      = "dev1"
  project    = "terraform"
  roles      = ["PROJECT_READ_ONLY"]
}

resource "rafay_groupassociation" "groupassociation1" {
  group      = "dev2"
  project    = "terraform"
  namespaces = ["ns1", "ns2"]
  roles      = ["NAMESPACE_ADMIN"]
  add_users  = ["user1@org"]
}

resource "rafay_groupassociation" "groupassociation2" {
  group      = "dev2"
  project    = "terraform"
  namespaces = ["ns1", "ns2"]
  custom_roles      = ["test-custom-role"]
  add_users  = ["user1@org"]
}