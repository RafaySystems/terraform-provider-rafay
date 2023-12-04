resource "rafay_ztkapolicy" "demo_rafay_ztkapolicy" {
  metadata {
    name = "test-ztkapolicy-terraform-1"
  }
  spec {
    ztka_rule_list {
      name    = "ztka-rule1"
      version = "v1"
    }
    ztka_rule_list {
      name    = "ztka-rule2"
      version = "v2"
    }
    version = "v1"
  }
}