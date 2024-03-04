# rafay_workload_cd_operator (Resource)

Helm Workload Deployer

## Table of Contents
- [Overview](#overview)
- [Usage](#usage)
  - [Features](#features)
  - [Repository Folder Structure](#repository-folder-structure)
    - [Path Match Pattern Examples](#path-match-pattern-examples)
      - [Example1](#example1)
	  - [Example2](#example2)
  - [Base Path](#base-path)
    - [Chart Selection for Deployment](#chart-selection-for-deployment)
  - [Managing Different Chart Versions](#managing-different-chart-versions)
- [Examples](#examples)
- [Variables](#variables)
  - [metadata](#metadata)
  - [spec](#spec)
  - [status](#status)
- [Examples](#examples)
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
  Allows the deployment of different chart versions for various workloads.

- **Workload Republishing:**
  Facilitates the republishing of workloads when the repository undergoes updates.

- **Workload Deletion and Unpublishing:**
  - Compares existing workloads with the repository.
  - Deletes or unpublishes workloads from the project when corresponding folders or values are removed from the repository.




### Repository Folder Structure
The repository folder structure is designed to organize Helm charts in a way that reflects the project name, namespace, and workload name. The folder structure is flexible and can be customized to suit your specific requirements using `path_match_pattern`.

Note:  `project` and `namespace` are mandatory, and `workload` is optional.

Examples:

- `/:project/:namespace/:workload` <br>
- `/somefolder/:project/:namespace/:workload` <br>
- `/somefolder/:project/:namespace/:workload/anotherfolder` <br>
- `/somefolder/production/:project/:namespace/:workload` <br>
- `/:workload/:namespace/:project` etc. <br>

#### Path Match Pattern Examples

#### Example1: All three folders are present
If the folder structure is like below, then the `path_match_pattern` can be specified in the following format:<br> `/:project/:namespace/:workload` <br>
```
- common
  - chart-1.0.1.tgz
- project1
  - namespace1
    - workload1
	  - values.yaml
- project2
  - namespace1
    - workload1
	  - values.yaml
```


In this scenario, if `base_path = "common"`, two workloads will be deployed in respective projects using the same chart file located in the `common` folder. <br>
1. The first workload will use the chart file from the `common` folder and values from the `project1/namespace1/workload1` folder. <br>
2. The second workload will use the chart file from the `common` folder and values from the `project2/namespace1/workload1` folder. <br>


#### Example2: Only `project` and `namespace` are present
If the folder structure is like below, then the `path_match_pattern` can be specified in the following format:<br> `/:project/:namespace` <br>
```
- common
  - chart-1.0.1.tgz
- project1
  - namespace1
	- values.yaml
- project2
  - namespace1
	- values.yaml
```


In this scenario, if `base_path = "common"`, two workloads will be deployed in respective projects using the same chart file located in the `common` folder. <br>
1. The first workload will use the chart file from the `common` folder and values from the `project1/namespace1` folder. <br>
2. The second workload will use the chart file from the `common` folder and values from the `project2/namespace1` folder. <br>

The workload name will be derived from the chartname `chart-1-0-1`. <br>


#### Example3: All three folders are present inside another folder
If the folder structure is like below, then the `path_match_pattern` can be specified in the following format:<br> `/production/:project/:namespace/:workload` <br>
```
- production
  - common
    - chart-1.0.1.tgz
  - project1
    - namespace1
      - workload1
	    - values.yaml 
  - project2
    - namespace1
      - workload1
	  - values.yaml
```

In this scenario, if `base_path = "production/common"`, two workloads will be deployed in respective projects using the same chart file located in the `production/common` folder. <br>
1. The first workload will use the chart file from the `production/common` folder and values from the `production/project1/namespace1/workload1` folder. <br>
2. The second workload will use the chart file from the `production/common` folder and values from the `production/project2/namespace1/workload1` folder. <br>


### Base Path
The `base_path` is the common path for the chart.

#### Chart and Value Selection for Deployment

If the chart is found in the respective folder, it takes precedence for deployment over the specified `base_path`. This ensures that the deployment process prioritizes the chart present in the workload folder, providing flexibility to deploy different version of the applicatin.

If the chart or value file is missing for a specific folder, the entire folder will be skipped during the deployment process. This ensures that the deployment focuses only on folders with the necessary files, avoiding any errors caused by missing components.

If more than one value file is present in the folder, the deployment will use all the files with extenstions `.yaml` or `.yml` present in the folder.

If `include_base_values` is set to `true`, the values from the `base_path` will be merged with the values from the workload folder. 

### Managing Different Chart Versions

When dealing with different chart versions for various workloads, follow these steps:

1. **Do not set `base_path`.**
2. Include the chart file in the respective workload folder.
3. Ensure the presence of a `values.yaml` file in the same workload folder.

The deployment process will utilize the chart file located in the workload folder.


## Examples
Using `placement_labels`
```hcl
resource "rafay_workload_cd_operator" "operator-demo" {
	metadata {
	  name    = "operator-demo"
	  project = "demo"
	}
	spec {
	  repo_local_path = "./application-repo"
	  repo_url        = "https://github.com/helm-workloads/test.git"
	  repo_branch     = "main"
	  credentials {
		username = "demo-user"
		token = "ghp_XXXXAPIKEYXXXX"
	  }
  
	  base_path = "common"
	  path_match_pattern = "/:project/:namespace/:workload"
	  
	  placement_labels = {
		"team" = "myteam"
	  }
	}
	always_run = "${timestamp()}"
  }
```

Using `cluster_names`

```hcl
resource "rafay_workload_cd_operator" "operator-demo" {
	metadata {
	  name    = "operator-demo"
	  project = "demo"
	}
	spec {
	  repo_local_path = "./application-repo"
	  repo_url        = "https://github.com/helm-workloads/test.git"
	  repo_branch     = "main"
	  credentials {
		username = "demo-user"
		token = "ghp_XXXXAPIKEYXXXX"
	  }
  
	  base_path = "common"
	  path_match_pattern = "/:workload/:namespace/:project"
	  
	  cluster_names = "cluster1,cluster2"

	}
	always_run = "${timestamp()}"
  }
```


Using `delete_action = "unpublish"`

```hcl
resource "rafay_workload_cd_operator" "operator-demo" {
	metadata {
	  name    = "operator-demo"
	  project = "demo"
	}
	spec {
	  repo_local_path = "./application-repo"
	  repo_url        = "https://github.com/helm-workloads/test.git"
	  repo_branch     = "main"
	  credentials {
		username = "demo-user"
		token = "ghp_XXXXAPIKEYXXXX"
	  }
  
	  base_path = "common"
	  path_match_pattern = "/:workload/:namespace/:project"
	  
	  cluster_names = "cluster1,cluster2"
	  delete_action = "unpublish"

	}
	always_run = "${timestamp()}"
  }
```


### Variables
| Variable | Description | Type | Default |
|----------|-------------|------|---------|
| metadata | Metadata of the Helm workload | `list` | `[]` |
| spec     | Specification of the Helm repository and workload | `list` | `[]` |
| status   | Status of the Helm workload | `list` | `[]` |


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
| base_path           | Common path for the chart                          | `string`              | `""`    |
| repo_local_path     | Repository local path to clone                     | `string`              | `./apprepo`    |
| path_match_pattern  | Project/namespace/workload name path match pattern | `string` (required)   | -       |
| cluster_names       | Cluster names (comma-separated)                    | `string`              | `""`    |
| placement_labels    | Placement labels of the cluster                    | `map(string)`         | `{}`    |
| credentials         | Credentials for repository access                  | `list(object)`        | `[]`    |
| repo_url            | Repository URL                                     | `string` (required)   | -       |
| repo_branch         | Repository branch                                  | `string`              | `""`    |
| insecure            | Allow insecure connection to the repository        | `bool`                | `false` |
| delete_action       | Delete Workload                                    | `string`              | `none`  |
| type                | Repository type                                    | `string`              | `""`    |

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

