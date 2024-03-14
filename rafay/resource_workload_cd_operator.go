package rafay

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/proto/types/hub/appspb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/integrationspb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/systempb"
	rctl_cluster "github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/config"
	rctl_project "github.com/RafaySystems/rctl/pkg/project"

	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/RafaySystems/rafay-common/pkg/hub/codec"
	hub_types "github.com/RafaySystems/rafay-common/pkg/hub/conversion/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Repo Credentials
type CDCredentials struct {
	Username   string `json:"username,omitempty"`   // the username to access the repository
	Password   string `json:"password,omitempty"`   // the password to access the repository
	PrivateKey string `json:"privateKey,omitempty"` // the private key to access the repository
	Token      string `json:"token,omitempty"`      // the token to access the repository
}

// HelmOptions
type HelmOptions struct {
	Atomic                   bool     `json:"atomic,omitempty"`
	Wait                     bool     `json:"wait,omitempty"`
	Force                    bool     `json:"force,omitempty"`
	NoHooks                  bool     `json:"noHooks,omitempty"`
	MaxHistory               int32    `json:"maxHistory,omitempty"`
	RenderSubChartNotes      bool     `json:"renderSubChartNotes,omitempty"`
	ResetValues              bool     `json:"resetValues,omitempty"`
	ReuseValues              bool     `json:"reuseValues,omitempty"`
	SetString                []string `json:"setString,omitempty"`
	SkipCRDs                 bool     `json:"skipCRDs,omitempty"`
	Timeout                  string   `json:"timeout,omitempty"`
	CleanUpOnFail            bool     `json:"cleanUpOnFail,omitempty"`
	Description              string   `json:"description,omitempty"`
	DisableOpenAPIValidation bool     `json:"disableOpenAPIValidation,omitempty"`
	KeepHistory              bool     `json:"keepHistory,omitempty"`
	WaitForJobs              bool     `json:"waitForJobs,omitempty"`
	WaitForUninstall         bool     `json:"waitForUninstall,omitempty"`
}

type Workload struct {
	Name               string            `json:"name,omitempty"`               // the name of the workload
	PathMatchPattern   string            `json:"pathMatchPattern,omitempty"`   // the path  pattern to extract project name from
	BasePath           string            `json:"basePath,omitempty"`           // the path  pattern to extract base chart from
	IncludeBaseValue   bool              `json:"includeBaseValue,omitempty"`   // include base value.yaml
	DeleteAction       string            `json:"enableDelete,omitempty"`       // delete the workload
	ClusterNames       string            `json:"clusterNames,omitempty"`       // the cluster names to deploy the workload
	PlacementLabels    map[string]string `json:"placementLabels,omitempty"`    // the placement labels for the clusters
	HelmChartName      string            `json:"helmChartName,omitempty"`      // the name of the helm chart
	HelmChartVersion   string            `json:"helmChartVersion,omitempty"`   // the version of the helm chart
	HelmOptions        *HelmOptions      `json:"helmChartOptions,omitempty"`   // the options for the helm chart
	ChartHelmRepoName  string            `json:"chartHelmRepoName,omitempty"`  // the name of the helm repo
	ChartGitRepoName   string            `json:"chartGitRepoName,omitempty"`   // the name of the git repo
	ChartGitRepoBranch string            `json:"chartGitRepoBranch,omitempty"` // the branch of the git repo
	ChartGitRepoPath   string            `json:"chartGitRepoPath,omitempty"`   // the path of the git repo
	ChartCatalogName   string            `json:"chartGitRepoPath,omitempty"`   // the name of the catalog to source the chart
}

// The config spec for the WorkloadCD resource
type WorkloadCDConfigSpec struct {
	Type                string                            `json:"type,omitempty"`                // type of the repository - not used for now
	RepoURL             string                            `json:"repourl,omitempty"`             // the url of the repository
	RepoBranch          string                            `json:"repobranch,omitempty"`          // the branch of the repository
	Credentials         *CDCredentials                    `json:"credentials,omitempty"`         // the credentials to access the repository
	Options             *integrationspb.RepositoryOptions `json:"options,omitempty"`             // the options for the repository
	Insecure            bool                              `json:"insecure,omitempty"`            // allow insecure connection
	RepositoryLocalPath string                            `json:"repositoryLocalPath,omitempty"` // the local path of the repository to clone
	Workloads           []*Workload                       `json:"workloads,omitempty"`           // the workloads to deploy
}

type WorkloadCDStatus struct {
	Project      string           `json:"project,omitempty"`      // the project of the workload
	Namespace    string           `json:"namespace,omitempty"`    // the namespace of the workload
	WorkloadName string           `json:"workloadName,omitempty"` // the name of the workload
	RepoFolder   string           `json:"repoFolder,omitempty"`   // the name of the workload
	Version      string           `json:"version,omitempty"`      // the version of the workload
	Status       *commonpb.Status `json:"status,omitempty"`       // the status of the workload
}
type WorkloadsUpsert struct {
	Project      string `json:"project,omitempty"`      // the project of the workload
	Namespace    string `json:"namespace,omitempty"`    // the namespace of the workload
	WorkloadName string `json:"workloadName,omitempty"` // the name of the workload
}
type WorkloadsDecommission struct {
	Project      string `json:"project,omitempty"`      // the project of the workload
	Namespace    string `json:"namespace,omitempty"`    // the namespace of the workload
	WorkloadName string `json:"workloadName,omitempty"` // the name of the workload
}

type WorkloadCDConfig struct {
	ApiVersion    string                   `json:"apiVersion,omitempty"`    // the api version of the resource
	Kind          string                   `json:"kind,omitempty"`          // the kind of the resource
	Metadata      *commonpb.Metadata       `json:"metadata,omitempty"`      // the metadata of the resource
	Spec          *WorkloadCDConfigSpec    `json:"spec,omitempty"`          // the specification of the resource
	Status        []*WorkloadCDStatus      `json:"status,omitempty"`        // the status of the resource
	Decommissions []*WorkloadsDecommission `json:"decommissions,omitempty"` // the status of the resource
	Upserts       []*WorkloadsUpsert       `json:"upserts,omitempty"`       // the status of the resource
}

const charset = "abcdefghijklmnopqrstuvwxyz"                                // 36 characters
var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano())) // a source to generate random numbers

var (
	hubYAMLCodec = codec.NewYAMLCodec(hub_types.DefaultScheme) // the codec to encode/decode hub resources
)

var (
	commitSHARegex = regexp.MustCompile("^[0-9A-Fa-f]{40}$")          // the regex to validate a commit SHA
	sshURLRegex    = regexp.MustCompile("^(ssh://)?([^/:]*?)@[^@]+$") // the regex to validate an SSH URL
	httpsURLRegex  = regexp.MustCompile("^(https://).*")              // the regex to validate an HTTPS URL
	httpURLRegex   = regexp.MustCompile("^(http://).*")               // the regex to validate an HTTP URL
)

var _dummyHandler = func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {} // a dummy handler to be used for routing

// guard is a channel that used to make sure that only N=10
// goroutines will run at a time
var guard = make(chan struct{}, 10)

