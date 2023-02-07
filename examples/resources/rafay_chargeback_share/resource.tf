# Chargeback share example
resource "rafay_chargeback_share" "tfdemochargebackshare" {
  metadata {
    name = "chargebackshare"
  }
  spec {
    share_unallocated = true
    share_type = "equal"
    share_common = true
    inclusions {
      namespace = "common-namespace-1"
    }
    inclusions {
      namespace = "common-namespace-2"
    }
  }
}
