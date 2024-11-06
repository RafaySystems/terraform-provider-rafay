resource "rafay_breakglassaccess" "test_user" {
  metadata {
    name = "test@rafay.co"
  }
  spec {
    groups {
      group_expiry {
        expiry     = 7
        name      = "grp-2"
        timezone = "America/Los_Angeles"
      }
      group_expiry {
        expiry     = 8
        name      = "grp-1"
        start_time = "2024-09-20T08:00:00Z"
      }
      user_type = "local"
    }
  }

  lifecycle {
    ignore_changes = [
      spec[0].groups[0].group_expiry[0].start_time
    ]
  }
}
