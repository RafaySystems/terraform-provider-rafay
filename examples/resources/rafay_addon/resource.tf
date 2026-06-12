# # YAML Upload Example
# resource "rafay_addon" "tfdemoaddon1" {
#   metadata {
#     name    = "tfdemoaddon1"
#     project = "terraform"
#   }
#   spec {
#     namespace = "tfdemonamespace1"
#     version   = "v1.0"
#     artifact {
#       type = "Yaml"
#       artifact {
#         paths {
#           name = "file://artifacts/tfdemoaddon1/busybox.yaml"
#         }

#       }
#     }
#     sharing {
#       enabled = false
#     }
#   }
# }


# # Helm Chart Upload Example
# resource "rafay_addon" "tfdemoaddon4" {
#   metadata {
#     name    = "tfdemoaddon4"
#     project = "terraform"
#   }
#   spec {
#     namespace = "tfdemonamespace1"
#     version   = "v1.0"
#     artifact {
#       type = "Helm"
#       artifact {
#         chart_path {
#           name = "file://artifacts/tfdemoaddon4/apache-9.0.9.tgz"
#         }
#       }
#       options {
#         max_history = 10
#         timeout     = "5m0s"
#       }
#     }
#     sharing {
#       enabled = true
#       projects {
#         name = "project1"
#       }
#       projects {
#         name = "project2"
#       }
#     }
#   }
# }

# # Catalog Example
# resource "rafay_addon" "tfdemoaddon2" {
#   metadata {
#     name    = "tfdemoaddon2"
#     project = "terraform"
#   }
#   spec {
#     namespace = "tfdemonamespace1"
#     version   = "v1.0"
#     artifact {
#       type = "Helm"
#       artifact {
#         catalog       = "catalogName"
#         chart_name    = "chartName"
#         chart_version = "chartVersion"
#         values_paths {
#           name = "file://relative/path/to/some/chart/values.yaml"
#         }
#       }
#       options {
#         max_history = 10
#         timeout     = "5m0s"
#       }
#     }
#   }
# }


# # Web YAML
# resource "rafay_addon" "tfdemoaddon5" {
#   metadata {
#     name    = "tfdemoaddon5"
#     project = "terraform"
#   }
#   spec {
#     namespace     = "tfdemonamespace1"
#     version       = "v1.0"
#     version_state = "active"
#     artifact {
#       type = "Yaml"
#       artifact {
#         url = ["https://raw.githubusercontent.com/kubernetes/website/main/content/en/examples/application/nginx-app.yaml"]
#       }
#     }
#     sharing {
#       enabled = false
#     }
#   }
# }

# resource "rafay_addon" "tfdemoaddon6" {
#   metadata {
#     name    = "tfdemoaddon6"
#     project = "terraform"
#   }
#   spec {
#     namespace = "default"
#     version   = "production"
#     artifact {
#       type = "Kustomize"
#       artifact {
#         path = "production"
#         file {
#           name = "file://artifacts/tfdemoaddon6/archive.tar.gz"
#         }
#       }
#     }
#     sharing {
#       enabled = false
#     }
#   }
# }

# resource "rafay_addon" "tfdemoaddon7" {
#   metadata {
#     name    = "tfdemoaddon7"
#     project = "terraform"
#   }
#   spec {
#     namespace = "default"
#     version   = "prod"
#     artifact {
#       type = "Kustomize"
#       artifact {
#         repository = "kustomize-repo"
#         revision   = "master"
#         directory  = "examples/multibases"
#         path       = "production"
#       }
#     }
#     sharing {
#       enabled = false
#     }
#   }
# }


# apiVersion: apps.k8smgmt.io/v3
# kind: Workload
# metadata:
#   name: workload1 
#   project: defaultproject
# spec:
#   artifact:
#     artifact:
#       chartPath:
#         name: file://my-pod-chart-0.1.0.tgz
#     options: {}
#     type: Helm4
#   namespace: ns4
#   placement:
#     selector: rafay.dev/clusterName=kratos-rishabh-1



# Helm Chart Upload Example
# resource "rafay_addon" "helm-addon" {
#   metadata {
#     name    = "helm-addon"
#     project = "defaultproject"
#   }
#   spec {
#     namespace = "ns4"
#     version   = "v1.0"
#     artifact {
#       type = "Helm4"
#       artifact {
#         chart_path {
#           name = "file://artifacts/v3/my-pod-chart-0.1.0.tgz"
#         }
#       }
#     }
#   }
# }

# Helm4 — Chart upload with options
# resource "rafay_addon" "tfdemoaddon-helm4-upload" {
#   metadata {
#     name    = "tfdemoaddon-helm4-upload"
#     project = "defaultproject"
#   }
#   spec {
#     namespace = "ns4"
#     version   = "v2.0"
#     artifact {
#       type = "Helm4"
#       artifact {
#         chart_path {
#           name = "file://artifacts/v3/my-pod-chart-0.1.0.tgz"
#         }
#       }
#       options {
#         max_history = 10
#         timeout     = "5m0s"
#         wait_for_jobs = true
#         wait_strategy = "watcher"
#         server_side_apply = "auto"
#         force_replace = true
#         force_conflicts = true
#         take_ownership = true
#         replace = true
#         max_history = 10
#         rollback_on_failure = true
#         cleanup_on_fail = true
#         reset_values = true
#       }
#     }
#   }
# }
