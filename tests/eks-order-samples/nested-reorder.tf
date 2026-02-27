# Sample: nested list reordering (taints, attach_ids, instance_types)
# Replace provider settings and cloud_provider with real values.

resource "rafay_eks_cluster" "nested_reorder" {
  cluster {
    metadata {
      name    = "sample-nested-reorder"
      project = "default"
    }

    spec {
      cloud_provider = "your-cloud-credential"
    }
  }

  cluster_config {
    metadata {
      name   = "sample-nested-reorder"
      region = "us-west-2"
    }

    managed_nodegroups {
      name          = "taints-ng"
      instance_type = "m5.large"
      asg_suspend_processes = ["Terminate", "Launch", "AZRebalance"]
      taints {
        key    = "b"
        effect = "NoSchedule"
        value  = "two"
      }
      taints {
        key    = "a"
        effect = "NoExecute"
        value  = "one"
      }
    }

    managed_nodegroups {
      name          = "attach-ng"
      instance_type = "m5.large"
      security_groups {
        attach_ids = ["sg-3", "sg-1", "sg-2"]
      }
    }

    managed_nodegroups {
      name           = "instance-ng"
      instance_types = ["m6a.large", "m5a.large", "m7i.large"]
    }

    node_groups {
      name          = "taints-ng"
      instance_type = "m5.large"
      asg_suspend_processes = ["Terminate", "Launch", "AZRebalance"]
      classic_load_balancer_names = ["lb-3", "lb-1", "lb-2"]
      target_group_arns = ["tg-3", "tg-1", "tg-2"]
      taints {
        key    = "b"
        effect = "NoSchedule"
        value  = "two"
      }
      taints {
        key    = "a"
        effect = "NoExecute"
        value  = "one"
      }
    }

    node_groups {
      name          = "attach-ng"
      instance_type = "m5.large"
      security_groups {
        attach_ids = ["sg-3", "sg-1", "sg-2"]
      }
    }

    node_groups {
      name           = "instance-ng"
      instances_distribution {
        instance_types = ["m6a.large", "m5a.large", "m7i.large"]
      }
    }
  }
}