// WorkloadCDRepositorySchema is the schema for the WorkloadCD resource
var WorkloadCDRepositorySchema = &schema.Resource{
	Description: "Workload CD Repository  definition",
	Schema: map[string]*schema.Schema{
		"always_run": &schema.Schema{ // always run the resource when this is set to time
			Description: "always run",
			Optional:    true,
			Type:        schema.TypeString,
		},
		"metadata": &schema.Schema{
			Description: "Metadata of the secret sealer resource",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"annotations": &schema.Schema{
					Description: "annotations of the resource",
					Elem:        &schema.Schema{Type: schema.TypeString},
					Optional:    true,
					Type:        schema.TypeMap,
				},
				"created_by": &schema.Schema{
					Description: "User who created this resource",
					Elem: &schema.Resource{Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Description: "Id of the Person",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"is_sso_user": &schema.Schema{
							Description: "Whether person is logged in using sso",
							Optional:    true,
							Type:        schema.TypeBool,
						},
						"username": &schema.Schema{
							Description: "Username fo the Person",
							Optional:    true,
							Type:        schema.TypeString,
						},
					}},
					MaxItems: 1,
					MinItems: 1,
					Optional: true,
					Type:     schema.TypeList,
				},
				"description": &schema.Schema{
					Description: "description of the resource",
					Optional:    true,
					Type:        schema.TypeString,
				},
				"display_name": &schema.Schema{
					Description: "Display Name of the resource",
					Optional:    true,
					Type:        schema.TypeString,
				},
				"labels": &schema.Schema{
					Description: "labels of the resource",
					Elem:        &schema.Schema{Type: schema.TypeString},
					Optional:    true,
					Type:        schema.TypeMap,
				},
				"modified_by": &schema.Schema{
					Description: "User who last modified this resource",
					Elem: &schema.Resource{Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Description: "Id of the Person",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"is_sso_user": &schema.Schema{
							Description: "Whether person is logged in using sso",
							Optional:    true,
							Type:        schema.TypeBool,
						},
						"username": &schema.Schema{
							Description: "Username fo the Person",
							Optional:    true,
							Type:        schema.TypeString,
						},
					}},
					MaxItems: 1,
					MinItems: 1,
					Optional: true,
					Type:     schema.TypeList,
				},
				"name": &schema.Schema{
					Description: "name of the resource",
					Optional:    true,
					Type:        schema.TypeString,
				},
				"project": &schema.Schema{
					Description: "Project of the resource",
					Optional:    true,
					Type:        schema.TypeString,
				},
			}},
			MaxItems: 1,
			MinItems: 1,
			Optional: true,
			Type:     schema.TypeList,
		},
		"spec": &schema.Schema{
			Description: "Specification of the repository resource",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"repo_local_path": &schema.Schema{
					Description: "repository local path",
					Optional:    true,
					Default:     "/tmp/apprepo",
					Type:        schema.TypeString,
				},
				"repo_url": &schema.Schema{
					Description: "repository repo_url",
					Required:    true,
					Type:        schema.TypeString,
				},
				"repo_branch": &schema.Schema{
					Description: "repository branch",
					Optional:    true,
					Type:        schema.TypeString,
				},
				"insecure": &schema.Schema{
					Description: "repository allow insecure connection",
					Optional:    true,
					Type:        schema.TypeBool,
				},
				"repo_type": &schema.Schema{
					Description: "repository type",
					Optional:    true,
					Type:        schema.TypeString,
				},
				"credentials": &schema.Schema{
					Description: "",
					Elem: &schema.Resource{Schema: map[string]*schema.Schema{
						"password": &schema.Schema{
							Description: "",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"private_key": &schema.Schema{
							Description: "",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"username": &schema.Schema{
							Description: "",
							Required:    true,
							Type:        schema.TypeString,
						},
						"token": &schema.Schema{
							Description: "",
							Optional:    true,
							Type:        schema.TypeString,
						},
					}},
					MaxItems: 1,
					MinItems: 1,
					Optional: true,
					Type:     schema.TypeList,
				},

				"workload": &schema.Schema{
					Description: "",
					Elem: &schema.Resource{Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Description: "workload name",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"path_match_pattern": &schema.Schema{
							Description: "project/namespace/workload name path match pattern",
							Required:    true,
							Type:        schema.TypeString,
						},
						"base_path": &schema.Schema{
							Description: "repository local path",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"include_base_value": &schema.Schema{
							Description: "include values from base path",
							Optional:    true,
							Type:        schema.TypeBool,
						},
						"cluster_names": &schema.Schema{
							Description: "cluster names ',' separated",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"placement_labels": &schema.Schema{
							Description: "placement labels of the cluster",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
							Type:        schema.TypeMap,
						},
						"delete_action": &schema.Schema{
							Description: "workload delete action",
							Optional:    true,
							Default:     "none",
							Type:        schema.TypeString,
						},
						"chart_helm_repo_name": &schema.Schema{
							Description: "rafay helm repo name to source chart",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"chart_git_repo_name": &schema.Schema{
							Description: "rafay git repo name to source chart",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"chart_catalog_name": &schema.Schema{
							Description: "rafay catalog name to source chart",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"chart_git_repo_branch": &schema.Schema{
							Description: "rafay git repo branch to source chart",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"chart_git_repo_path": &schema.Schema{
							Description: "rafay git repo path",
							Optional:    true,
							Default:     "/",
							Type:        schema.TypeString,
						},
						"helm_chart_name": &schema.Schema{
							Description: "helm chart name",
							Required:    true,
							Type:        schema.TypeString,
						},
						"helm_chart_version": &schema.Schema{
							Description: "helm chart version",
							Required:    true,
							Type:        schema.TypeString,
						},
						"helm_options": &schema.Schema{
							Description: "",
							Elem: &schema.Resource{Schema: map[string]*schema.Schema{
								"atomic": &schema.Schema{
									Description: "deploy Helm artifact with atomic flag",
									Optional:    true,
									Type:        schema.TypeBool,
								},
								"clean_up_on_fail": &schema.Schema{
									Description: "cleanup deployed resources when chart fails to deploy",
									Optional:    true,
									Type:        schema.TypeBool,
								},
								"description": &schema.Schema{
									Description: "custom description for the release",
									Optional:    true,
									Type:        schema.TypeString,
								},
								"disable_open_api_validation": &schema.Schema{
									Description: "disable OpenAPI validation while deploying the YAML",
									Optional:    true,
									Type:        schema.TypeBool,
								},
								"force": &schema.Schema{
									Description: "deploy YAML artifact with force flag",
									Optional:    true,
									Type:        schema.TypeBool,
								},
								"keep_history": &schema.Schema{
									Description: "keep release history after uninstalling",
									Optional:    true,
									Type:        schema.TypeBool,
								},
								"max_history": &schema.Schema{
									Description: "limit Helm artifact history",
									Optional:    true,
									Type:        schema.TypeInt,
								},
								"no_hooks": &schema.Schema{
									Description: "deploy Helm artifact without hooks",
									Optional:    true,
									Type:        schema.TypeBool,
								},
								"render_sub_chart_notes": &schema.Schema{
									Description: "render sub chart notes",
									Optional:    true,
									Type:        schema.TypeBool,
								},
								"reset_values": &schema.Schema{
									Description: "reset existing helm values",
									Optional:    true,
									Type:        schema.TypeBool,
								},
								"reuse_values": &schema.Schema{
									Description: "reuse existing values",
									Optional:    true,
									Type:        schema.TypeBool,
								},
								"set_string": &schema.Schema{
									Description: "pass custom helm values as key=value",
									Elem:        &schema.Schema{Type: schema.TypeString},
									Optional:    true,
									Type:        schema.TypeList,
								},
								"skip_crd": &schema.Schema{
									Description: "skip deploying crds",
									Optional:    true,
									Type:        schema.TypeBool,
								},
								"timeout": &schema.Schema{
									Description: "timeout for waiting for the resources to become ready",
									Optional:    true,
									Type:        schema.TypeString,
								},
								"wait": &schema.Schema{
									Description: "deploy Helm artifact with wait flag",
									Optional:    true,
									Type:        schema.TypeBool,
								},
								"wait_for_jobs": &schema.Schema{
									Description: "deploy Helm artifact with --wait-for-jobs flag",
									Optional:    true,
									Type:        schema.TypeBool,
								},
								"wait_for_uninstall": &schema.Schema{
									Description: "uninstall Helm artifact with --wait flag",
									Optional:    true,
									Type:        schema.TypeBool,
								},
							}},
							MaxItems: 1,
							MinItems: 1,
							Optional: true,
							Type:     schema.TypeList,
						},
					}},
					MaxItems: 0,
					MinItems: 0,
					Optional: true,
					Type:     schema.TypeList,
				},
			}},
			MaxItems: 1,
			MinItems: 1,
			Optional: true,
			Type:     schema.TypeList,
		},
		"workload_status": &schema.Schema{ // status of the resource get updated when the resource is created
			Description: "Status of the workload resource",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"project": &schema.Schema{
					Description: "Project of the resource",
					Optional:    true,
					Computed:    true,
					Type:        schema.TypeString,
				},
				"namespace": &schema.Schema{
					Description: "Namespace of the resource",
					Optional:    true,
					Computed:    true,
					Type:        schema.TypeString,
				},
				"workload_name": &schema.Schema{
					Description: "Workload Name of the resource",
					Optional:    true,
					Computed:    true,
					Type:        schema.TypeString,
				},
				"workload_version": &schema.Schema{
					Description: "Workload Name of the resource",
					Optional:    true,
					Computed:    true,
					Type:        schema.TypeString,
				},
				"repo_folder": &schema.Schema{
					Description: "repo path of the Workload resource",
					Optional:    true,
					Computed:    true,
					Type:        schema.TypeString,
				},
				"condition_status": &schema.Schema{
					Description: "Condition Status",
					Optional:    true,
					Computed:    true,
					Type:        schema.TypeInt,
				},
				"clusters": &schema.Schema{
					Description: "deployed clusters",
					Optional:    true,
					Computed:    true,
					Type:        schema.TypeString,
				},
				"condition_type": &schema.Schema{
					Description: "Status message",
					Optional:    true,
					Computed:    true,
					Type:        schema.TypeString,
				},
				"reason": &schema.Schema{
					Description: "Status message",
					Optional:    true,
					Computed:    true,
					Type:        schema.TypeString,
				},
			}},
			MaxItems: 0,
			MinItems: 0,
			Optional: true,
			Computed: true,
			ForceNew: true,
			Type:     schema.TypeList,
		},
		"workload_upserts": &schema.Schema{ // status of the resource get updated when the resource is created
			Description: "created/updated workload resources",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"project": &schema.Schema{
					Description: "Project of the resource",
					Optional:    true,
					Computed:    true,
					Type:        schema.TypeString,
				},
				"namespace": &schema.Schema{
					Description: "Namespace of the resource",
					Optional:    true,
					Computed:    true,
					Type:        schema.TypeString,
				},
				"workload_name": &schema.Schema{
					Description: "Workload Name of the resource",
					Optional:    true,
					Computed:    true,
					Type:        schema.TypeString,
				},
			}},
			MaxItems: 0,
			MinItems: 0,
			Optional: true,
			Computed: true,
			ForceNew: true,
			Type:     schema.TypeList,
		},
		"workload_decommissions": &schema.Schema{
			Description: "List of deleted/unpublished the workloads",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"project": &schema.Schema{
					Description: "Project of the resource",
					Optional:    true,
					Computed:    true,
					Type:        schema.TypeString,
				},
				"namespace": &schema.Schema{
					Description: "Namespace of the resource",
					Optional:    true,
					Computed:    true,
					Type:        schema.TypeString,
				},
				"workload_name": &schema.Schema{
					Description: "Workload Name of the resource",
					Optional:    true,
					Computed:    true,
					Type:        schema.TypeString,
				},
			}},
			MaxItems: 0,
			MinItems: 0,
			Optional: true,
			Computed: true,
			ForceNew: true,
			Type:     schema.TypeList,
		},
	},
}

