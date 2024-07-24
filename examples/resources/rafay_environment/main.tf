
resource "rafay_environment" "mvsphani-tf-clust" {
  metadata {
    name    = "mvsphani-tf-clust"
    project = "defaultproject"
  }
  spec {
    template {
      name    = "clusteret"
      version = "v1"
    }
    variables {
      name       = "blueprintName"
      value_type = "text"
      value      = "minimal"
      options {
        sensitive   = false
        required    = true
      }
    }
     variables {
      name       = "blueprintVersion"
      value_type = "text"
      value      = "Latest"
      options {
        sensitive   = false
        required    = true
      }
    }
    
     variables {
      name       = "cloudCredentials"
      value_type = "text"
      value      = "ekscreds"
      options {
        sensitive   = false
        required    = true
      }
    }
     variables {
      name       = "clusterName"
      value_type = "text"
      value      = "mvsphani-tf-clust"
      options {
        sensitive   = false
        required    = true
      }
    }
     variables {
      name       = "cniName"
      value_type = "text"
      value      = "aws-cni"
      options {
        sensitive   = false
        required    = true
      }
    }
     variables {
      name       = "customCniCidr"
      value_type = "text"
      value      = "100.0.0.0/16"
      options {
        sensitive   = false
        required    = true
      }
    }
     variables {
      name       = "k8sVersion"
      value_type = "text"
      value      = "\"1.28\""
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
      name       = "projectName"
      value_type = "text"
      value      = "defaultproject"
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
      value_type = "json"
      value      = <<EOF
         [
              {
                  "name": "coredns",
                  "version": "v1.10.1-eksbuild.4"
              },
              {
                  "name": "vpc-cni",
                  "version": "v1.15.1-eksbuild.1"
              },
              {
                  "name": "kube-proxy",
                  "version": "v1.28.2-eksbuild.2"
              },
              {
                  "name": "aws-ebs-csi-driver",
                  "version": "latest"
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
      value_type = "json"
      value      = <<EOF
        [
              {
                  "amiFamily": "AmazonLinux2",
                  "desiredCapacity": 2,
                  "instanceTypes": [
                      "t3.medium"
                  ],
                  "maxSize": 2,
                  "minSize": 2,
                  "name": "ng-1",
                  "version": "1.28",
                  "volumeSize": 80,
                  "volumeType": "gp3"
              }
        ]
        EOF
      options {
        sensitive   = false
        required    = true
      }
    }
  }
}