resource "rafay_resource_template" "tfcredrt" {
  metadata {
    name    = "tf-cred-rt"
    project = "defaultproject"
  }
  spec {
    version  = "v2"
    provider = "system"
    provider_options {
      system {
        kind = "credential"
      }
    }
    sharing {
      enabled = true
      projects {
       name = "eaas"
      }
    }
    hooks {
      provider {
        system {
          deploy {
            apply {
              before {
                name = "beforeapplyrt"
                type = "http"
                options {
                  http {
                    endpoint = "https://jsonplaceholder.typicode.com/todos/1"
                    method   = "POST"
                    headers = {
                      X-TOKEN = "token"
                    }
                    body              = "request-body"
                    success_condition = "200OK"
                  }
                }
              }
              after {
                name = "afterapplyrt"
                type = "http"
                options {
                  http {
                    endpoint = "https://jsonplaceholder.typicode.com/todos/1"
                    method   = "POST"
                    headers = {
                      X-TOKEN = "token"
                    }
                    body              = "request-body"
                    success_condition = "200OK"
                  }
                }
              }
            }
          }
          destroy {
            destroy {
              before {
                name = "beforedestroyrt"
                type = "http"
                options {
                  http {
                    endpoint = "https://jsonplaceholder.typicode.com/todos/1"
                    method   = "POST"
                    headers = {
                      X-TOKEN = "token"
                    }
                    body              = "request-body"
                    success_condition = "200OK"
                  }
                }
              }
              after {
                name = "afterdestroyrt"
                type = "http"
                options {
                  http {
                    endpoint = "https://jsonplaceholder.typicode.com/todos/1"
                    method   = "POST"
                    headers = {
                      X-TOKEN = "token"
                    }
                    body              = "request-body"
                    success_condition = "200OK"
                  }
                }
              }
            }
          }
        }
      }
    }
    variables {
      name       = "cloudCredentials"
      value_type = "text"
      options {
        description = "cloud credentials name"
        sensitive   = false
        required    = true
      }
    }
    variables {
      name       = "credentialType"
      value_type = "text"
      options {
        description = "credentials type"
        sensitive   = false
        required    = true
      }
    }

    variables {
      name       = "credentialProvider"
      value_type = "text"
      options {
        description = "credentials provider eg: aws"
        sensitive   = false
        required    = true
      }
    }

    variables {
      name       = "awsCredentialsType"
      value_type = "text"
      options {
        description = "aws credential type eg: AccessKey"
        sensitive   = false
        required    = true
      }
    }

    variables {
      name       = "awsAccessId"
      value_type = "text"
      options {
        description = "aws access id"
        sensitive   = false
        required    = true
      }
    }

    variables {
      name       = "awsSecretKey"
      value_type = "text"
      options {
        description = "aws secret key"
        sensitive   = true
        required    = true
      }
    }
    agents {
      name = "tbagent"
    }
  }
}


resource "rafay_resource_template" "tfclusterrt" {
  metadata {
    name    = "tf-cluster-rt"
    project = "defaultproject"
  }
  spec {
    version  = "v1"
    provider = "system"
    provider_options {
      system {
        kind = "cluster"
      }
    }
    variables {
      name       = "cloudCredentials"
      value_type = "text"
      options {
        description = "cloud credentials name"
        sensitive   = false
        required    = true
      }
    }
    variables {
      name       = "clusterName"
      value_type = "text"
      options {
        description = "cluster name"
        sensitive   = false
        required    = true
      }
    }

    variables {
      name       = "blueprintName"
      value_type = "text"
      options {
        description = "blueprint name"
        sensitive   = false
        required    = true
      }
    }

    variables {
      name       = "blueprintVersion"
      value_type = "text"
      options {
        description = "blueprint version"
        sensitive   = false
        required    = true
      }
    }

    variables {
      name       = "region"
      value_type = "text"
      options {
        description = "aws region"
        sensitive   = false
        required    = true
      }
    }

    variables {
      name       = "k8sVersion"
      value_type = "text"
      options {
        description = "k8s version"
        sensitive   = false
        required    = true
      }
    }

    variables {
      name       = "tags"
      value_type = "json"
      options {
        description = "cluster tags"
        sensitive   = false
        required    = true
      }
    }

    variables {
      name       = "addons"
      value_type = "json"
      options {
        description = "cluster addons"
        sensitive   = false
        required    = true
      }
    }

    variables {
      name       = "managedNodeGroups"
      value_type = "json"
      options {
        description = "cluster managed nodegroups"
        sensitive   = false
        required    = true
      }
    }

    variables {
      name       = "cniName"
      value_type = "text"
      options {
        description = "cni name"
        sensitive   = false
        required    = true
      }
    }

    variables {
      name       = "vpcCidr"
      value_type = "text"
      options {
        description = "vpc cidr block"
        sensitive   = false
        required    = true
      }
    }

     variables {
      name       = "privateAccess"
      value_type = "text"
      options {
        description = "private access"
        sensitive   = false
        required    = true
      }
    }

    variables {
      name       = "publicAccess"
      value_type = "text"
      options {
        description = "public access"
        sensitive   = false
        required    = true
      }
    }

    variables {
      name       = "clusterType"
      value_type = "text"
      options {
        description = "cluster type eg: aws-eks"
        sensitive   = false
        required    = true
      }
    }




    agents {
      name = "tbagent"
    }
  }
}