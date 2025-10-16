resource "rafay_eks_cluster" "iam_custom" {
  cluster {
    metadata {
      name    = "${var.cluster_name}"
      project = var.rafay_project
      labels  = { security = "custom-iam" }
    }

    spec {
      cloud_provider = "aws"
      type           = "aws-eks"
    }
  }

  cluster_config {
    metadata2 {
      name   = "${var.cluster_name}"
      region = var.aws_region
      version = var.kubernetes_version
    }

    managed_nodegroups_map = {
      "iam-ng" = {
        instance_type = var.node_instance_type
        desired_capacity = 1
        iam5 = {
          instance_profile_arn = "arn:aws:iam::123456789012:instance-profile/custom-profile"
          instance_role_arn    = "arn:aws:iam::123456789012:role/custom-node-role"
          attach_policy_arns = [
            "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"
          ]
          iam_node_group_with_addon_policies5 = {
            auto_scaler = true
            image_builder = true
          }
        }
      }
    }
  }
}