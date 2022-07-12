#Kubernetes agent

resource "rafay_irsa" "basic_irsa" {
  metadata {
    name    = "terraform test"
  }
  spec {
    cluster_name = "ClusterAgent"
    namespace = ""
    permissions_boundary = ""
    role_only = false
    policy_arns = [""]
    //tags =
    policy_document = data.aws_iam_policy_document.example.json
  }
}

data "aws_iam_policy_document" "example" {
  statement {
    sid    = ""
    effect = "Allow"

    principals {
      identifiers = ["lambda.amazonaws.com"]
      type        = "Service"
    }

    actions = ["sts:AssumeRole"]
  }
}


