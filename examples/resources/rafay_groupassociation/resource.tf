resource "rafay_group" "group1" {
  name        = "terraform-group1"
  description = "dev1-description"
}
resource "rafay_group" "group2" {
  name        = "terraform-group2"
  description = "dev1-description"
}
resource "rafay_group" "group3" {
  name        = "terraform-group3"
  description = "dev1-description"
}
resource "rafay_group" "group4" {
  name        = "terraform-group4"
  description = "dev1-description"
}


resource "rafay_groupassociation" "groupassociation" {
  group      = rafay_group.group1.name
  project      = "a-ankit1"
  roles      = ["PROJECT_READ_ONLY"]
}

resource "rafay_groupassociation" "groupassociation1" {
  group      = rafay_group.group2.name
  project      = "a-ankit2"
  namespaces = ["a2ns1", "a2ns2", "a2ns3"]
  roles      = ["NAMESPACE_ADMIN"]
  add_users  = ["ankit+relay@rafay.co"]
}
resource "rafay_groupassociation" "groupassociation2" {
  group      = rafay_group.group3.name
  project      = "ankit-project"
  namespaces = ["ns-01", "ns-02", "ns-03"]
  custom_roles      = ["ztka-custom-rule"]
  add_users  = ["ankit+relay@rafay.co"]
}

resource "rafay_groupassociation" "groupassociation3" {
  group      = rafay_group.group4.name
  project      = "a-ankit3"
  namespaces = ["a3ns1", "a3ns2", "a3ns3"]
  roles      = ["NAMESPACE_ADMIN"]
  custom_roles      = ["ankit-cr-role"]
  add_users  = ["ankit+ws@rafay.co"]
}

resource "rafay_groupassociation" "groupassociation3" {
  group      = "dev1"
  roles      = ["ADMIN"]
}