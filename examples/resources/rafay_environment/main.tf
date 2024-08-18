
resource "rafay_environment" "mvsphani-tf-clust" {
  metadata {
    name    = "mvsphani-tf-clust"
    project = "defaultproject"
  }
  spec {
    template {
      name    = "tf-cred-cluster-et"
      version = "v1"
    }
    agents {
      name = "tbagent"
    }
    variables {
      name       = "credentialType"
      value_type = "text"
      value      = "ClusterProvisioning"
      options {
        sensitive   = false
      }
    }


    variables {
      name       = "credentialProvider"
      value_type = "text"
      value      = "aws"
      options {
        sensitive   = false
      }
    }

     variables {
      name       = "awsCredentialsType"
      value_type = "text"
      value      = "AccessKey"
      options {
        sensitive   = false
      }
    }


    variables {
      name       = "blueprintName"
      value_type = "text"
      value      = "minimal"
      options {
        sensitive   = false
      }
    }
     variables {
      name       = "blueprintVersion"
      value_type = "text"
      value      = "2.8.0"
      options {
        sensitive   = false
      }
    }
    
     variables {
      name       = "cloudCredentials"
      value_type = "text"
      value      = "tf-creds"
      options {
        sensitive   = false
      }
    }
     variables {
      name       = "clusterName"
      value_type = "text"
      value      = "mvsphani-tf-clust"
      options {
        sensitive   = false
      }
    }
     variables {
      name       = "cniName"
      value_type = "text"
      value      = "aws-cni"
      options {
        sensitive   = false
      }
    }
    
     variables {
      name       = "k8sVersion"
      value_type = "text"
      value      = "\"1.29\""
      options {
        sensitive   = false
        required    = true
      }
    }
     variables {
      name       = "privateAccess"
      value_type = "text"
      value      = "true"
      options {
        sensitive   = false
        required    = true
      }
    }
    
     variables {
      name       = "publicAccess"
      value_type = "text"
      value      = "false"
      options {
        sensitive   = false
        required    = true
      }
    }
     variables {
      name       = "region"
      value_type = "text"
      value      = "us-west-2"
      options {
        sensitive   = false
        required    = true
      }
    }
     variables {
      name       = "vpcCidr"
      value_type = "text"
      value      = "192.168.0.0/16"
      options {
        sensitive   = false
        required    = true
      }
    }

     variables {
      name       = "tags"
      value_type = "json"
      value = <<EOF
        {
          "email": "mvsphani@rafay.co",
          "env": "dev"
        }
        EOF
      options {
        sensitive   = false
        required    = true
      }
    }
     variables {
      name       = "addons"
      value_type = "hcl"
      value      = <<EOF
         [
              {
                  "name" = "coredns"
                  "version" = "latest"
              },
              {
                  "name" = "vpc-cni"
                  "version" = "latest"
              },
              {
                  "name" = "kube-proxy"
                  "version" = "latest"
              },
              {
                  "name" = "aws-ebs-csi-driver"
                  "version" = "latest"
              }
          ]
        EOF
      options {
        sensitive   = false
        required    = true
      }
    }
     variables {
      name       = "managedNodeGroups"
      value_type = "hcl"
      value      = <<EOF
        [
              {
                  "amiFamily" = "AmazonLinux2"
                  "desiredCapacity" = 2
                  "instanceTypes" = [
                      "t3.medium"
                  ],
                  "maxSize" = 2
                  "minSize" = 2
                  "name" = "ng-2"
                  "version" = "1.28"
                  "volumeSize" = 80
                  "volumeType" = "gp3"
              }
        ]
        EOF
      options {
        sensitive   = false
        required    = true
      }
    }
     variables {
      name       = "clusterType"
      value_type = "text"
      value      = "aws-eks"
      options {
        sensitive   = false
        required    = true
      }
    }
  }
}