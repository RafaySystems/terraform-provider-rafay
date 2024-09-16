# Goals

## Moving towards Terrafrom Framework


The current rafay provider uses Terraform SDK, and most of feature development has been stopped for SDK and it's been recommended to migrate to new plugin framework as the new framework comes with couple of advantages and addresses the on-going issues with ordering issues with list and unnecessary diff during `terraform plan`. These issues can be well addressed if move towards the [plugin framework](https://developer.hashicorp.com/terraform/plugin/framework-benefits). Instead of migrating all the resource to framework, we're starting to use it for newly added resources. So, we need to a way co-exist both SDKV2 and plugin framework in a single provider. So We're using [multiplexing](https://developer.hashicorp.com/terraform/plugin/mux) to acheive that. 



## How to add a new Resource to Terraform

You can generate the plugin framework code using provider spec json. You need to come up with initial provider spec which 

```yaml

```



## Identifying Terraform Resource, Version and Kind
Refer to the [Resources](./Resources.md) to identify the group, version and kind of the resource.

## Goals
The hub version of a resource is meant to be the Gitops/External view of one or more resources exposed by internal micro services. Hub exposes both `OpenAPI3` and `OpenAPI2` definition for each resource and generates the TF SDK schema. The current generator converts each proto field to Terraform type.The conversion is straight forward for primitive types and but If proto field uses `Oneof` of nested objects, the current generator flattens every attribute and make it available for it's parent Kind. 

For ex:

If you take a look at cluster proto, It has `Config` field which can be any of MKS/EKS/GKE/AKS cluster but we're treating these clusters as an individual resource in Terraform. This is where we're deviating from having standarized schema for all the interfaces. Another issue is that convertor extracts all the fields and flattens under `Config`. If nested objects have same field name but different definition, You can have only expose either one of the definitions.


## Seperate package for each concept

For example, the datasource package contains the functionality for implementing data sources, and the provider package contains the functionality for implementing the provider. This separation helps make it clear how and when to use each type.



In contrast, the SDK requires you to implement abstract, recursive types, such as helper/schema.Resource type and helper/schema.Schema type. A schema.Resource implementation could be a managed resource, a data source, or block definition within a schema. These generic abstractions make it difficult to understand the specific requirements for each type. For example, a data source requires a schema and read functionality while a block only requires a schema.



## Expanded Access to Configuration, Plan, and State Data


Providers receive up to three sources of schema-based data during Terraform operation requests: `configuration`, `plan`, and `prior state`. The SDKv2 combines this data into a single schema.ResourceData type, which you implement differently depending on the operation. Certain ResourceData methods are only valid during certain operations and trying to get data from an explicit source is problematic in many cases.

In the following SDKv2 example, the code comments highlight issues with the single data type:

```go:
func MksClusterResourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
  d.Get("...") // plan unless unknown; no explicit access to configuration
  d.GetChange("...") // extraneous old value, use d.Get() instead
  d.HasChange("...") // always true, no prior state
  d.Set("...") // saved into new state
}

func MksClusterResourceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
  d.Get("...") // prior state
  d.GetChange("...") // no changes as only prior state is available
  d.HasChange("...") // always false
  d.Set("...") // saved into new state
}

func MksClusterResourceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
  d.Get("...") // plan unless unknown; no explicit access to configuration or prior state
  d.GetChange("...") // prior state and plan unless unknown
  d.HasChange("...") // comparison of prior state and plan
  d.Set("...") // saved into new state
}

func MksClusterResourceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
  d.Get("...") // prior state
  d.GetChange("...") // no changes as only prior state is available
  d.HasChange("...") // always false
  d.Set("...") // extraneous, resource destroy leaves no state
}
```


The framework alleviates these issues by exposing configuration, plan, and state data as separate attributes on request and response types that only expose the data available to the given operation.

In the following framework example, the code comments show the available data that matches each operation.

