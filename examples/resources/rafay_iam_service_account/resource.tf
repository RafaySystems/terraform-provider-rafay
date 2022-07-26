resource "rafay_irsa" "basic_irsa" {
  metadata {
    name    = "terraform-test"
  }
  spec {
    cluster_name = "ClusterAgent"
    namespace = "terraform-test"
    permissions_boundary = ""
    role_only = false
    policy_arns = [""]
    //tags =
    policy_document = <<EOF
    {
      "Version": "2012-10-17",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": "ec2:Describe*",
          "Resource": "*"
        },
        {
          "Effect": "Allow",
          "Action": "ec2:AttachVolume",
          "Resource": "*"
        },
        {
          "Effect": "Allow",
          "Action": "ec2:DetachVolume",
          "Resource": "*"
        },
        {
          "Effect": "Allow",
          "Action": ["ec2:*"],
          "Resource": ["*"]
        },
        {
          "Effect": "Allow",
          "Action": ["elasticloadbalancing:*"],
          "Resource": ["*"]
        }
      ]
    }
    EOF
  }
}
/*
data "rafay_irsa" "example" {
  statement {
    sid    = ""
    effect = "Allow"

    principals {
      identifiers = ["lambda.amazonaws.com"]
      type        = "Service"
    }

    actions = ["sts:AssumeRole"]
  }
}*/


