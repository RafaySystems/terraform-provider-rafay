# rafay_workload_cd_operator (Resource)

Helm Workload Deployer

## Table of Contents
- [Overview](#overview)
- [Usage](#usage)
  - [Features](#workload-deployer-features)
  - [Quick Start - Deploying Applications from a Repository](#quick-start---deploying-applications-from-a-repository)
  - [Helm Chart Sourcing Options](#helm-chart-sourcing-options)
  - [Repository Folder Structure](#repository-folder-structure)
    - [Base Path](#base-path)
    - [Path Match Pattern Examples](#path-match-pattern-and-selecting-helm-charts-for-deployment)
      - [Example1](#example1---chart-from-a-common-folder)
	  - [Example2](#example2---chart-from-the-project-folder)
	  - [Example3](#example3---chart-from-a-common-folder-for-one-project-and-from-the-project-folder-for-another-project)
    - [Chart Selection for Deployment](#chart-and-value-selection-for-deployment)
  - [Managing Chart From External Sources](#managing-chart-from-external-sources)
    - [Chart From a Helm Repo](#chart-from-a-helm-repo)
	- [Chart From a Git Repo](#chart-from-a-git-repo)
- [Example - Deploying Applications from a Repository](#example---deploying-applications-from-a-repository)
- [Variables](#variables)
  - [metadata](#metadata)
  - [spec](#spec)
  - [workload](#workload-object)
  - [credentials](#credentials-object)
  - [status](#status)
- [Delete](#delete)



## Overview
This terraform resource facilitates the deployment of Helm workloads (applications) from a Git repository. It streamlines the process of managing and deploying Helm charts directly from version-controlled repositories, providing a seamless GitOps integration for your workloads.

## Usage

The Workload Deployer resource facilitates the cloning of the provided repository and orchestrates the deployment of Helm charts to designated clusters. 

### Workload Deployer Features

- **Cloning Repository:**
  The deployer initiates the process by cloning the specified repository.

- **Rafay Workloads Creation:**
  It then creates Rafay Workloads to deploy Helm charts to the designated clusters.

- **Flexible Configuration:**
  Supports flexible combinations of `project`, `namespace`, and `workload` folder structures.

- **Version Control:**
  Allows the deployment of different chart and values for various workloads.

- **Workload Republishing:**
  Facilitates the republishing of workloads when the repository undergoes updates.

- **Workload Deletion and Unpublishing:**
  - Compares existing workloads with the repository.
  - Deletes or unpublishes workloads from the project when corresponding folders or values are removed from the repository.


### Quick Start - Deploying Applications from a Repository

Please refer to the [test-tfcd](https://github.com/stephan-rafay/test-tfcd) repository for detailed information and configurations.

  
### Helm Chart Sourcing Options

  The Helm chart can be obtained from the following sources:
  
  - **Folder:** Located within the Git repo.
  - **Helm Repo:** Fetched from a Helm repository.
  - **Another Git Repo:** Pulled from a different Git repository.
  
### Repository Folder Structure
The repository folder structure is designed to organize Helm charts in a way that reflects the project name, namespace, and workload name. The folder structure is flexible and can be customized to suit your specific requirements using `path_match_pattern`.

Note:  `project`, `namespace` and `workload` are mandatory.

Examples:

- `/:project/:namespace/:workload` <br>
- `/somefolder/:project/:namespace/:workload` <br>
- `/somefolder/:project/:namespace/:workload/anotherfolder` <br>
- `/somefolder/production/:project/:namespace/:workload` <br>
- `/:workload/:namespace/:project` etc. <br>



#### Base Path
The `base_path` is the common path for the chart and values.

#### Path Match Pattern and Selecting Helm Charts for Deployment 

##### Example1 - Chart from a common folder 
If the folder structure is like below, then the `path_match_pattern` can be specified in the following format:<br> `/:project/:workload/:namespace` <br>
```
- common
  - chart-1.0.1.tgz
- project1
  - workload1
    - namespace1
	  - values.yaml
- project2
  - workload1
    - namespace1
	  - values.yaml
```

Terraform Configuration
```hcl
workload {
  name = "workload1"
  helm_chart_version = "1.0.1"
  helm_chart_name = "chart"
  path_match_pattern = "/:project/:workload/:namespace"
  base_path = "common"
  delete_action = "delete"
  placement_labels = {
    "workload1" = "enabled"
  }
}
```


In this scenario, two workloads will be deployed in respective projects using the same chart file located in the `common` folder. <br>
1. The first workload will use the chart file from the `common` folder and values from the `project1/workload1/namespace1` folder. <br>
2. The second workload will use the chart file from the `common` folder and values from the `project2/workload1/namespace1` folder. <br>


##### Example2 - Chart from the project folder
If the folder structure is like below, then the `path_match_pattern` can be specified in the following format:<br> `/:project/:workload/:namespace` <br>
```
- project1
  - workload1
    - namespace1
      - chart-1.0.1.tgz
	  - values.yaml
- project2
  - workload1
    - namespace1
      - chart-1.0.2.tgz
	  - values.yaml
```

Terraform Configuration
```hcl
workload {
  name = "workload1"
  helm_chart_version = "1.0.1"
  helm_chart_name = "chart"
  path_match_pattern = "/:project/:workload/:namespace"
  delete_action = "delete"
  placement_labels = {
    "workload1" = "enabled"
  }
}
```

In this scenario, two workloads will be deployed in respective projects. <br>
1. The workload name `workload1` in `project1` will use the chart file `chart-1.0.1.tgz` and value from the `project1/workload1/namespace1` folder. <br>
2. The workload name `workload1` in `project2` will use the chart file `chart-1.0.2.tgz` and value from the `project2/workload1/namespace1` folder. <br>


#### Example3 - Chart from a common folder for one project and from the project folder for another project
If the folder structure is like below, then the `path_match_pattern` can be specified in the following format:<br> `/production/:project/:workload/:namespace` <br>
```
- production
  - common
    - chart-1.0.1.tgz
  - project1
    - workload1
      - namespace1
	    - values.yaml
  - project2
    - workload1
      - namespace1
        - chart-1.0.2.tgz
	    - values.yaml
```

Terraform Configuration
```hcl
workload {
  name = "workload1"
  helm_chart_version = "1.0.1"
  helm_chart_name = "chart"
  path_match_pattern = "/:project/:workload/:namespace"
  base_path = "production/common"
  delete_action = "delete"
  placement_labels = {
    "workload1" = "enabled"
  }
}
```

In this scenario, two workloads will be deployed in respective projects. <br>
1. The workload name `workload1` in `project1` will use the chart file `chart-1.0.1.tgz` from `production/common` folder and values from the `production/project1/workload1/namespace1` folder. <br>
2. The workload name `workload1` in `project2` will use the chart file `chart-1.0.2.tgz` and value from the `production/project2/workload1/namespace1` folder. <br>

#### Chart and Value Selection for Deployment

If the chart is found in the respective folder, it takes precedence for deployment over the specified `base_path`. This ensures that the deployment process prioritizes the chart present in the workload folder, providing flexibility to deploy different version of the applicatin.

If the chart or value file is missing for a specific folder, the entire folder will be skipped during the deployment process. This ensures that the deployment focuses only on folders with the necessary files, avoiding any errors caused by missing components.

If more than one value file is present in the folder, the deployment will use all the files with extenstions `.yaml` or `.yml` present in the folder.

If `include_base_values` is set to `true`, the values from the `base_path` will be merged with the values from the workload folder. 

### Managing Chart From External Sources

Obtaining Helm charts from external sources involves two primary options:

1. **Helm Repository (Helm Repo):** Fetching charts from a centralized Helm repository, allowing for versioning, easy updates, and broader accessibility.

2. **Different Git Repository (Git Repo):** Pulling charts from a separate Git repository, providing flexibility for maintaining charts independently from the main project repository.

These external sources enhance scalability and modularity in managing Helm charts for various deployment scenarios.

#### Chart From a Helm Repo

##### Steps for Helm Repository Integration

Follow these steps to integrate Helm repositories into your workflow:

1. **Create a Helm Repository:** Set up a Helm repository on the Rafay platform. [Learn more](https://docs.rafay.co/integrations/repositories/overview/)

2. **Share the Repository:** Share the Helm repository with your projects.

3. **Utilize Helm Charts:** Access and deploy charts directly from the Helm repository for streamlined application deployment.

Terraform Configuration
```hcl
workload {
  name = "echoserver"
  chart_helm_repo_name = "echo-server" # Rafay Helm Repository Name
  helm_chart_version = "0.5.0"
  helm_chart_name = "echo-server"
  path_match_pattern = "/:project/:workload/:namespace"
  base_path = "echoserver-common"
  delete_action = "delete"
  placement_labels = {
    "echoserver" = "enabled"
  }
}
```


#### Chart From a Git Repo

##### Steps for Git Repository Integration

Follow these steps to integrate Helm repositories into your workflow:

1. **Create a Git Repository:** Set up a Git repository on the Rafay platform. [Learn more](https://docs.rafay.co/integrations/repositories/overview/)

2. **Share the Repository:** Share the Helm repository with your projects.

3. **Utilize Helm Charts:** Access and deploy charts directly from the Helm repository for streamlined application deployment.

Terraform Configuration
```hcl
workload {
  name = "hello"
  chart_git_repo_path = "/hello-common/hello-0.1.3.tgz" # External Git Repository Chart Path
  chart_git_repo_branch = "main"   
  helm_chart_version = "0.1.3"
  helm_chart_name = "hello"
  chart_git_repo_name = "hello-repo"                    # Rafay Git Repository Name
  path_match_pattern = "/:project/:workload/:namespace"
  base_path = "hello-common"
  include_base_value = true
  delete_action = "delete"
  placement_labels = {
    "hello" = "enabled"
  }
}
```



## Example - Deploying Applications from a Repository

The example below demonstrates deploying three applications to projects using the same repository:

1. **echoserver:** Deployed using a chart from a Helm repository.
    - The chart is deployed from a Helm repository to all projects.
    - Utilize cluster labels to deploy the workload to specific clusters labeled with `"echoserver" = "enabled"`.

2. **hello:** Deployed using a chart from a Git repository.
    - The chart is deployed from a Git repository to all projects.
    - Utilize cluster labels to deploy the workload to specific clusters labeled with `"hello" = "enabled"`.

3. **httpecho-us:** Deployed using a chart from the project folder.
    - Utilize cluster labels to deploy the workload to specific clusters labeled with `"httpecho-us" = "enabled"`.

4. **httpecho-eu:** Deployed using a chart from the project folder.
    - Utilize cluster labels to deploy the workload to specific clusters labeled with `"httpecho-eu" = "enabled"`.


### Pre-Configuration Steps

1. **Create Projects:**
   - In the Rafay console, create three projects: `project1`, `project2`, and `project3`.

2. **Create/Import Clusters:**
   - For each project, create or import clusters as needed.
   - Label the clusters according to the examples provided in the repository.

   Example Labels:
   - To deploy application `echoserver`, label clusters with `"echoserver" = "enabled"`.
   - To deploy application `hello`, label clusters with `"hello" = "enabled"`.
   - To deploy application `httpecho-us`, label clusters with `"httpecho-us" = "enabled"`.
   - To deploy application `httpecho-eu`, label clusters with `"httpecho-eu" = "enabled"`.


   These labels are utilized in the Terraform configurations for deploying workloads.

3. **Create the following namespaces in the clusters:**
   - For the `echoserver` application, create the namespace `ns-echoserver`.
   - For the `hello` application, create the namespace `ns-hello`.
   - For the `httpecho-us` and `httpecho-eu` applications, create the namespace `ns-httpecho`.

### Repository Structure

The repository structure is organized to demonstrate deploying applications to the specified projects. Refer to the examples and configurations within the repository for a better understanding.

The folder structure for the projects is available on GitHub. You can find it [here](https://github.com/stephan-rafay/test-tfcd.git).


Terraform Configuration
```hcl
resource "rafay_workload_cd_operator" "operatordemo" {
	metadata {
	  name    = "operatordemo"
	  project = "demoorg"
	}
	spec {
	  repo_local_path = "/tmp/application-repo"
	  repo_url        = "https://github.com/stephan-rafay/test-tfcd.git"
	  repo_branch     = "main"
	  credentials {
	    username = var.git_user
	    token = var.git_token
	  }
  
	  workload {
		name = "echoserver"
		chart_helm_repo_name = "echo-server"
		helm_chart_version = "0.5.0"
		helm_chart_name = "echo-server"
		path_match_pattern = "/:project/:workload/:namespace"
		base_path = "echoserver-common"
		include_base_value = true
		delete_action = "delete"
		placement_labels = {
		  "echoserver" = "enabled"
		}
	  }

	  workload {
		name = "hello"
		chart_git_repo_path = "/hello-common/hello-0.1.3.tgz"
		chart_git_repo_branch = "main"   
		helm_chart_version = "0.1.3"
		helm_chart_name = "hello"
		chart_git_repo_name = "hello-repo"
		path_match_pattern = "/:project/:workload/:namespace"
		base_path = "hello-common"
		include_base_value = true
		delete_action = "delete"
		placement_labels = {
		  "hello" = "enabled"
		}
	  }
  
	  workload {
		name = "httpecho-us"
		helm_chart_version = "0.3.4"
		helm_chart_name = "http-echo"
		path_match_pattern = "/:project/:workload/:namespace"
		base_path = "httpecho-common"
		include_base_value = true
		delete_action = "delete"
		placement_labels = {
		  "httpecho-us" = "enabled"
		}
	  }
  
	  workload {
		name = "httpecho-eu"
		helm_chart_version = "0.3.4"
		helm_chart_name = "http-echo"
		path_match_pattern = "/:project/:workload/:namespace"
		base_path = "httpecho-common"
		include_base_value = true
		delete_action = "delete"
		placement_labels = {
		  "httpecho-eu" = "enabled"
		}
	  }	 
	}
	always_run = "${timestamp()}"
  }
```



Using  variables
```hcl

variable "git_token" {
  description = "git token for authentication"
  sensitive   = true
}
  
variable "git_user" {
  description = "git user for authentication"
  sensitive   = true
}
  
variable "delete_action_value" {
  description = "git user for authentication"
  # "none" (or) "delete" or "unpublish"
  default = "none" 
}
```

Using output.tf
```hcl
  output "workload_status" {
	description = "workload status"
	value       = "${rafay_workload_cd_operator.operator-demo.workload_status}"
  }
  
  output "workload_upserts" {
	description = "workload created or updated"
	value       ="${rafay_workload_cd_operator.operator-demo.workload_upserts}"
  }
  
  output "workload_decommissions" {
	description = "workload deleted (or) unpublished"
	value       = "${rafay_workload_cd_operator.operator-demo.workload_decommissions}"
  }
```


### Variables
| Variable | Description | Type | Default |
|----------|-------------|------|---------|
| metadata | Metadata of the Helm workload | `list` | `[]` |
| spec     | Specification of the Helm repository and workload | `list` | `[]` |
| workload_status   | Status of the Helm workloads | `list` | `[]` |
| workload_decommissions | List of Deleted/Unpublished Helm workloads | `list` | `[]` |
| workload_upserts | List of Updated/Created Helm workloads | `list` | `[]` |

#### `metadata`
Metadata of the secret sealer resource

| Property      | Description                                  | Type             | Default |
|---------------|----------------------------------------------|------------------|---------|
| description   | Description of the resource                  | `string`         | `""`    |
| name          | Name of the resource                         | `string`         | `""`    |
| project       | Project of the resource                      | `string`         | `""`    |


#### `spec`
Specification of the resource

| Property            | Description                                        | Type                  | Default |
|---------------------|----------------------------------------------------|-----------------------|---------|
| repo_url            | Repository URL                                     | `string` (required)   | -       |
| repo_branch         | Repository branch                                  | `string`              | `""`    |
| repo_local_path     | Repository local path to clone                     | `string`              | `/tmp/apprepo`    |
| credentials         | Credentials for repository access                  | `list(object)`        | `[]`    |
| insecure            | Allow insecure connection to the repository        | `bool`                | `false` |
| workload            | Workload specification                             | `list(object)`        | `[]`    |
| type                | Repository type                                    | `string`              | `""`    |


#### `workload` Object
Specification of the resource

| Property            | Description                                        | Type                  | Default |
|---------------------|----------------------------------------------------|-----------------------|---------|
| name                | Workload Name                                      | `string` (required)   | -       |
| helm_chart_name     | Helm Chart Name                                    | `string` (required)   | -       |
| helm_chart_version  | Helm Chart Version                                 | `string` (required)   | -       |
| chart_helm_repo_name| Helm Repository Name                               | `string`              | `""`    |
| chart_git_repo_name | Git Repository Name                                | `string`              | `""`    |
| chart_git_repo_path | Git Repository Path                                | `string`              | `""`    |
| chart_git_repo_branch| Git Repository Branch                             | `string`              | `""`    |
| include_base_value  | Include base values                                | `bool`                | `false` |
| base_path           | Common path for the chart                          | `string`              | `""`    |
| path_match_pattern  | Project/namespace/workload name path match pattern | `string` (required)   | -       |
| cluster_names       | Cluster names (comma-separated)                    | `string`              | `""`    |
| placement_labels    | Placement labels of the cluster                    | `map(string)`         | `{}`    |
| delete_action       | Delete Workload                                    | `string`              | `none`  |


##### `credentials` Object
Git Repository credentials for access
| Property     | Description                     | Type     | Default |
|--------------|---------------------------------|----------|---------|
| password     | Password for repository access   | `string` | `""`    |
| username     | Username for repository access   | `string` | `""`    |
| token        | Token for repository access      | `string` | `""`    |

#### `status`
Status of the workload resource.

The status is automatically generated by the resource upon the deployment of workloads.

| Property           | Description                           | Type                  | Default |
|--------------------|---------------------------------------|-----------------------|---------|
| project            | Project of the resource                | `string`              | `""`    |
| namespace          | Namespace of the resource              | `string`              | `""`    |
| workload_name      | Workload Name of the resource          | `string`              | `""`    |
| workload_version   | Workload Version of the resource       | `string`              | `""`    |
| repo_folder        | Repo path of the Workload resource     | `string`              | `"./apprepo"`|
| condition_status   | Condition Status                       | `number`              | `0`     |
| clusters           | Deployed clusters                      | `string`              | `""`    |
| condition_type     | Condition Type                         | `string`              | `""`    |
| reason             | Status message                         | `string`              | `""`    |

The status of the workloads can be displayed using `terraform show` command.
```hcl
status {
	condition_status = 2
	condition_type   = "WorkloadReady"
	namespace        = "ns0"
	project          = "parent-project"
	reason           = "workload is ready"
	repo_folder      = "application-repo/parent-project/ns0/echoserver"
	workload_name    = "echoserver"
	workload_version = "976d300"
}
```

### Delete

The resource provides the option to delete workloads when `delete_action` is set to `delete` (or) `unpublish`.

`delete` - Deletes the workload from the project.

`unpublish` - Unpublishes the workload from the project. Workloads are not deleted from the project, but are removed from the clusters.

The deployer performs the following steps:

1. Retrieves a list of all workloads created by it, based on a specific label key `k8smgmt.io/helm-deployer-tfcd`.
2. Delete (or) Unpublish the workloads from the project when corresponding folders or values are removed from the repository.
3. If `base_path` is specified and the chart file is removed from the base path, the deployer will delete (or) unpublish the workloads from all the project.

