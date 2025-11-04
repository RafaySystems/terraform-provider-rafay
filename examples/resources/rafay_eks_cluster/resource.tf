resource "rafay_eks_cluster" "ekscluster-basic" {
  cluster {
    kind = "Cluster"
    metadata {
      name    = "bharath-eks-clustertest2"
      project = "defaultproject"
    }
    spec {
      type              = "eks"
      blueprint         = "default"
      blueprint_version = "Latest"
      cloud_provider    = "bharat-eks"
      cni_provider      = "aws-cni"
      proxy_config      = {}
    }
  }
  cluster_config {
    apiversion = "rafay.io/v1alpha5"
    kind       = "ClusterConfig"
    metadata {
      name    = "bharath-eks-clustertest2"
      region  = "us-west-2"
      version = "1.32"
      tags = {
        env   = "dev"
        email = "bharath.reddy@rafay.co"
      }
    }
    addons_config {
      disable_ebs_csi_driver = false
    }
    # secrets_encryption {
    #   encrypt_existing_secrets = true
    #   key_arn = "arn:aws:kms:us-west-2:679196758854:key/11f8e3ba-55f7-4688-82be-96e1ec8d55fb"
    # } 
    iam {
      with_oidc = true
      service_accounts {
        metadata {
          name      = "test-irsa"
          namespace = "yaml1"
        }
        attach_policy = jsonencode({
          "Version" : "2012-10-17",
          "Statement" : [
            {
              "Effect" : "Allow",
              "Action" : "ec2:Describe*",
              "Resource" : "*"
            },
            {
              "Effect" : "Allow",
              "Action" : [
                "ec2:AttachVolume",
              ],
              "Resource" : "*"
            },
            {
              "Effect" : "Allow",
              "Action" : "ec2:DetachVolume",
              "Resource" : "*"
            },
            {
              "Effect" : "Allow",
              "Action" : ["elasticloadbalancing:*"],
              "Resource" : [
                "*"
              ]
            }
          ]
        })
      }
    }


    vpc {
      cidr = "192.168.0.0/16"
      cluster_endpoints {
        private_access = true
        public_access  = false
      }
      nat {
        gateway = "Single"
      }
    }
    managed_nodegroups {
      name = "ng-1"
      # ami = "ami-07a1409f173fe796b"
      ami_family = "AmazonLinux2"
      iam {
        iam_node_group_with_addon_policies {
          image_builder = true
          auto_scaler   = true
        }
      }
      instance_type      = "t3.large"
      desired_capacity   = 2
      min_size           = 1
      max_size           = 2
      max_pods_per_node  = 50
      version            = "1.32"
      volume_size        = 80
      volume_type        = "gp3"
      private_networking = true
      labels = {
        app       = "infra"
        dedicated = "true"
      }
    }
    managed_nodegroups {
      name       = "ng-2"
      ami_family = "AmazonLinux2"
      iam {
        iam_node_group_with_addon_policies {
          image_builder = true
          auto_scaler   = true
        }
      }
      instance_type      = "t3.large"
      desired_capacity   = 2
      min_size           = 1
      max_size           = 2
      max_pods_per_node  = 50
      version            = "1.32"
      volume_size        = 80
      volume_type        = "gp3"
      private_networking = true
      labels = {
        app       = "infra"
        dedicated = "true"
      }
    }
  }
}