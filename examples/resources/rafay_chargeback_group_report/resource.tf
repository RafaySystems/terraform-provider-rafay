# Example chargeback group report resource
resource "rafay_chargeback_group_report" "tfdemochargebackgroupreport1" {
  metadata {
    name = "tfdemosummarycbgroup1"
  }
  spec {
    group_name = "tfdemosummarycbgroup1"
    start_date {
      seconds = 1669011971
    }
    end_date {
      seconds = 1669616771
    }
  }
}
