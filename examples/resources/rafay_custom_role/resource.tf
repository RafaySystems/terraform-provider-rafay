resource "rafay_customrole" "demo_rafay_customrole" {
  metadata {
    name = "test-customrole-terraform-1"
  }
  spec {
    abac_policy_list {
      name    = "abac-policy1"
      version = "v1"
    }
    abac_policy_list {
      name    = "abac-policy2"
      version = "v2"
    }
    ztka_policy_list {
      name    = "ztka-policy1"
      version = "v1"
    }
    ztka_policy_list {
      name    = "ztka-policy2"
      version = "v2"
    }
    base_role = "NAMESPACE_ADMIN"
  }
}