// WorkloadCDRepositorySchema is the schema for the WorkloadCD resource
func resourceWorkloadCDOperator() *schema.Resource {
	modSchema := WorkloadCDRepositorySchema.Schema
	return &schema.Resource{
		CreateContext: resourceWorkloadCDOperatorCreate,
		ReadContext:   resourceWorkloadCDOperatorRead,
		UpdateContext: resourceWorkloadCDOperatorUpdate,
		DeleteContext: resourceWorkloadCDOperatorDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        modSchema,
	}
}

// IsHTTPSURL returns true if supplied URL is HTTPS URL
func IsHTTPSURL(url string) bool {
	return httpsURLRegex.MatchString(url)
}

// StringWithCharset returns a random string of length with the supplied charset
func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// RandomString returns a random string of length
func RandomString(length int) string {
	return StringWithCharset(length, charset)
}

func expandStatus(p []interface{}) []*WorkloadCDStatus {
	if len(p) == 0 || p[0] == nil {
		return []*WorkloadCDStatus{}
	}

	out := make([]*WorkloadCDStatus, len(p))

	for i := range p {
		obj := WorkloadCDStatus{}
		in := p[i].(map[string]interface{})

		if v, ok := in["project"].(string); ok && len(v) > 0 {
			obj.Project = v
		}

		if v, ok := in["namespace"].(string); ok && len(v) > 0 {
			obj.Namespace = v
		}

		if v, ok := in["workload_name"].(string); ok && len(v) > 0 {
			obj.WorkloadName = v
		}

		if v, ok := in["repo_folder"].(string); ok && len(v) > 0 {
			obj.RepoFolder = v
		}

		if v, ok := in["workload_version"].(string); ok && len(v) > 0 {
			obj.Version = v
		}

		out[i] = &obj

	}

	return out
}

func readWorkLoadStatus(in *schema.ResourceData) []*WorkloadCDStatus {
	if in == nil {
		return nil
	}

	if v, ok := in.Get("workload_status").([]interface{}); ok && len(v) > 0 {
		return expandStatus(v)
	}

	return nil
}

func resourceWorkloadCDOperatorCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resourceWorkloadCDOperator create starts")
	diags := resourceWorkloadCDOperatorUpsert(ctx, d, m)

	return diags
}

func resourceWorkloadCDOperatorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resourceWorkloadCDOperator read starts")
	var diags diag.Diagnostics

	// tflog := os.Getenv("TF_LOG")
	// if tflog == "TRACE" || tflog == "DEBUG" {
	// 	ctx = context.WithValue(ctx, "debug", "true")
	// }

	// meta := GetMetaData(d)
	// if meta == nil {
	// 	return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	// }

	// if d.State() != nil && d.State().ID != "" {
	// 	meta.Name = d.State().ID
	// }

	// ret := readWorkLoadStatus(d)
	// d.Set("status", ret)

	return diags
}

func resourceWorkloadCDOperatorUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resourceWorkloadCDOperator create starts")
	diags := resourceWorkloadCDOperatorUpsert(ctx, d, m)

	return diags
}

func resourceWorkloadCDOperatorDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resourceWorkloadCDOperator create starts")
	var diags diag.Diagnostics

	return diags
}