```go:

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                     = &MksClusterResource{}
	_ resource.ResourceWithConfigure        = &MksClusterResource{}
	_ resource.ResourceWithImportState      = &MksClusterResource{}
	_ resource.ResourceWithConfigValidators = &MksClusterResource{}
)

func NewMksClusterResource() resource.Resource {
	return &MksClusterResource{}
}

// MksClusterResource defines the resource implemSharentation.
type MksClusterResource struct {
	client typed.Client
}

func (r *MksClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mks_cluster"
}

func (r *MksClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fw.MksClusterResourceSchema(ctx)
}

func (r *MksClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, _ := req.ProviderData.(typed.Client)
	// Save the client for use in CRUD operations
	r.client = client
}

func (r *MksClusterResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("spec.cloud_credentials"),
			path.MatchRoot("spec.config.cluster_ssh"),
		),
	}
}

func (r *MksClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	var data fw.MksClusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert the Terraform model to a Hub model
	hub, daig := fw.ConvertMksClusterToHub(ctx, data)
	if daig.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create cluster, got error: %s", daig))
		return
	}

	// Create the cluster
	err := cluster.ApplyMksV3Cluster(ctx, r.client, hub)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create cluster, got error: %s", err))
		return
	}
}

func (r *MksClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform prior state data into the model
	var state fw.MksClusterModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the cluster from the Hub
	c, err := r.client.InfraV3().Cluster().Get(ctx, options.GetOptions{
		Name:    state.Metadata.Name.ValueString(),
		Project: state.Metadata.Project.ValueString(),
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read the cluster, got error: %s", err))
		return
	}

	// Convert the Hub model to a Terraform model
	daigs := fw.ConvertMksClusterFromHub(ctx, c, &state)
	if daigs.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to convert the cluster, got error: %s", daigs))
		return
	}
	// Save the refreshed state into Terraform
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func (r *MksClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	var plan fw.MksClusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert the Terraform model to a Hub model
	hub, daigs := fw.ConvertMksClusterToHub(ctx, plan)
	if daigs.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update the cluster, got error: %s", daigs))
		return
	}

	// Call the Hub to Apply the cluster
	err := cluster.ApplyMksV3Cluster(ctx, r.client, hub)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update cluster, got error: %s", err))
		return
	}

	// Wait for the cluster operation to complete
	ticker := time.NewTicker(time.Duration(60) * time.Second)
	defer ticker.Stop()
	timeout := time.After(time.Duration(90) * time.Minute)
	daigs = fw.WaitForClusterApplyOperation(ctx, r.client, hub, timeout, ticker)

	if daigs.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update cluster, got error: %s", daigs))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}







func (r MksClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
  req.Config // configuration data
  req.Plan // plan data
  // No req.State as it is always null
  // No resp.Config as configuration cannot be set by provider during creation
  // No resp.Plan as plan cannot be set by provider during creation
  resp.State // new state data to save
}

func (r MksClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.CreateResponse) {
  // No req.Config as configuration cannot be read by provider during read
  // No req.Plan as there is no plan during read
  req.State // prior state data
  // No resp.Config as configuration cannot be set by provider during read
  // No resp.Plan as there is no plan during read
  resp.State // new state data to save
}

func (r MksClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
  req.Config // configuration data
  req.Plan // plan data
  req.State // prior state data
  // No resp.Config as configuration cannot be set by provider during update
  // No resp.Plan as plan cannot be set by provider during update
  resp.State // new state data to save
}

func (r MksClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
  // No req.Config as configuration cannot be read by provider during delete
  // No req.Plan as it is always null
  req.State // prior state data
  // No resp.Config as configuration cannot be set by provider during delete
  // No resp.Plan as it cannot be adjusted
  resp.State // only available to explicitly remove on error
}

func (r *MksClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	
}
```






The decouping of the hub resource view from the storage resource view, is facilitated by two abstrations, *convertors* and *accessors*.

