resource "rafay_pipeline" "bharathtest1" {
  metadata {
    name    = "bharathtest1"
    project = "kbr-test"
  }
  spec = {
      "stages" = [
        {
          "config" = {
            "approvers" = [
              {
                "ssoUser" = false
                "userName" = "bharath.reddy@rafay.co"
              },
            ]
            "timeout" = "120s"
            "type" = "Email"
          }
          "name" = "stage1"
          # "next" = [
          #   {
          #     "name" = "stage2"
          #   },
          # ]
          "type" = "Approval"
        },
        # {
        #   "config" = {
        #     "approvers" = [
        #       {
        #         "ssoUser" = false
        #         "userName" = "varun.chandrashekar@rafay.co"
        #       },
        #     ]
        #     "timeout" = "120s"
        #     "type" = "Email"
        #   }
        #   "name" = "stage2"
        #   "type" = "Approval"
        # },
      ]
    }
  }