// resourceWorkloadCDOperatorUpsert create or update the WorkloadCD resource
func resourceWorkloadCDOperatorUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var wg, dwg, cwg sync.WaitGroup
	var mu sync.Mutex
	var golbalWorkloadList = appspb.WorkloadList{}

	log.Printf("resourceWorkloadCDOperator upsert starts")

	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	if d.State() != nil && d.State().ID != "" {
		n := GetMetaName(d)
		if n != "" && n != d.State().ID {
			log.Printf("metadata name change not supported")
			d.State().Tainted = true
			return diag.FromErr(fmt.Errorf("%s", "metadata name change not supported"))
		}
	}

	workloadCDConfig, err := expandWorkloadCDConfig(d)
	if err != nil {
		log.Printf("addon expandAddon error")
		return diag.FromErr(err)
	}

	cdConf := spew.Sprintf("%+v", workloadCDConfig)
	log.Println("expandWorkloadCDConfig  ", cdConf)
	var output []string
	var ssh_key_path string
	if workloadCDConfig.Spec.Credentials != nil && workloadCDConfig.Spec.Credentials.PrivateKey != "" {
		output, ssh_key_path, err = cloneRepoSSH(workloadCDConfig)
		if ssh_key_path != "" {
			defer os.Remove(ssh_key_path)
		}
	} else {
		output, err = cloneRepo(workloadCDConfig)
	}
	if err != nil {
		log.Println("cloneRepo error", err)
		return diag.FromErr(err)
	}
	log.Println("cloneRepo output", output)

	// Get all the projects
	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("checkProject client error", err)
		return diag.FromErr(err)
	}
	projectList, err := client.SystemV3().Project().List(ctx, options.ListOptions{})
	if err != nil {
		log.Println("resourceWorkloadCDOperatorUpsert failed to get projectList error", err)
		return diag.FromErr(err)
	}

	for _, pr := range projectList.Items {
		// as we loop through put an empty struct to channel guard.
		// If the channel is still empty, the process will continue.
		// Else, the process will be blocked until there are rooms in the channel to put the empty struct.
		guard <- struct{}{}
		wg.Add(1)
		go getProjectWorkloadList(ctx, pr, &golbalWorkloadList, &mu, &wg)
	}
	//wait for all the go routines to finish
	wg.Wait()

	log.Println("resourceWorkloadCDOperatorUpsertprocess Spec.Workloads", workloadCDConfig.Spec.Workloads)

	for _, workload := range workloadCDConfig.Spec.Workloads {
		log.Println("resourceWorkloadCDOperatorUpsert process delete workload", workload)

		folders, files, baseChart, baseValues, err := walkRepo(workloadCDConfig, workload)
		if err != nil {
			log.Println("getRepoFiles error", err)
			return diag.FromErr(err)
		}
		log.Println("resourceWorkloadCDOperatorUpsert ", "files", files)
		log.Println("resourceWorkloadCDOperatorUpsert ", "baseChart", baseChart)
		log.Println("resourceWorkloadCDOperatorUpsert ", "folders", folders)
		log.Println("resourceWorkloadCDOperatorUpsert ", "baseValues", baseValues)

		if workload.DeleteAction != "none" {
			dwg.Add(1)
			go processApplicationFoldersForDelete(ctx, workloadCDConfig, workload, baseChart, folders, &golbalWorkloadList, &dwg)
			time.Sleep(time.Duration(10) * time.Second)
		}
	}
	//wait for all the go routines to finish
	dwg.Wait()

	for _, workload := range workloadCDConfig.Spec.Workloads {
		log.Println("resourceWorkloadCDOperatorUpsert process create workload", workload)

		folders, files, baseChart, baseValues, err := walkRepo(workloadCDConfig, workload)
		if err != nil {
			log.Println("getRepoFiles error", err)
			return diag.FromErr(err)
		}
		log.Println("resourceWorkloadCDOperatorUpsert ", "files", files)
		log.Println("resourceWorkloadCDOperatorUpsert ", "baseChart", baseChart)
		log.Println("resourceWorkloadCDOperatorUpsert ", "folders", folders)
		log.Println("resourceWorkloadCDOperatorUpsert ", "baseValues", baseValues)

		cwg.Add(1)
		go processApplicationFolders(ctx, workloadCDConfig, workload, baseChart, baseValues, folders, &golbalWorkloadList, &cwg)
		time.Sleep(time.Duration(10) * time.Second)
	}
	//wait for all the go routines to finish
	cwg.Wait()

	if workloadCDConfig.Status != nil && len(workloadCDConfig.Status) > 0 {
		log.Println("workloadCDConfig.Status", workloadCDConfig.Status)
		v, ok := d.Get("workload_status").([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret := flattenWorkloadStatus(workloadCDConfig.Status, v)
		if ret == nil {
			log.Println("flattenWorkloadStatus returned nil")
		}
		err = d.Set("workload_status", ret)
		if err != nil {
			log.Println("failed to set status error ", err)
		}
	} else {
		log.Println("Set workload_status nil")
		d.Set("workload_status", nil)
	}

	if workloadCDConfig.Decommissions != nil && len(workloadCDConfig.Decommissions) > 0 {
		log.Println("workloadCDConfig.Decommissions", workloadCDConfig.Decommissions)
		v, ok := d.Get("workload_decommissions").([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret := flattenWorkloadDecommisions(workloadCDConfig.Decommissions, v)
		if ret == nil {
			log.Println("flattenWorkloadDecommisions returned nil")
		}
		err = d.Set("workload_decommissions", ret)
		if err != nil {
			log.Println("failed to set decommissions error ", err)
		}
		log.Println("flattenWorkloadDecommisions returned ret", ret)
	} else {
		log.Println("Set workload_decommissions nil")
		d.Set("workload_decommissions", nil)
	}

	if workloadCDConfig.Upserts != nil && len(workloadCDConfig.Upserts) > 0 {
		log.Println("workloadCDConfig.Upserts", workloadCDConfig.Upserts)
		v, ok := d.Get("workload_upserts").([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret := flattenWorkloadUpserts(workloadCDConfig.Upserts, v)
		if ret == nil {
			log.Println("flattenWorkloadUpserts returned nil")
		}
		err = d.Set("workload_upserts", ret)
		if err != nil {
			log.Println("failed to set upserts error ", err)
		}
	} else {
		log.Println("Set workload_upserts nil")
		d.Set("workload_upserts", nil)
	}

	d.SetId(workloadCDConfig.Metadata.Name)
	return diags
}

func getProjectWorkloadList(ctx context.Context, pr *systempb.Project, gWorkloadList *appspb.WorkloadList, mu *sync.Mutex, wg *sync.WaitGroup) error {
	defer func() {
		wg.Done()
		<-guard
	}()

	var tmpList = appspb.WorkloadList{}
	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("getProjectWorkloadList client error", err)
		return err
	}
	// Get all the workloads created by the operator

	// Get all the workloads in the project
	wList, err := client.AppsV3().Workload().List(ctx, options.ListOptions{
		Project: pr.Metadata.Name,
	})
	if err != nil {
		log.Println("getProjectWorkloadList failed to get workload List for project", pr.Metadata.Name, "error", err)
		return err
	}
	for _, w := range wList.Items {
		for k, _ := range w.Metadata.Labels {
			if k == "k8smgmt.io/helm-deployer-tfcd" {
				log.Println("found operator deployed workload", w.Metadata.Name, "project", w.Metadata.Project, "namespace", w.Spec.Namespace)
				tmpList.Items = append(tmpList.Items, w)
			}
		}
	}
	log.Println("getProjectWorkloadList pr.Metadata.Name", pr.Metadata.Name, "tmpList", tmpList.Items)
	mu.Lock()
	gWorkloadList.Items = append(gWorkloadList.Items, tmpList.Items...)
	mu.Unlock()
	return nil
}

func flattenWorkloadStatus(input []*WorkloadCDStatus, p []interface{}) []interface{} {
	log.Println("flattenWorkloadStatus")
	if input == nil {
		return nil
	}

	log.Println("flattenWorkloadStatus len input", len(input))

	out := make([]interface{}, len(input))
	for i, in := range input {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Namespace) > 0 {
			obj["namespace"] = in.Namespace
		}

		if len(in.Project) > 0 {
			obj["project"] = in.Project
		}

		if len(in.WorkloadName) > 0 {
			obj["workload_name"] = in.WorkloadName
		}

		if len(in.Version) > 0 {
			obj["workload_version"] = in.Version
		}

		if len(in.RepoFolder) > 0 {
			obj["repo_folder"] = in.RepoFolder
		}

		if len(in.Status.ConditionType) > 0 {
			obj["condition_type"] = in.Status.ConditionType
		}

		obj["condition_status"] = in.Status.ConditionStatus

		if len(in.Status.DeployedClusters) > 0 {
			obj["clusters"] = strings.Join(in.Status.DeployedClusters[:], ",")
		}

		if len(in.Status.Reason) > 0 {
			obj["reason"] = in.Status.Reason
		}

		out[i] = &obj
	}

	return out
}

func flattenWorkloadDecommisions(input []*WorkloadsDecommission, p []interface{}) []interface{} {
	log.Println("flattenWorkloadDecommisions")
	if input == nil {
		return nil
	}

	log.Println("flattenWorkloadDecommisions len input", len(input))

	out := make([]interface{}, len(input))
	for i, in := range input {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}
		log.Println("flattenWorkloadDecommisions in", in)

		if len(in.Namespace) > 0 {
			obj["namespace"] = in.Namespace
		}

		if len(in.Project) > 0 {
			obj["project"] = in.Project
		}

		if len(in.WorkloadName) > 0 {
			obj["workload_name"] = in.WorkloadName
		}

		out[i] = &obj
	}

	return out
}

func flattenWorkloadUpserts(input []*WorkloadsUpsert, p []interface{}) []interface{} {
	log.Println("flattenWorkloadUpserts")
	if input == nil {
		return nil
	}

	log.Println("flattenWorkloadUpserts len input", len(input))

	out := make([]interface{}, len(input))
	for i, in := range input {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Namespace) > 0 {
			obj["namespace"] = in.Namespace
		}

		if len(in.Project) > 0 {
			obj["project"] = in.Project
		}

		if len(in.WorkloadName) > 0 {
			obj["workload_name"] = in.WorkloadName
		}

		out[i] = &obj
	}

	return out
}

func expandWorkloadCDConfig(in *schema.ResourceData) (*WorkloadCDConfig, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand WorkloadCD empty input")
	}
	obj := &WorkloadCDConfig{}

	if v, ok := in.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok {
		objSpec, err := expandWorkloadCDConfigSpec(v)
		if err != nil {
			return nil, err
		}
		// XXX Debug
		cdSpec := spew.Sprintf("%+v", objSpec)
		log.Println("expandWorkloadCDConfig  ", cdSpec)
		obj.Spec = objSpec
	}

	obj.ApiVersion = "apps.k8smgmt.io/v3"
	obj.Kind = "Workload"
	return obj, nil
}

func expandWorkloadCDConfigSpec(p []interface{}) (*WorkloadCDConfigSpec, error) {
	obj := &WorkloadCDConfigSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandAddonSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["repo_type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	if v, ok := in["repo_url"].(string); ok && len(v) > 0 {
		obj.RepoURL = v
	}

	if v, ok := in["repo_branch"].(string); ok && len(v) > 0 {
		obj.RepoBranch = v
	}

	if v, ok := in["insecure"].(bool); ok {
		obj.Insecure = v
	}

	if v, ok := in["repo_local_path"].(string); ok && len(v) > 0 {
		abs, err := filepath.Abs(v)
		if err == nil {
			obj.RepositoryLocalPath = abs
		} else {
			obj.RepositoryLocalPath = "/tmp/apprepo"
		}
	}

	if v, ok := in["credentials"].([]interface{}); ok && len(v) > 0 {
		// XXX Debug
		objCreds, err := expandWorkloadCDCredentials(v)
		if err != nil {
			return nil, err
		}
		// XXX Debug
		creds := spew.Sprintf("%+v", objCreds)
		log.Println("expandWorkloadCDConfigSpec expand ", creds)
		obj.Credentials = objCreds
	}

	if v, ok := in["workload"].([]interface{}); ok && len(v) > 0 {
		obj.Workloads = expandWorkloads(v)
	}

	return obj, nil
}

