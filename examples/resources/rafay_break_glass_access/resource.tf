resource "rafay_breakglassaccess" "test_user" {
  metadata {
    name = "test@rafay.co"
  }
  spec {
    groups {
      group_expiry {
        expiry     = 7
        name      = "grp3"
      }
      group_expiry {
        expiry     = 8
        name      = "grp1"
        start_time = "2024-09-20T08:00:00Z"
      }
      user_type = "local"
    }
  }
}
