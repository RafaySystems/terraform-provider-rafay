resource "rafay_groupassociation" "groupassociation" {
  group      = "testgroup-benny1"
  project    = "benny-test1"
  roles      = ["NAMESPACE_READ_ONLY"]
  namespaces = ["kube-system"]
}
#avaliable roles: ["ADMIN", "PROJECT_ADMIN", "PROJECT_READ_ONLY", "INFRA_ADMIN", "INFRA_READ_ONLY", "NAMESPACE_READ_ONLY", "NAMESPACE_ADMIN"]
#avaliable namespaces as for your configuration, only provide when selected roles are namespace options