func expandWorkloads(p []interface{}) []*Workload {
	if len(p) == 0 || p[0] == nil {
		return []*Workload{}
	}

	out := make([]*Workload, len(p))

	for i := range p {
		obj := Workload{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["cluster_names"].(string); ok && len(v) > 0 {
			obj.ClusterNames = v
		}

		if v, ok := in["delete_action"].(string); ok && len(v) > 0 {
			obj.DeleteAction = strings.ToLower(v)
		}

		if v, ok := in["placement_labels"].(map[string]interface{}); ok && len(v) > 0 {
			obj.PlacementLabels = toMapString(v)
		} else {
			obj.PlacementLabels = nil
		}

		if v, ok := in["path_match_pattern"].(string); ok && len(v) > 0 {
			obj.PathMatchPattern = v
		}

		if v, ok := in["base_path"].(string); ok && len(v) > 0 {
			obj.BasePath = v
		}

		if v, ok := in["include_base_value"].(bool); ok {
			obj.IncludeBaseValue = v
		}

		if v, ok := in["chart_catalog_name"].(string); ok && len(v) > 0 {
			obj.ChartCatalogName = v
		}

		if v, ok := in["chart_helm_repo_name"].(string); ok && len(v) > 0 {
			obj.ChartHelmRepoName = v
		}

		if v, ok := in["chart_git_repo_name"].(string); ok && len(v) > 0 {
			obj.ChartGitRepoName = v
		}

		if v, ok := in["chart_git_repo_branch"].(string); ok && len(v) > 0 {
			obj.ChartGitRepoBranch = v
		}

		if v, ok := in["chart_git_repo_path"].(string); ok && len(v) > 0 {
			obj.ChartGitRepoPath = v
		}

		if v, ok := in["helm_chart_name"].(string); ok && len(v) > 0 {
			obj.HelmChartName = v
		}

		if v, ok := in["helm_chart_version"].(string); ok && len(v) > 0 {
			obj.HelmChartVersion = v
		}

		if v, ok := in["helm_options"].([]interface{}); ok && len(v) > 0 {
			obj.HelmOptions = expandHelmOptions(v)
		}

		out[i] = &obj
	}

	return out
}

func expandHelmOptions(p []interface{}) *HelmOptions {
	obj := &HelmOptions{}
	if len(p) == 0 || p[0] == nil {
		log.Println("expandHelmOptions empty input")
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["atomic"].(bool); ok {
		obj.Atomic = v
	}

	if v, ok := in["clean_up_on_fail"].(bool); ok {
		obj.CleanUpOnFail = v
	}

	if v, ok := in["description"].(string); ok && len(v) > 0 {
		obj.Description = v
	}

	if v, ok := in["disable_open_api_validation"].(bool); ok {
		obj.DisableOpenAPIValidation = v
	}

	if v, ok := in["force"].(bool); ok {
		obj.Force = v
	}

	if v, ok := in["keep_history"].(bool); ok {
		obj.KeepHistory = v
	}

	if v, ok := in["max_history"].(int32); ok {
		obj.MaxHistory = v
	}

	if v, ok := in["no_hooks"].(bool); ok {
		obj.NoHooks = v
	}

	if v, ok := in["render_sub_chart_notes"].(bool); ok {
		obj.RenderSubChartNotes = v
	}

	if v, ok := in["reset_values"].(bool); ok {
		obj.ResetValues = v
	}

	if v, ok := in["reuse_values"].(bool); ok {
		obj.ReuseValues = v
	}

	if v, ok := in["set_string"].([]interface{}); ok && len(v) > 0 {
		obj.SetString = toArrayString(v)
	}

	if v, ok := in["skip_crd"].(bool); ok {
		obj.SkipCRDs = v
	}

	if v, ok := in["timeout"].(string); ok && len(v) > 0 {
		obj.Timeout = v
	}

	if v, ok := in["wait"].(bool); ok {
		obj.Wait = v
	}

	if v, ok := in["wait_for_jobs"].(bool); ok {
		obj.WaitForJobs = v
	}

	if v, ok := in["wait_for_uninstall"].(bool); ok {
		obj.WaitForUninstall = v
	}

	return obj
}

func expandWorkloadCDCredentials(p []interface{}) (*CDCredentials, error) {
	obj := &CDCredentials{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandWorkloadCDCredentials empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["password"].(string); ok && len(v) > 0 {
		obj.Password = v
	}

	if v, ok := in["private_key"].(string); ok && len(v) > 0 {
		obj.PrivateKey = v
	}

	if v, ok := in["token"].(string); ok && len(v) > 0 {
		obj.Token = v
	}

	if v, ok := in["username"].(string); ok && len(v) > 0 {
		obj.Username = v
	}

	return obj, nil
}

// runGitCmd is a convenience function to run a command in a given directory and return its output
func runGitCmd(workloadCdCfg *WorkloadCDConfig, cmdDir string, senstive bool, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	return runCmdOutput(workloadCdCfg, cmdDir, cmd, senstive)
}

// runCmd is a convenience function to run a command in a given directory and return its output
func runCmd(workloadCdCfg *WorkloadCDConfig, cmnd, cmdDir string, senstive bool, args ...string) (string, error) {
	cmd := exec.Command(cmnd, args...)
	return runCmdOutput(workloadCdCfg, cmdDir, cmd, senstive)
}

func runCmdOutput(workloadCdCfg *WorkloadCDConfig, cmdDir string, cmd *exec.Cmd, senstive bool) (string, error) {
	cmd.Dir = cmdDir
	cmd.Env = append(os.Environ(), cmd.Env...)
	// Set $HOME to nowhere, so we can be execute Git regardless of any external
	// authentication keys (e.g. in ~/.ssh) -- this is especially important for
	// running tests on local machines and/or CircleCI.
	cmd.Env = append(cmd.Env, "HOME=/dev/null")
	// Skip LFS for most Git operations except when explicitly requested
	cmd.Env = append(cmd.Env, "GIT_LFS_SKIP_SMUDGE=1")
	// Disable Git terminal prompts in case we're running with a tty
	cmd.Env = append(cmd.Env, "GIT_TERMINAL_PROMPT=false")

	// For HTTPS repositories, we need to consider insecure repositories as well
	// as custom CA bundles from the cert database.
	if IsHTTPSURL(workloadCdCfg.Spec.RepoURL) {
		if workloadCdCfg.Spec.Insecure {
			cmd.Env = append(cmd.Env, "GIT_SSL_NO_VERIFY=true")
		}
	}

	opts := ExecRunOpts{
		TimeoutBehavior: TimeoutBehavior{
			Signal:     syscall.SIGTERM,
			ShouldWait: true,
		},
	}

	out, err := RunWithExecRunOpts(cmd, opts)
	if !senstive {
		log.Println("runCmdOutput", cmd, "out", out, "err", err)
	}
	return out, err
}

func cloneRepoSSH(workloadCdCfg *WorkloadCDConfig) ([]string, string, error) {
	var out string
	var sshKey string

	log.Printf("cloneRepoSSH starts")
	repo_url := workloadCdCfg.Spec.RepoURL
	//repo_branch := workloadCdCfg.Spec.RepoBranch
	if workloadCdCfg.Spec.Credentials != nil {
		sshKey = workloadCdCfg.Spec.Credentials.PrivateKey
	}

	path := workloadCdCfg.Spec.RepositoryLocalPath
	// remove the local repo if it exists
	runCmd(workloadCdCfg, "rm", ".", false, "-rf", path)
	f, err := os.CreateTemp("", "tfcd_ssh_key")
	if err != nil {
		log.Println("failed to create file", err)
		return nil, "", err
	}
	f.Write([]byte(sshKey))
	f.Close()
	ssh_key_path := f.Name()

	time.Sleep(5 * time.Second)
	// if the repo doesn't exist, we need to clone it
	// git clone -c "core.sshCommand=ssh -i ssh_key_path" --branch <branchname> git@github.com:stephan-rafay/test-tfcd.git <path>
	sshCmd := "core.sshCommand=ssh -i " + ssh_key_path
	if workloadCdCfg.Spec.RepoBranch != "" {
		out, err = runGitCmd(workloadCdCfg, ".", true, "clone", "-c", sshCmd, "--branch", workloadCdCfg.Spec.RepoBranch, repo_url, path)
	} else {
		out, err = runGitCmd(workloadCdCfg, ".", true, "clone", "-c", sshCmd, repo_url, path)
	}
	if err != nil {
		return nil, ssh_key_path, fmt.Errorf("failed to clone repo: error %+v out %s", err, out)
	}

	// remove last element, which is blank regardless of whether we're using nullbyte or newline
	ss := strings.Split(out, "\000")
	return ss[:len(ss)-1], ssh_key_path, nil
}

func cloneRepo(workloadCdCfg *WorkloadCDConfig) ([]string, error) {
	var out string
	var user, password, token string

	log.Printf("cloneRepo starts")
	repo_url := workloadCdCfg.Spec.RepoURL
	//repo_branch := workloadCdCfg.Spec.RepoBranch
	if workloadCdCfg.Spec.Credentials != nil {
		user = workloadCdCfg.Spec.Credentials.Username
		password = workloadCdCfg.Spec.Credentials.Password
		token = workloadCdCfg.Spec.Credentials.Token
	}
	path := workloadCdCfg.Spec.RepositoryLocalPath

	//git -C /tmp/apprepo pull
	out, err := runGitCmd(workloadCdCfg, ".", false, "-C", path, "pull")
	if err == nil {
		if workloadCdCfg.Spec.RepoBranch != "" {
			_, err := runGitCmd(workloadCdCfg, path, false, "checkout", workloadCdCfg.Spec.RepoBranch)
			if err != nil {
				log.Println("failed to checkout branch error", err)
			}
		}
	}
	if err != nil {
		var url string

		// remove the local repo if it exists
		runCmd(workloadCdCfg, "rm", ".", false, "-rf", path)
		// if the repo doesn't exist, we need to clone it
		// git clone --branch <branchname> https://stephan-rafay:api-key@url <path>
		if strings.Contains(repo_url, "https://") {
			strs := strings.Split(repo_url, "https://")
			if password != "" {
				url = fmt.Sprintf("https://%s:%s@%s", user, password, strs[1])
			} else if token != "" {
				url = fmt.Sprintf("https://%s:%s@%s", user, token, strs[1])
			} else {
				url = repo_url
			}
		} else if strings.Contains(repo_url, "http://") {
			strs := strings.Split(repo_url, "http://")
			if password != "" {
				url = fmt.Sprintf("http://%s:%s@%s", user, password, strs[1])
			} else if token != "" {
				url = fmt.Sprintf("http://%s:%s@%s", user, token, strs[1])
			} else {
				url = repo_url
			}
		}
		if workloadCdCfg.Spec.RepoBranch != "" {
			out, err = runGitCmd(workloadCdCfg, ".", true, "clone", "--branch", workloadCdCfg.Spec.RepoBranch, url, path)
		} else {
			out, err = runGitCmd(workloadCdCfg, ".", true, "clone", url, path)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to clone repo: %s", out)
		}
	}
	// remove last element, which is blank regardless of whether we're using nullbyte or newline
	ss := strings.Split(out, "\000")
	return ss[:len(ss)-1], nil
}

// walkRepo walks the repository and returns all files and folders
// it also returns the base chart if it exists
// it returns an error if the walk fails
func walkRepo(cfg *WorkloadCDConfig, wl *Workload) ([]string, []string, string, []string, error) {
	var files []string
	var folders []string
	var baseChart string
	var baseValues []string

	root := cfg.Spec.RepositoryLocalPath
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".git") {
			// ignore .git folder
			return nil
		}
		if !info.IsDir() {
			abs, err := filepath.Abs(path)
			if err == nil {
				if strings.Contains(abs, wl.Name) {
					files = append(files, abs)
				}
			} else {
				log.Println("failed to get absolute path for files", err)
			}
			if wl.BasePath != "" {
				if strings.Contains(path, wl.BasePath) && strings.HasSuffix(path, ".tgz") {
					abs, err := filepath.Abs(path)
					if err == nil {
						baseChart = abs
					} else {
						log.Println("failed to get absolute path for baseChart", err)
					}
				}
				if strings.Contains(path, wl.BasePath) &&
					(strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
					abs, err := filepath.Abs(path)
					if err == nil {
						baseValues = append(baseValues, abs)
					} else {
						log.Println("failed to get absolute path for baseValues", err)
					}
				}
			}
		} else {
			// check if the folder is a leaf
			var isLeaf = true
			filepath.Walk(path, func(path1 string, info1 os.FileInfo, err1 error) error {
				if path != path1 && info1.IsDir() {
					isLeaf = false
					return nil
				}
				return nil
			})
			if isLeaf {
				abs, err := filepath.Abs(path)
				if err == nil {
					if strings.Contains(abs, wl.Name) {
						folders = append(folders, abs)
					}
				} else {
					log.Println("failed to get absolute path for folders", err)
				}
			}

		}
		return nil
	})
	return folders, files, baseChart, baseValues, err
}