Convertors handles hub<->spoke conversions. They have to implement [hub.Convertor](https://github.com/RafaySystems/rafay-common/blob/7b780ecae71386fcc6538ec1132d93aa265c70d8/pkg/hub/interface.go) interface. Spoke in this case a logical view of the underlying storage resource(s).
```go
type Convertor interface {
	Init(interface{}) error
	// ConvertTo converts spoke version to hub version
	ConvertTo(dst Hub) error
	// ConvertFrom converts hub version to spoke version
	ConvertFrom(src Hub) error
}
```

Accessor handles accessing the Spoke view from underlying storage (RPC to service(s)). They have to implement [storage.Accessor](./pkg/hub/storage/interface.go).
```go
type Accessor interface {
	// Get fetches a resource from storage
	Get(context.Context, GetOptions) (hub.Hub, error)
	// List fetches a list of resources from storage
	List(context.Context, ListOptions) ([]hub.Hub, error)
	// Apply creates or updates a resource in storage
	Apply(context.Context, hub.Hub, ApplyOptions) error
	// Delete deletes a resource in storage
	Delete(context.Context, DeleteOptions) error
	// Status returns the status of a resource in storage
	Status(context.Context, StatusOptions) (hub.HubStatus, error)
}
``` 

The relationship between hub view, accessor, convertor, storage view and the service can illustrated as below,

```
+----------+       +-----------+      +--------------+
| Hub View |<----->| T |<---->| Storage View |
+----------+       +-----+-----+      +--------------+
     ^                   |                   ^
     |                   |                   |
     |                   |                   |
     |                   |                   |
     v                   |                   v
+----------+             |              +---------+
| Accessor | <-----------+------------->| Service |
+----------+                            +---------+
                                                
```

Here, the Convertor converts to and fro from the hub view and storage view. The storage view can refer to a single resource or multiple resources.

The backing service(s) only understand the storage view and the Accessor only understands the hub view. Convertor facilitates the 
coversion between storage and hub view and vice versa.

The hub definition of a resource follow the following format at a high level

```yaml
apiVersion:
kind:
metadata:
spec:
```

## API Version & Kind

```apiVersion``` and ```kind``` are specified during the resource [registration](https://github.com/RafaySystems/rafay-common/blob/7b4e2843532361aba9f5272fcd205ee082b9c847/pkg/hub/conversion/types/hub/infra/register.go). 

## Metadata

```metadata``` should be [*commonpb.Metadata](https://github.com/RafaySystems/rafay-common/blob/0c26421c81fed2aedbb5ba02509674171c6a7b17/proto/types/hub/commonpb/common.proto). Only fields that will be serialized in the metadata will be ```name, description, labels, annotations and project```. ```name``` and ```project``` fields in metadata are **manadatory** for every resource

## Spec
Spec should contain all the necessary information to describe the intented state of a resource. If the resource is refering to other resources, then it should only do so by names. **IDs** should not be used anywhere. 
## Sharing (Spec and Rules)
If the resource supports sharing across projects, the sharing definition should be [*commonpb.SharingSpec](https://github.com/RafaySystems/rafay-common/blob/0c26421c81fed2aedbb5ba02509674171c6a7b17/proto/types/hub/commonpb/common.proto) and the field name should be ```sharing```.

```yaml
spec:
 sharing:
 enabled: "true or false"
 projects:
 - name: "project name or '*' for all projects in the organization"
```

The following rules should be followed for defining the sharing spec

1. If the resource is only shared with the owner project, then sharing spec should as below. This is the default value.

```yaml
spec:
 sharing:
   enabled: false
``` 

2. If the resource is shared with all the projects in the organization then sharing spec should be 
```yaml
spec:
 sharing:
   enabled: true
   projects:
   - name: "*"
```

3. If the resource is shared with a subset of projects in the organization then sharing
spec should be
```yaml
spec:
 sharing:
   enabled: true
   projects:
   - name: "project-1"
   - name: "project-2"
   - name: "project-3"
```

This complexity should be absorbed in the spoke convertor.
## Hub View interfaces
We need to implement few interafaces for the hub resources, so that
we can automatically detect and enable some behaviours.
### Hydrator
As we only refer to the projects in the resources by name in the Hub view and the storage refers to the projects by ID, we need to convert transparently convert the projects name to ID and back. This is facilitated by [Project Hydrator](./pkg/storage/proxy/project_hydrator.go). Which is implemented as a [Storage Proxy](./pkg/storage/proxy/interface.go). Storage proxy is can be though of like a storage middleware, it perform actions before or after the upstream Accessor method is called. 

```go
func(storage.Accessor) storage.Accessor
```

Any Hub resource which implements the [Project Hydrator Interface](https://github.com/RafaySystems/rafay-common/blob/7b780ecae71386fcc6538ec1132d93aa265c70d8/pkg/hub/interface.go) will be automatically hydrated (project name -> id) and dehydrated (id -> project name). 


```go
type ProjectHydrator interface {
	HydrateProject(func(meta *commonpb.ProjectMeta) error) error
}
```

[Blueprint](https://github.com/RafaySystems/rafay-common/blob/7b780ecae71386fcc6538ec1132d93aa265c70d8/proto/types/hub/infrapb/blueprint_hydrate.go) implements the Hydrator interface can be used as a reference for implement other resources.

```go
func (b *Blueprint) Hydrate(f func(meta *commonpb.ProjectMeta) error) error {

	projectMeta := &commonpb.ProjectMeta{
		Name: b.Metadata.Project,
		Id:   b.Metadata.ProjectID,
	}

	err := f(projectMeta)
	if err != nil {
		err = errors.Wrap(err, "unable to hydrate metadata")
		return err
	}

	b.Metadata.Project = projectMeta.Name
	b.Metadata.ProjectID = projectMeta.Id

	if b.Spec.Sharing != nil {
		for i := range b.Spec.Sharing.Projects {
			if b.Spec.Sharing.Projects[i].Name == "*" {
				continue
			}
			err := f(b.Spec.Sharing.Projects[i])
			if err != nil {
				err = errors.Wrap(err, "unable to hydrate sharing project")
				return err
			}
		}
	}

	return nil
}
```

Resouce specific hydrators should be implemented in ```proto/types/{group}pb/{kind}_hydrator.go```
### Lister
```go
// HubLister is the interface
// for working with lists of
// hub resources
type HubLister interface {
	GetApiVersion() string
	GetKind() string
	SetAPIGroupVersionKind(schema.APIGroupVersionKind)
	GetMetadata() *commonpb.ListMetadata
	GetHubs() []Hub
	AddHubs([]Hub)
}
```
All the hub resources should implement `HubLister` interface
to work with lists of hub resources. This is needed for
the `codec` to encode and decode lists of hub resources.
They should be implemented under `rafay-common/proto/types/hub/{group}pb/{kind}_list.go`

```go
// .../rafay-common/proto/types/hub/infrapb/blueprint_list.go
package infrapb

import (
	"github.com/RafaySystems/rafay-common/pkg/hub"
	"github.com/RafaySystems/rafay-common/pkg/hub/schema"
)

var _ hub.HubLister = (*BlueprintList)(nil)

func (x *BlueprintList) GetHubs() []hub.Hub {
	var hubs []hub.Hub
	for _, item := range x.Items {
		hubs = append(hubs, item)
	}
	return hubs
}

func (x *BlueprintList) AddHubs(hubs []hub.Hub) {
	for _, h := range hubs {
		x.Items = append(x.Items, h.(*Blueprint))
	}
}

func (x *BlueprintList) SetAPIGroupVersionKind(agvk schema.APIGroupVersionKind) {
	x.ApiVersion = agvk.APIVersion()
	x.Kind = agvk.Kind
}

```
### Artifact Accessor
```go
// ArtifactAccessor is the interface for
// accessing artifacts from a Hub resource
type ArtifactAccessor interface {
	ArtifactGet(name string) ([]byte, error)
	ArtifactList() ([]string, error)
	ArtifactSet(name string, data []byte) error
}
```
Some hub resources might need to work with files (as opaque binary blobs). The files are supposed to be read from local file system and are referenced in the schema as relative paths. If a resource 
implements the `ArtifactAccessor` then the hub will transparently handle the gitops sync and add the respective API endpoints for
working with the artifacts. They should be implemented under 
`rafay-common/proto/types/hub/{group}pb/{kind}_artifacts.go`.
```go
// .../rafay-common/proto/types/hub/appspb/workload_artifacts.go
package appspb

import (
	"github.com/RafaySystems/rafay-common/pkg/hub"
	commonpb "github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/pkg/errors"
)

var _ hub.ArtifactAccessor = (*Workload)(nil)

func (x *Workload) ArtifactSet(name string, data []byte) error {
	if x.Spec != nil {
		if x.Spec.Artifact != nil {
			switch a := x.Spec.Artifact.Artifact.(type) {
			case *commonpb.ArtifactSpec_UploadedHelm:
				if a.UploadedHelm.ChartPath != nil {
					if a.UploadedHelm.ChartPath.Name == name {
						a.UploadedHelm.ChartPath.Data = data
						return nil
					}
				}
				for i := range a.UploadedHelm.ValuesPaths {
					if a.UploadedHelm.ValuesPaths[i].Name == name {
						a.UploadedHelm.ValuesPaths[i].Data = data
						return nil
					}
				}
			case *commonpb.ArtifactSpec_UploadedYAML:
				for i := range a.UploadedYAML.Paths {
					if a.UploadedYAML.Paths[i].Name == name {
						a.UploadedYAML.Paths[i].Data = data
						return nil
					}
				}
			case *commonpb.ArtifactSpec_HelmInHelmRepo:
				for i := range a.HelmInHelmRepo.ValuesPaths {
					if a.HelmInHelmRepo.ValuesPaths[i].Name == name {
						a.HelmInHelmRepo.ValuesPaths[i].Data = data
						return nil
					}
				}
			}
		}
	}

	return errors.Errorf("artifact not found '%s'", name)
}

func (x *Workload) ArtifactGet(name string) ([]byte, error) {
	if x.Spec != nil {
		if x.Spec.Artifact != nil {
			switch a := x.Spec.Artifact.Artifact.(type) {
			case *commonpb.ArtifactSpec_UploadedHelm:
				if a.UploadedHelm.ChartPath != nil {
					if a.UploadedHelm.ChartPath.Name == name {
						return a.UploadedHelm.ChartPath.Data, nil
					}
				}
				for i := range a.UploadedHelm.ValuesPaths {
					if a.UploadedHelm.ValuesPaths[i].Name == name {
						return a.UploadedHelm.ValuesPaths[i].Data, nil
					}
				}
			case *commonpb.ArtifactSpec_UploadedYAML:
				for i := range a.UploadedYAML.Paths {
					if a.UploadedYAML.Paths[i].Name == name {
						return a.UploadedYAML.Paths[i].Data, nil
					}
				}
			case *commonpb.ArtifactSpec_HelmInHelmRepo:
				for i := range a.HelmInHelmRepo.ValuesPaths {
					if a.HelmInHelmRepo.ValuesPaths[i].Name == name {
						return a.HelmInHelmRepo.ValuesPaths[i].Data, nil
					}
				}
			}
		}
	}

	return nil, errors.Errorf("artifact not found '%s'", name)
}

func (x *Workload) ArtifactList() ([]string, error) {
	var ret []string
	if x.Spec != nil {
		if x.Spec.Artifact != nil {
			switch a := x.Spec.Artifact.Artifact.(type) {
			case *commonpb.ArtifactSpec_UploadedHelm:
				if a.UploadedHelm.ChartPath != nil {
					ret = append(ret, a.UploadedHelm.ChartPath.Name)
				}
				for i := range a.UploadedHelm.ValuesPaths {
					ret = append(ret, a.UploadedHelm.ValuesPaths[i].Name)
				}
			case *commonpb.ArtifactSpec_UploadedYAML:
				for i := range a.UploadedYAML.Paths {
					ret = append(ret, a.UploadedYAML.Paths[i].Name)
				}
			case *commonpb.ArtifactSpec_HelmInHelmRepo:
				for i := range a.HelmInHelmRepo.ValuesPaths {
					ret = append(ret, a.HelmInHelmRepo.ValuesPaths[i].Name)
				}
			}
		}
	}
	return ret, nil
}
```
### Versioner
```go
// HubVersioner is the interface for
// accessing version of a Hub resource
type HubVersioner interface {
	Version() string
}
```
Some hub resources need to support a version. In the gitops context
only the latest version is tracked. When the resource implements
`Versioner` interface, hub transparently adds support handling versions. They should be implemented under 
`rafay-common/proto/types/hub/{group}pb/{kind}_version.go`.
```go
// .../rafay-common/proto/types/hub/infrapb/addon_version.go
package infrapb

import (
	"github.com/RafaySystems/rafay-common/pkg/hub"
)

var _ hub.HubVersioner = (*Addon)(nil)

func (x *Addon) Version() string {
	if x.GetSpec() != nil {
		return x.GetSpec().Version
	}
	return ""
}

```
### Status
```go
// HubStatus is the interace for
// accessing status of a hub resource
type HubStatus interface {
	Hub
	GetStatus() *commonpb.Status
	SetStatus(status *commonpb.Status)
}
```
Some hub resources support accessing status of the resource. This
status is used to track the status of the gitops sync. This status
should be treated as a compound status (ex. tracking the progress of a workload deployment or provisioing of a cluster). If a resource 
implements the `HubStatus` then the hub will transparently handle the gitops sync status and add the respective API endpoints for
working with the status. hey should be implemented under 
`rafay-common/proto/types/hub/{group}pb/{kind}_status.go`.
```go
// .../rafay-common/proto/types/hub/appspb/workload_status.go
package appspb

import (
	"github.com/RafaySystems/rafay-common/pkg/hub"
	commonpb "github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
)

var _ hub.HubStatus = (*Workload)(nil)

func (x *Workload) SetStatus(status *commonpb.Status) {
	x.Status = status
}
```

## Definition
Hub resource should be defined using protobuf (v3) IDL under ```rafay-common/proto/types/hub/{group}pb/{kind}.proto```. We should only use scalar, repeated, map types. We should use strings for Enumarated types. Refer to the  [Language Guide](https://developers.google.com/protocol-buffers/docs/proto) for knowing more about supported types in protobuf.

The defined resources can be compiled using ```rafay-common/build/hub.sh```. The compiled resources are used for serialzation and deserialzation. As the generated resources follow protobuf (JSON) serialization semantics, the default values of resources will not be serialzed. To serialize default values, use ```gomodifytags``` to override the default json tags. The overrides should go into ```rafay-common/proto/types/hub/{group}pb/fix.go```. commonpb has some overriden tags which can be used as an [example](../rafay-common/proto/types/commobpb/fix.go) to implement in other places. 

The definition of the Hub resource should be self contained. It can refer to other resources (by name), but the definition by itself should not be split into multiple sub resources.

## Defaulting
```go
type HubDefaulter func(Hub)
```

We need to implement a defaulting function for every hub resource we define. The defaulting function is called everytime a hub is 
created. An [blueprint default](https://github.com/RafaySystems/rafay-common/blob/bb385f240a39620f08919e166daed223ab40e84f/pkg/hub/conversion/types/hub/infra/blueprint_default.go) for Blueprint defaulting function can be used as example for implementing 
other resources. 

## Hub Resource Scheme Registration
Once the hub is defined, we need to register it in [Scheme](https://github.com/RafaySystems/rafay-common/blob/7b4e2843532361aba9f5272fcd205ee082b9c847/pkg/hub/scheme.go) so that other components can dynamically discover the resource at runtime. The scheme is also used to dynamically create (and initalize) the hub resource at runtime.

The Hub type registration should be done at ```rafay-common/pkg/hub/conversion/types/{group}/register.go```. Registration in scheme should include the Hub resource and optionally the defaulting function.

```go
// .../rafay-common/pkg/hub/conversion/types/hub/infra/register.go
package infra

import (
	"github.com/RafaySystems/rafay-common/pkg/hub"
	"github.com/RafaySystems/rafay-common/pkg/hub/schema"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
)

const (
	Group = "infra"
)

var (
	DefaultInfraGroupVersion = schema.APIGroupVersion{
		Group:       Group,
		Version:     "v3",
		GroupSuffix: schema.DefaultGroupSuffix,
	}
	localSchemeBuilder   = &DefaultSchemeBuilder
	DefaultAddToScheme   = localSchemeBuilder.AddToScheme
	DefaultSchemeBuilder = SchemeBuilderForGroupSuffix(schema.DefaultGroupSuffix)
)

// SchemeBuilderForGroupSuffix builds scheme for custom group suffix
// should be used for white labeling
func SchemeBuilderForGroupSuffix(suffix string) hub.SchemeBuilder {
	return hub.NewSchemeBuilder(addKnownTypeForGroupSuffix(suffix))
}

func addKnownTypeForGroupSuffix(suffix string) func(scheme *hub.Scheme) error {
	return func(scheme *hub.Scheme) error {
		scheme.AddKnownHubs(schema.APIGroupVersion{
			Group:       Group,
			Version:     "v3",
			GroupSuffix: suffix,
		},
			hub.HubConfig{
				Hub:       (*infrapb.Blueprint)(nil),
				Defaulter: defaultHubBlueprint,
				HubList:   (*infrapb.BlueprintList)(nil),
			},
			hub.HubConfig{
				Hub:       (*infrapb.Addon)(nil),
				Defaulter: defaultHubAddon,
				HubList:   (*infrapb.AddonList)(nil),
			},
		)

		return nil
	}
}
```

## Resource Spoke Definition and Convertor
We have to define the Spoke version of the resource and the associated Convertor. The Spoke version of a resource(s) should logically map to one. Note that at the Spoke there can be multiple resource can be logically mapped one Hub resource. For example, Blueprint and Blueprint Snapshot logically maps to one Bluperint Hub resource. So the Spoke 
definition of Blueprint will include both of them. And the convertor needs both the resources to convert to Hub and back. [Blueprint](https://github.com/RafaySystems/rafay-common/blob/d18a46fd25b183f34ab79f4b7a92b0f14565c3eb/pkg/hub/conversion/types/spoke/storage/infra/blueprint.go) can be used as an example to 
implement new resources.

### Spoke initialization
Like Hub there is no explicit initalization (or defaulting) for the
spoke as the source of truth does not lie with the "rafay-hub" service. 
It is expected that we receive a full initialized version of the 
the spoke.

## Resource Spoke Scheme Registration
Once the Spoke Definition and the Convertor are created, we need to
register the spoke for the hub in the scheme. This should be done
in ```rafay-common/pkg/hub/conversion/types/spoke/storage/{group}/register.go```.
Note that "storage" is a kind of spoke, we can will additional spokes in the future to handle "CLI", "Swagger" views.

```go
scheme.AddSpokesForHub(&infrapb.Blueprint{}, hub.SpokeConvertor{
	Type:      reflect.TypeOf((*StorageBlueprint)(nil)),
	Convertor: &blueprintConvertor{},
})
```

## Resource Storage Accessor
Storage Accessor provides a way to access and mutate the underlying storage. Just like the storage spoke, it is possible that there are
multiple resources (or services) we need to talk to make this happen.
All this complexity should be absorbed by the Accessor. The accessor
should be defined in ```pkg/storage/{group}/{kind}.go```. 

### Caching
Due to the nature of syncing, we need to access the resources multiple times and this might result in large number of RPC calls as the cardinality of the sync set increases. 

We can cache the access to avoid duplicate calls to the upstream services. Caching has to done for "Get" and "List" operations and
invalidation has to done whenever, we "Apply" or "Delete" a resource.

[groupcache](https://github.com/mailgun/groupcache) is used for caching and sharding the cache across multiple replicas of the hub service. The cache sharding (through discovering other replicas) is automatically setup during the service startup. Each storage accessor has to setup up its own individual cache group for "get" and "list" operations. It is
most effective to cache the hub directly, so that we avoid the conversion penality as well. We setup the cache to act as a read through LRU (with expiry) cache. 

### Cache group initalization
The cache group should be initalized in the constructor of the storage
accessor [example](./pkg/storage/infra/blueprint.go). Names for the "get" and "list" group along with their sizes and expiry should be decided when initializing the cache group. 

```go
func setupBlueprintsCache(configPool configrpcv2.ConfigPool, ncf hub.NewConvertorFunc, nhf hub.NewHubFunc) {
	blueprintsOnce.Do(func() {
		groupcache.NewGroup(blueprintsGroupName, blueprintsCacheSize, groupcache.GetterFunc(
			func(ctx context.Context, key string, dest groupcache.Sink) error {
				_log.Infow("fetching blueprint", "key", key)

				cc, err := configPool.NewClient(ctx)
				if err != nil {
					err = errors.Wrap(err, "unable to create config client")
					return err
				}

				defer cc.Close()

				blueprint, err := getBlueprint(ctx, cc, ncf, nhf, key)
				if err != nil {
					return err
				}
				if err := dest.SetProto(blueprint, time.Now().Add(blueprintsCacheExpiry)); err != nil {
					return err
				}
				return nil
			},
		))

		groupcache.NewGroup(blueprintsListGroupName, blueprintsListCacheSize, groupcache.GetterFunc(
			func(ctx context.Context, key string, dest groupcache.Sink) error {
				_log.Infow("fetching blueprint list", "key", key)

				cc, err := configPool.NewClient(ctx)
				if err != nil {
					err = errors.Wrap(err, "unable to create config client")
					return err
				}

				defer cc.Close()

				blueprint, err := getBlueprints(ctx, cc, ncf, nhf, key)
				if err != nil {
					return err
				}
				if err := dest.SetProto(blueprint, time.Now().Add(blueprintsCacheExpiry)); err != nil {
					return err
				}
				return nil
			},
		))
	})
}

```

### Readthrough Caching
The readthrough cache is used for accessing "Get" or "List" in the storage accessor. The "GetOptions" or "ListOptions" are serialized to JSON and used as a cache key. [Blueprint](./pkg/storage/infra/blueprint_util.go) can be used to imlement this for other resources.
The cache should be invalidated whever "Apply" or "Delete" methods are called in the accessor. Additionally a notification is registed in the
upstream service for invalidating cache when the storage resource is mutated.

```go
func (b *blueprintAccessor) Get(ctx context.Context, opts storage.GetOptions) (hub.Hub, error) {

	group := groupcache.GetGroup(blueprintsGroupName)

	h, err := b.newHub()
	if err != nil {
		err = errors.Wrap(err, "unable to create hub")
		return nil, err
	}

	blueprint := h.(*infrapb.Blueprint)

	key, _ := opts.MarshalJSON()

	err = group.Get(ctx, string(key), groupcache.ProtoSink(blueprint))
	if err != nil {
		return nil, err
	}

	return blueprint, nil
}
```

## Resource Storage Accessor Registration
Storage accessor has to registed so that it can be discovered and
and accessed at runtime. While registering the accessor, we need to 
provide the "name" by which this storage is accessed. If cache groups are used we need to register them, so tht we can automatically
invalidate cache whenever an upstream send a change notfication
for a resource of interest. This can be used as an [example](./pkg/storage/infra/blueprint.go) to implement storage accessor for other resources.

```go
blueprintAccessor, err := NewBlueprintAccessor(configPool, scheme)
if err != nil {
	return err
}

if err := registry.AddHubStorage((*infrapb.Blueprint)(nil), "blueprints", blueprintAccessor); err != nil {
	return err
}

if err := registry.AddCacheGroups("blueprints", blueprintsGroupName, blueprintsListGroupName); err != nil {
	return err
}
```

## Lineage Annotation for System Sync 
All the v3 resources that support System <-> Git sync must support storing Lineage information in `Metadata.Annotations` field with `rafay.dev/lineage` key. v3 resources implementation need not worry about the contents or usage of Lineage, it is populated and used during git -> system sync process to decide if a resource needs to be deleted from system. Expectations from hub resource implementation: 
1. `v3.Get`: Should return lineage only if `GetOptions.IncludeLineage` flag is set to `true`, otherwise filter-out lineage annotation before returning the resource. 
2. `v3.List`: Same as GET operation, check for `ListOptions.IncludeLineage` flag to decide if lineage information needs to be filtered out for each resource being returned.
3. `v3.Apply`:  
	a. If the resource is being created: Check for `Applyoptions.InlcudeLineage` flag and filter out Lineage annotation before saving the resource.  
	b. If the resource is being updated:  
		&emsp;i. `Applyoptions.IncludeLineage` is true: No extra steps needed  
		&emsp;ii. `Applyoptions.IncludeLineage` is false: Preserve old value for Lineage Annotation. 
4. Implement lineage interface in the newly added v3 object, e.g. https://github.com/RafaySystems/rafay-common/blob/master/proto/types/hub/eaaspb/resource_lineage.go 
```
package eaaspb

import (
	"github.com/RafaySystems/rafay-common/pkg/hub"
)

func (x *Resource) LineageFilter() {

}

var _ hub.LineageAnnotationFilter = (*Resource)(nil)
```

If the resource only supports v3 API then this logic can be implemented in Accessors, but if the resource also supports v2 then this logic has to be implemented in upstream services.

### InternalTriggerable for System Sync
All the v3 resources that support System <-> Git auto sync upon any change in system configuration should implement InternalTriggerable interface e.g. https://github.com/RafaySystems/rafay-common/blob/master/proto/types/hub/eaaspb/resource_trigger.go

```
package eaaspb

import "github.com/RafaySystems/rafay-common/pkg/hub"

func (x *Resource) TriggerInternal() {

}

var _ hub.InternalTriggerable = (*Resource)(nil)
```