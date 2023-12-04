# Basic project example
resource "rafay_chargeback_share" "tfdemochargebackshare" {
  metadata {
    name = "chargebackshare"
  }
  spec {
    share_unallocated_cost = true
    share_type = "tenancy"
  }
}