func processApplicationFoldersForDelete(ctx context.Context, cfg *WorkloadCDConfig, workload *Workload, baseChart string, folders []string, gWorkloadList *appspb.WorkloadList, dwg *sync.WaitGroup) error {
	var wg sync.WaitGroup
	var wrkPrunedList appspb.WorkloadList

	defer dwg.Done()
	for _, folder := range folders {
		// prune workload list
		var project, namespace, workloadName string
		var chartPath string

		projectCheck := httprouter.New()
		pattern := strings.TrimPrefix(strings.TrimSuffix(cfg.Spec.RepositoryLocalPath, "/"), ".") + workload.PathMatchPattern
		log.Println("Delete folder:", folder, "PathMatchPattern", pattern)
		projectCheck.Handle("POST", pattern, _dummyHandler)
		h, p, _ := projectCheck.Lookup("POST", folder)
		log.Println("h:", h)
		if h != nil {
			// got a hit for URL
			project = p.ByName("project")
			log.Println("project:", project)

			namespace = p.ByName("namespace")
			log.Println("namespace:", namespace)

			workloadName = p.ByName("workload")
			log.Println("workload:", workloadName)
		}

		if workloadName != workload.Name {
			// not interested in this workload
			continue
		}

		// get the chart in the folder
		chartPath, _ = getChartInFolder(folder)
		if chartPath == "" && baseChart != "" {
			// get chart from baseChart
			chartPath = baseChart
		}
		valuePaths, _ := getValuesInFolder(folder)

		log.Println("prepare pruned list", workloadName, " folder:", folder, "chartPath", chartPath, "valuePaths", valuePaths)
		for _, w := range gWorkloadList.Items {
			if w.Metadata.Name == workload.Name && w.Spec.Namespace == namespace && w.Metadata.Project == project {
				if (chartPath != "" ||
					workload.ChartCatalogName != "" ||
					workload.ChartHelmRepoName != "" ||
					workload.ChartGitRepoName != "") && len(valuePaths) > 0 {
					wrkPrunedList.Items = append(wrkPrunedList.Items, w)
				}
			}
		}
	}

	for _, w := range gWorkloadList.Items {
		if w.Metadata.Name != workload.Name {
			// not interested in this workload
			continue
		}
		log.Println("find app to delete", w.Metadata.Project, w.Metadata.Name, w.Spec.Namespace)
		found := false
		for _, pw := range wrkPrunedList.Items {
			if w.Metadata.Name == pw.Metadata.Name &&
				w.Spec.Namespace == pw.Spec.Namespace &&
				w.Metadata.Project == pw.Metadata.Project {
				found = true
				break
			}
		}
		if !found {
			// delete application
			wg.Add(1)
			log.Println("deleteApplication", w.Metadata.Project, w.Metadata.Name)
			go deleteApplication(ctx, cfg, workload, w.Metadata.Project, w.Spec.Namespace, w.Metadata.Name, &wg)
		}
	}

	wg.Wait()
	return nil
}

func deleteApplication(ctx context.Context, cfg *WorkloadCDConfig, workload *Workload, project, namespace, workloadName string, wg *sync.WaitGroup) error {
	defer wg.Done()

	resp, err := rctl_project.GetProjectByName(project)
	if err != nil {
		log.Println("project does not exist ", "error", err)
		return err
	}
	pr, err := rctl_project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		log.Println("project does not exist ", "error", err)
		return err
	}

	auth := config.GetConfig().GetAppAuthProfile()

	decommission := WorkloadsDecommission{}
	decommission.Project = project
	decommission.Namespace = namespace
	decommission.WorkloadName = workloadName

	if workload.DeleteAction == "delete" {
		uri := fmt.Sprintf("/v2/config/project/%s/workload/%s", pr.ID, workloadName)
		println("delete uri", uri)
		_, err := auth.AuthAndRequest(uri, "DELETE", nil)
		if err != nil {
			log.Println("delete workload uri", uri, "error", err)
			return err
		}
		cfg.Decommissions = append(cfg.Decommissions, &decommission)
	} else if workload.DeleteAction == "unpublish" {

		uri := fmt.Sprintf("/v2/config/project/%s/workload/%s/unpublish", pr.ID, workloadName)
		println("unpublish uri", uri)
		_, err := auth.AuthAndRequest(uri, "POST", nil)
		if err != nil {
			log.Println("unpublish workload uri", uri, "error", err)
			return err
		}
		cfg.Decommissions = append(cfg.Decommissions, &decommission)
	}
	return nil
}

