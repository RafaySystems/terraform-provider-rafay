# Example chargeback group report resource
resource "rafay_chargeback_group_report" "tfdemochargebackgroupreport1" {
  metadata {
    name = "tfdemodetailedcbgroup1"
  }
  spec {
    group_name = "tfdemodetailedcbgroup1"
    start_date {
      seconds = 1669679998
    }
    end_date {
      seconds = 1669680000
    }
  }
}
