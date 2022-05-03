resource "rafay_pipeline" "tfdemopipeline1" {
  metadata {
    name    = "tfdemopipeline1"
    project = "defaultproject"
  }
  spec = {
      "stages" = [
        {
          "config" = {
            "approvers" = [
              {
                "ssoUser" = false
                "userName" = "sougat@rafay.coRepository"
              },
            ]
            "timeout" = "120s"
            "type" = "Email"
          }
          "name" = "stage1"
          "next" = [
            {
              "name" = "stage2"
            },
          ]
          "type" = "Approval"
        },
        {
          "config" = {
            "approvers" = [
              {
                "ssoUser" = false
                "userName" = "benny@rafay.co"
              },
            ]
            "timeout" = "120s"
            "type" = "Email"
          }
          "name" = "stage2"
          "type" = "Approval"
        },
      ]
    }
  }
}