func processApplicationFolders(ctx context.Context, cfg *WorkloadCDConfig, workload *Workload, baseChart string, baseValues, folders []string, gWorkloadList *appspb.WorkloadList, cwg *sync.WaitGroup) error {
	var chartPath string
	var wg sync.WaitGroup
	defer cwg.Done()

	for _, folder := range folders {
		var project, namespace, workloadName string
		var valuePaths []string
		// process folder and create application
		chartPath = ""

		projectCheck := httprouter.New()
		pattern := strings.TrimPrefix(strings.TrimSuffix(cfg.Spec.RepositoryLocalPath, "/"), ".") + workload.PathMatchPattern
		log.Println("folder:", folder, "PathMatchPattern", pattern)
		projectCheck.Handle("POST", pattern, _dummyHandler)
		h, p, _ := projectCheck.Lookup("POST", folder)
		log.Println("h:", h)

		if h != nil {
			// got a hit for URL
			project = p.ByName("project")
			log.Println("project:", project)

			namespace = p.ByName("namespace")
			log.Println("namespace:", namespace)

			workloadName = p.ByName("workload")
			log.Println("workloadName:", workloadName)

		}

		if project == "" || namespace == "" || workloadName == "" {
			log.Println("createApplication: project, namespace or workload is empty ignore folder", folder)
			continue
		}
		if workloadName != workload.Name {
			// not interested in this workload
			continue
		}

		// get the chart in the folder
		chartPath, _ = getChartInFolder(folder)
		if chartPath == "" && baseChart != "" {
			chartPath = baseChart
		}

		if workload.IncludeBaseValue {
			log.Println("processApplicationFolders include base values", baseValues)

			// add base values
			if len(baseValues) > 0 {
				valuePaths = append(valuePaths, baseValues...)
			}
		}

		vPaths, err := getValuesInFolder(folder)
		if err == nil && len(vPaths) > 0 {
			valuePaths = append(valuePaths, vPaths...)
		}

		// create application
		if (chartPath != "" ||
			workload.ChartCatalogName != "" ||
			workload.ChartHelmRepoName != "" ||
			workload.ChartGitRepoName != "") && len(valuePaths) > 0 {
			wg.Add(1)
			go createApplication(ctx, cfg, workload, folder, project, namespace, workload.Name, chartPath, valuePaths, &wg)
			time.Sleep(time.Duration(5) * time.Second)
		} else {
			log.Println("processApplicationFolders ignore folder ", folder, "  chartPath or valuePaths (or) catalog (or) helm-repo (or) gitrepo is empty")
		}
	}
	wg.Wait()
	return nil
}

func getChartInFolder(folder string) (string, error) {
	var chartPath string
	root := folder
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".tgz") {
			chartPath = path
			return nil
		}
		return nil
	})
	return chartPath, err
}

func getValuesInFolder(folder string) ([]string, error) {
	var valuePaths []string
	root := folder
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".yaml") || strings.Contains(path, ".yml") {
			valuePaths = append(valuePaths, path)
		}
		return nil
	})
	return valuePaths, err
}

func pruneRepolocalPath(localPtah, path string) string {
	lp1 := strings.TrimPrefix(localPtah, "./")
	if strings.HasPrefix(path, lp1) {
		return strings.TrimPrefix(path, lp1+"/")
	}
	return path
}

/*
Example Workload Spec

apiVersion: apps.k8smgmt.io/v3
kind: Workload
metadata:

	name: {{ .WorkloadName }}
	project: {{ .Project }}

spec:

	artifact:
	  artifact:
	    chartPath:
	      name: {{ .ChartPath }}
	    valuesPaths:
	    {{ .ValuesPaths }}
	  options:
	    maxHistory: 10
	    timeout: 5m0s
	  type: Helm
	namespace: {{ .Namespace }}
	placement:
	  selector: rafay.dev/clusterName in ({{ .ClusterNames }})
	  labels:
	    - key: {{ .LabelKey }}
		  value: {{ .LabelValue }}
	version: {{ .Version }}
*/

// getWorkLoadSpec returns the Workload spec
func getWorkLoadSpec(cfg *WorkloadCDConfig, workload *Workload, project, namespace, workloadName, chartPath, clusterNames, version string, valuePaths []string) string {
	var vPth string
	var spec string

	for _, valuePath := range valuePaths {
		vPth += "      - name: file://" + valuePath + "\n"
	}
	spec += "apiVersion: apps.k8smgmt.io/v3\n"
	spec += "kind: Workload\n"
	spec += "metadata:\n"
	spec += "  name: " + workloadName + "\n"
	spec += "  project: " + project + "\n"
	spec += "  labels:\n"
	spec += "    k8smgmt.io/helm-deployer-tfcd: cd-operator\n"
	spec += "spec:\n"
	spec += "  artifact:\n"
	spec += "    artifact:\n"
	if chartPath != "" {
		spec += "      chartPath:\n"
		spec += "        name: file://" + chartPath + "\n"
	} else if workload.ChartCatalogName != "" {
		spec += "      catalog: " + workload.ChartCatalogName + "\n"
	} else if workload.ChartHelmRepoName != "" {
		spec += "      repository: " + workload.ChartHelmRepoName + "\n"
	} else if workload.ChartGitRepoName != "" {
		spec += "      repository: " + workload.ChartGitRepoName + "\n"
		spec += "      revision: " + workload.ChartGitRepoBranch + "\n"
		spec += "      chartPath:\n"
		spec += "        name: " + workload.ChartGitRepoPath + "\n"

	}
	if workload.ChartGitRepoName == "" {
		if workload.HelmChartName != "" {
			spec += "      chartName: " + workload.HelmChartName + "\n"
		}
		if workload.HelmChartVersion != "" {
			spec += "      chartVersion: " + workload.HelmChartVersion + "\n"
		}
	}
	spec += "      valuesPaths:\n"
	spec += vPth
	spec += "    options:\n"
	spec += "      maxHistory: 10\n"
	spec += "      timeout: 5m0s\n"
	spec += "    type: Helm\n"
	spec += "  namespace: " + namespace + "\n"
	spec += "  placement:\n"
	if clusterNames != "" {
		spec += "    selector: rafay.dev/clusterName in (" + clusterNames + ")\n"
	}
	if len(workload.PlacementLabels) > 0 {
		spec += "    labels:\n"
		for k, v := range workload.PlacementLabels {
			spec += "      - key: " + k + "\n"
			if v != "" {
				spec += "        value: " + v + "\n"
			}
		}
	}

	spec += "  version: " + version + "\n"

	return spec

}

func createApplication(ctx context.Context, cfg *WorkloadCDConfig, workload *Workload, folder, project, namespace, workloadName, chartPath string, valuePaths []string, wg *sync.WaitGroup) error {
	// create application
	var clusterNames []string
	var chartVersion string
	var valueVersion string
	var workloadVersion string
	defer wg.Done()

	// check if project exist
	_, clusterList, err := checkProject(ctx, project)
	if err != nil {
		log.Println("createApplication: checkProject error", err)
		status := WorkloadCDStatus{}
		status.RepoFolder = folder
		status.Project = project
		status.Namespace = namespace
		status.WorkloadName = workload.Name
		status.Status.ConditionType = "Failed"
		status.Status.Reason = err.Error()
		cfg.Status = append(cfg.Status, &status)
		return err
	}

	if workload.ClusterNames == "" && len(workload.PlacementLabels) <= 0 {
		// get cluster names from clusterList in the project
		if len(clusterList) <= 0 {
			err = fmt.Errorf("createApplication: no clusters found for project %s", project)
			log.Println(err)
			status := WorkloadCDStatus{}
			status.RepoFolder = folder
			status.Project = project
			status.Namespace = namespace
			status.WorkloadName = workload.Name
			status.Status = &commonpb.Status{}
			status.Status.ConditionType = "Failed"
			status.Status.Reason = err.Error()
			cfg.Status = append(cfg.Status, &status)
			return err
		}
		// get cluster names from clusterList in the project
		clusterNames = append(clusterNames, clusterList...)
	}

	if workload.ClusterNames != "" {
		clusterNames = append(clusterNames, workload.ClusterNames)
	}

	if chartPath != "" {
		// get chartPath version
		// git log -n1 --oneline --pretty=format:%H
		trimPath := pruneRepolocalPath(cfg.Spec.RepositoryLocalPath, chartPath)
		out, err := runGitCmd(cfg, cfg.Spec.RepositoryLocalPath, false, "log", "-n1", "--oneline", "--pretty=format:%H", trimPath)
		if err != nil {
			log.Println("failed to runGitCmd ", err, "trimPath", trimPath)
			chartVersion = RandomString(7)
		} else {
			chartVersion = out[:7]
		}
	} else if workload.ChartCatalogName != "" {
		hashVar := sha256.New()
		hashVar.Write([]byte(workload.ChartCatalogName + workload.HelmChartName + workload.HelmChartVersion))
		bs := hashVar.Sum(nil)
		chartVersion = fmt.Sprintf("%x", bs)
	} else if workload.ChartHelmRepoName != "" {
		hashVar := sha256.New()
		hashVar.Write([]byte(workload.ChartHelmRepoName + workload.HelmChartName + workload.HelmChartVersion))
		bs := hashVar.Sum(nil)
		chartVersion = fmt.Sprintf("%x", bs)
	} else if workload.ChartGitRepoName != "" {
		hashVar := sha256.New()
		hashVar.Write([]byte(workload.ChartGitRepoName + workload.ChartGitRepoBranch + workload.ChartGitRepoPath + workload.HelmChartName + workload.HelmChartVersion))
		bs := hashVar.Sum(nil)
		chartVersion = fmt.Sprintf("%x", bs)
	}
	// get valuePath version
	for _, valuePath := range valuePaths {
		trimPath := pruneRepolocalPath(cfg.Spec.RepositoryLocalPath, valuePath)
		out, err := runGitCmd(cfg, cfg.Spec.RepositoryLocalPath, false, "log", "-n1", "--oneline", "--pretty=format:%H", trimPath)
		if err != nil {
			log.Println("failed to runGitCmd ", err, "trimPath", trimPath)
			valueVersion += "." + RandomString(7)
		} else {
			valueVersion += "." + out[:7]
		}
	}

	version := chartVersion + valueVersion
	hashVar := sha256.New()
	hashVar.Write([]byte(version))
	bs := hashVar.Sum(nil)
	workloadVersion = fmt.Sprintf("%x", bs)
	log.Println("createApplication: chart and values commit", version, "workloadVersion", workloadVersion[:7])

	// check worklaod version exist
	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("deployWorkload client error", err)
		return err
	}
	wl, err := client.AppsV3().Workload().Get(ctx, options.GetOptions{
		Name:    workload.Name,
		Project: project,
	})
	if err == nil {
		if wl.Spec.Version == workloadVersion[:7] {
			log.Println("workload version exist NOOP", workloadVersion[:7])
			st, err := getWorkLoadStatus(ctx, cfg, wl, folder, workloadVersion[:7])
			if err == nil {
				cfg.Status = append(cfg.Status, st)
			}
			return nil
		}
	}

	clusters := strings.Join(clusterNames, ",")
	log.Println("createApplication project:", project)
	workloadSpec := getWorkLoadSpec(cfg, workload, project, namespace, workloadName, chartPath, clusters, workloadVersion[:7], valuePaths)
	log.Println("workloadSpec", "\n---\n", workloadSpec, "\n---")

	err = deployWorkload(ctx, cfg, workloadSpec, folder, workloadVersion[:7])
	if err != nil {
		log.Println("createApplication: deployWorkload error", err)
		status := WorkloadCDStatus{}
		status.RepoFolder = folder
		status.Project = project
		status.Namespace = namespace
		status.WorkloadName = workload.Name
		status.Version = workloadVersion[:7]
		status.Status = &commonpb.Status{}
		status.Status.ConditionType = "Failed"
		status.Status.Reason = err.Error()
		cfg.Status = append(cfg.Status, &status)
		return err
	}

	return nil
}

func checkProject(ctx context.Context, project string) (string, []string, error) {
	// check if project exist
	var clusterNames []string

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("checkProject client error", err)
		return "", nil, err
	}

	_, err = client.SystemV3().Project().Get(ctx, options.GetOptions{
		Name: project,
	})
	if err != nil {
		if strings.Contains(err.Error(), "code 404") {
			log.Println("Resource Read ", "error", err)
			return "", nil, err
		}
	}
	resp, err := rctl_project.GetProjectByName(project)
	if err != nil {
		log.Println("project does not exist ", "error", err)
		return "", nil, err
	}
	pr, err := rctl_project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		log.Println("project does not exist ", "error", err)
		return "", nil, err
	}

	clusterList, err := rctl_cluster.ListAllClusters(pr.ID, "")
	if err != nil {
		log.Println("failed to list clusters in project ", "error", err)
	}
	for _, cl := range *clusterList {
		clusterNames = append(clusterNames, cl.Name)
	}
	log.Println("project", project, "clusterNames in the project", clusterNames)
	return pr.ID, clusterNames, nil
}

func deployWorkload(ctx context.Context, cfg *WorkloadCDConfig, workloadSpec, folder, version string) error {
	// deploy the workload
	h, err := hubYAMLCodec.Decode([]byte(workloadSpec))
	if err != nil {
		err = errors.Wrapf(err, "unable to decode spec")
		log.Println("deployWorkload spec decode error", err)
		return err
	}

	//artifactspec decode
	wl, ok := h.(*appspb.Workload)
	if !ok {
		err = errors.Wrapf(err, "unable to decode spec to appspb.Workload")
		log.Println("deployWorkload spec appspb.Workload decode error", err)
		return err
	}

	artifactNames, err := wl.ArtifactList()
	if err != nil {
		err = errors.Wrapf(err, "unable to list artifacts from spec %s", workloadSpec)
		log.Println("deployWorkload ArtifactList error", err)
		return err
	}

	log.Println("deployWorkload artifactNames", artifactNames)

	for _, artifactName := range artifactNames {
		//don't load artifacts not begining with file://
		if !strings.HasPrefix(artifactName, "file://") {
			continue
		}
		artifactFullPath := artifactName[7:]
		log.Println("deployWorkload artifactFullPath", artifactFullPath)
		artifactData, err := os.ReadFile(artifactFullPath)
		if err != nil {
			err = errors.Wrapf(err, "unable to read artifact at '%s'", artifactFullPath)
			log.Println("deployWorkload Artifact read error", err)
			return err
		}
		err = wl.ArtifactSet(artifactName, artifactData)
		if err != nil {
			err = errors.Wrapf(err, "unable to set artifact %s at path '%s'", artifactName, artifactFullPath)
			log.Println("deployWorkload ArtifactSet error", err)
			return err
		}
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("deployWorkload client error", err)
		return err
	}

	err = client.AppsV3().Workload().Apply(ctx, wl, options.ApplyOptions{})
	if err != nil {
		log.Println("deployWorkload workload apply error", err)
		return err
	}

	upsert := WorkloadsUpsert{}
	upsert.Project = wl.Metadata.Project
	upsert.Namespace = wl.Spec.Namespace
	upsert.WorkloadName = wl.Metadata.Name
	cfg.Upserts = append(cfg.Upserts, &upsert)

	status := WorkloadCDStatus{}
	status.RepoFolder = folder
	status.Project = wl.Metadata.Project
	status.Namespace = wl.Spec.Namespace
	status.WorkloadName = wl.Metadata.Name
	status.Version = version
	// wait for publish
	for {
		time.Sleep(15 * time.Second)
		wls, err := client.AppsV3().Workload().Status(ctx, options.StatusOptions{
			Name:    wl.Metadata.Name,
			Project: wl.Metadata.Project,
		})
		if err != nil {
			log.Println("deployWorkload workload status check error", err)
			return err
		}
		log.Println("wls.Status", wls.Status)
		status.Status = wls.Status
		if wls.Status != nil {
			//check if workload can be placed on a cluster, if true break out of loop
			if wls.Status.ConditionStatus == commonpb.ConditionStatus_StatusOK ||
				wls.Status.ConditionStatus == commonpb.ConditionStatus_StatusNotSet {
				break
			}
			if wls.Status.ConditionStatus == commonpb.ConditionStatus_StatusFailed {
				log.Println("failed to publish workload", wls.Status)
				return (fmt.Errorf("%s %s", "failed to publish workload", wls.Status))
			}
		} else {
			break
		}

	}
	log.Println("deployWorkload: workload status", status)
	cfg.Status = append(cfg.Status, &status)
	return nil
}

func getWorkLoadStatus(ctx context.Context, cfg *WorkloadCDConfig, wl *appspb.Workload, folder, version string) (*WorkloadCDStatus, error) {
	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("deployWorkload client error", err)
		return nil, err
	}

	err = client.AppsV3().Workload().Apply(ctx, wl, options.ApplyOptions{})
	if err != nil {
		log.Println("deployWorkload workload apply error", err)
		return nil, err
	}

	status := WorkloadCDStatus{}
	status.RepoFolder = folder
	status.Project = wl.Metadata.Project
	status.Namespace = wl.Spec.Namespace
	status.WorkloadName = wl.Metadata.Name
	status.Version = version

	wls, err := client.AppsV3().Workload().Status(ctx, options.StatusOptions{
		Name:    wl.Metadata.Name,
		Project: wl.Metadata.Project,
	})
	if err != nil {
		log.Println("deployWorkload workload status check error", err)
		return nil, err
	}
	log.Println("wls.Status", wls.Status)
	status.Status = wls.Status
	if wls.Status != nil {
		//check if workload can be placed on a cluster, if true break out of loop
		if wls.Status.ConditionStatus == commonpb.ConditionStatus_StatusOK ||
			wls.Status.ConditionStatus == commonpb.ConditionStatus_StatusNotSet {
		}
		if wls.Status.ConditionStatus == commonpb.ConditionStatus_StatusFailed {
			log.Println("failed to publish workload", wls.Status)
			return nil, (fmt.Errorf("%s %s", "failed to publish workload", wls.Status))
		}
	}
	return &status, nil
}
