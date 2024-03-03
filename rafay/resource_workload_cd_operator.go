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

// The config spec for the WorkloadCD resource
type WorkloadCDConfigSpec struct {
	Type                string                            `json:"type,omitempty"`                // type of the repository - not used for now
	RepoURL             string                            `json:"repourl,omitempty"`             // the url of the repository
	RepoBranch          string                            `json:"repobranch,omitempty"`          // the branch of the repository
	Credentials         *CDCredentials                    `json:"credentials,omitempty"`         // the credentials to access the repository
	Options             *integrationspb.RepositoryOptions `json:"options,omitempty"`             // the options for the repository
	Insecure            bool                              `json:"insecure,omitempty"`            // allow insecure connection
	PathMatchPattern    string                            `json:"pathMatchPattern,omitempty"`    // the path  pattern to extract project name from
	RepositoryLocalPath string                            `json:"repositoryLocalPath,omitempty"` // the local path of the repository to clone
	BasePath            string                            `json:"basePath,omitempty"`            // the path  pattern to extract base chart from
	IncludeBaseValue    bool                              `json:"includeBaseValue,omitempty"`    // include base value.yaml
	DeleteAction        string                            `json:"enableDelete,omitempty"`        // delete the workload
	ClusterNames        string                            `json:"clusterNames,omitempty"`        // the cluster names to deploy the workload
	PlacementLabels     map[string]string                 `json:"placementLabels,omitempty"`     // the placement labels for the clusters
}

type WorkloadCDStatus struct {
	Project      string           `json:"project,omitempty"`      // the project of the workload
	Namespace    string           `json:"namespace,omitempty"`    // the namespace of the workload
	WorkloadName string           `json:"workloadName,omitempty"` // the name of the workload
	RepoFolder   string           `json:"repoFolder,omitempty"`   // the name of the workload
	Version      string           `json:"version,omitempty"`      // the version of the workload
	Status       *commonpb.Status `json:"status,omitempty"`       // the status of the workload
}
type WorkloadCDConfig struct {
	ApiVersion string                `json:"apiVersion,omitempty"` // the api version of the resource
	Kind       string                `json:"kind,omitempty"`       // the kind of the resource
	Metadata   *commonpb.Metadata    `json:"metadata,omitempty"`   // the metadata of the resource
	Spec       *WorkloadCDConfigSpec `json:"spec,omitempty"`       // the specification of the resource
	Status     []WorkloadCDStatus    `json:"status,omitempty"`     // the status of the resource
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
				"repo_local_path": &schema.Schema{
					Description: "repository local path",
					Optional:    true,
					Default:     "./apprepo",
					Type:        schema.TypeString,
				},
				"path_match_pattern": &schema.Schema{
					Description: "project/namespace/workload name path match pattern",
					Required:    true,
					Type:        schema.TypeString,
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
				"delete_action": &schema.Schema{
					Description: "workload delete action",
					Optional:    true,
					Default:     "none",
					Type:        schema.TypeString,
				},
				"options": &schema.Schema{
					Description: "repository options",
					Elem: &schema.Resource{Schema: map[string]*schema.Schema{
						"ca_cert": &schema.Schema{
							Description: "ca certificate",
							Elem: &schema.Resource{Schema: map[string]*schema.Schema{
								"data": &schema.Schema{
									Description: "data is the base64 encoded contents of the file",
									Optional:    true,
									Type:        schema.TypeString,
								},
								"mount_path": &schema.Schema{
									Description: "specify mount path of the file",
									Optional:    true,
									Type:        schema.TypeString,
								},
								"name": &schema.Schema{
									Description: "Name or relative path of a artifact",
									Optional:    true,
									Type:        schema.TypeString,
								},
								"options": &schema.Schema{
									Description: "specify options for the file",
									Elem: &schema.Resource{Schema: map[string]*schema.Schema{
										"description": &schema.Schema{
											Description: "Description of the file",
											Optional:    true,
											Type:        schema.TypeString,
										},
										"override": &schema.Schema{
											Description: "Override options for file",
											Elem: &schema.Resource{Schema: map[string]*schema.Schema{"type": &schema.Schema{
												Description: "Specify the type of override this file supports",
												Optional:    true,
												Type:        schema.TypeString,
											}}},
											MaxItems: 1,
											MinItems: 1,
											Optional: true,
											Type:     schema.TypeList,
										},
										"required": &schema.Schema{
											Description: "Determines whether the file is required / mandatory",
											Optional:    true,
											Type:        schema.TypeBool,
										},
										"sensitive": &schema.Schema{
											Description: "data is encrypted  if sensitive is set to true",
											Optional:    true,
											Type:        schema.TypeBool,
										},
									}},
									MaxItems: 1,
									MinItems: 1,
									Optional: true,
									Type:     schema.TypeList,
								},
								"sensitive": &schema.Schema{
									Description: "Deprected: use options.sensitive. data is encrypted  if sensitive is set to true",
									Optional:    true,
									Type:        schema.TypeBool,
								},
							}},
							MaxItems: 1,
							MinItems: 1,
							Optional: true,
							Type:     schema.TypeList,
						},
						"enable_lfs": &schema.Schema{
							Description: "enable git large file support",
							Optional:    true,
							Type:        schema.TypeBool,
						},
						"enable_submodules": &schema.Schema{
							Description: "enable git submodules",
							Optional:    true,
							Type:        schema.TypeBool,
						},
						"insecure": &schema.Schema{
							Description: "insecure",
							Optional:    true,
							Type:        schema.TypeBool,
						},
						"max_retires": &schema.Schema{
							Description: "max retries",
							Optional:    true,
							Type:        schema.TypeInt,
						},
					}},
					MaxItems: 1,
					MinItems: 1,
					Optional: true,
					Type:     schema.TypeList,
				},
				"type": &schema.Schema{
					Description: "repository type",
					Optional:    true,
					Type:        schema.TypeString,
				},
			}},
			MaxItems: 1,
			MinItems: 1,
			Optional: true,
			Type:     schema.TypeList,
		},
		"status": &schema.Schema{ // status of the resource get updated when the resource is created
			Description: "Status of the workload resource",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"project": &schema.Schema{
					Description: "Project of the resource",
					Optional:    true,
					Type:        schema.TypeString,
				},
				"namespace": &schema.Schema{
					Description: "Namespace of the resource",
					Optional:    true,
					Type:        schema.TypeString,
				},
				"workload_name": &schema.Schema{
					Description: "Workload Name of the resource",
					Optional:    true,
					Type:        schema.TypeString,
				},
				"workload_version": &schema.Schema{
					Description: "Workload Name of the resource",
					Optional:    true,
					Type:        schema.TypeString,
				},
				"repo_folder": &schema.Schema{
					Description: "repo path of the Workload resource",
					Optional:    true,
					Type:        schema.TypeString,
				},
				"condition_status": &schema.Schema{
					Description: "Condition Status",
					Optional:    true,
					Type:        schema.TypeInt,
				},
				"clusters": &schema.Schema{
					Description: "deployed clusters",
					Optional:    true,
					Type:        schema.TypeString,
				},
				"condition_type": &schema.Schema{
					Description: "Status message",
					Optional:    true,
					Type:        schema.TypeString,
				},
				"reason": &schema.Schema{
					Description: "Status message",
					Optional:    true,
					Type:        schema.TypeString,
				},
			}},
			MaxItems: 0,
			MinItems: 0,
			Optional: true,
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

func resourceWorkloadCDOperatorCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resourceWorkloadCDOperator create starts")
	diags := resourceWorkloadCDOperatorUpsert(ctx, d, m)

	return diags
}

func resourceWorkloadCDOperatorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resourceWorkloadCDOperator create starts")
	var diags diag.Diagnostics
	// dummy for now
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

	output, err := cloneRepo(workloadCDConfig)
	if err != nil {
		log.Println("cloneRepo error", err)
		return diag.FromErr(err)
	}
	log.Println("cloneRepo output", output)

	folders, files, baseChart, baseValues, err := walkRepo(workloadCDConfig)
	if err != nil {
		log.Println("getRepoFiles error", err)
		return diag.FromErr(err)
	}
	log.Println("cloneRepo files", files)
	log.Println("baseChart", baseChart)
	log.Println("folders", folders)
	log.Println("baseValues", baseValues)

	if workloadCDConfig.Spec.DeleteAction != "none" {
		processApplicationFoldersForDelete(ctx, workloadCDConfig, baseChart, folders)
	}

	processApplicationFolders(ctx, workloadCDConfig, baseChart, baseValues, folders)

	if workloadCDConfig.Status != nil && len(workloadCDConfig.Status) > 0 {
		v, ok := d.Get("stattus").([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret := flattenWorkloaStatus(workloadCDConfig.Status, v)
		d.Set("status", ret)
	}

	d.SetId(workloadCDConfig.Metadata.Name)
	return diags
}

func flattenWorkloaStatus(input []WorkloadCDStatus, p []interface{}) []interface{} {
	log.Println("flattenWorkloaStatus")
	if input == nil {
		return nil
	}

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

	if v, ok := in["type"].(string); ok && len(v) > 0 {
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
		obj.RepositoryLocalPath = v
	}

	if v, ok := in["path_match_pattern"].(string); ok && len(v) > 0 {
		obj.PathMatchPattern = v
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

	if v, ok := in["base_path"].(string); ok && len(v) > 0 {
		obj.BasePath = v
	}

	if v, ok := in["include_base_value"].(bool); ok {
		obj.IncludeBaseValue = v
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

	return obj, nil
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

// runCmd is a convenience function to run a command in a given directory and return its output
func runCmd(workloadCdCfg *WorkloadCDConfig, cmdDir string, senstive bool, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
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

func cloneRepo(workloadCdCfg *WorkloadCDConfig) ([]string, error) {
	var out string

	log.Printf("cloneRepo starts")
	repo_url := workloadCdCfg.Spec.RepoURL
	//repo_branch := workloadCdCfg.Spec.RepoBranch
	user := workloadCdCfg.Spec.Credentials.Username
	password := workloadCdCfg.Spec.Credentials.Password
	token := workloadCdCfg.Spec.Credentials.Token
	path := workloadCdCfg.Spec.RepositoryLocalPath

	//git -C ./apprepo pull
	out, err := runCmd(workloadCdCfg, ".", false, "-C", path, "pull")
	if err == nil {
		if workloadCdCfg.Spec.RepoBranch != "" {
			_, err := runCmd(workloadCdCfg, path, false, "checkout", workloadCdCfg.Spec.RepoBranch)
			if err != nil {
				log.Println("failed to checkout branch error", err)
			}
		}
	}
	if err != nil {
		var url string

		// remove the local repo if it exists
		runCmd(workloadCdCfg, ".", false, "rm -rf", path)
		// if the repo doesn't exist, we need to clone it
		// git clone --branch <branchname> https://stephan-rafay:api-key@url <path>
		if strings.Contains(repo_url, "https://") {
			strs := strings.Split(repo_url, "https://")
			if password != "" {
				url = fmt.Sprintf("https://%s:%s@%s", user, password, strs[1])
			} else if token != "" {
				url = fmt.Sprintf("https://%s:%s@%s", user, token, strs[1])
			}
		} else if strings.Contains(repo_url, "http://") {
			strs := strings.Split(repo_url, "http://")
			if password != "" {
				url = fmt.Sprintf("http://%s:%s@%s", user, password, strs[1])
			} else if token != "" {
				url = fmt.Sprintf("http://%s:%s@%s", user, token, strs[1])
			}
		}
		if workloadCdCfg.Spec.RepoBranch != "" {
			out, err = runCmd(workloadCdCfg, ".", true, "clone", "--branch", workloadCdCfg.Spec.RepoBranch, url, path)
		} else {
			out, err = runCmd(workloadCdCfg, ".", true, "clone", url, path)
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
func walkRepo(cfg *WorkloadCDConfig) ([]string, []string, string, []string, error) {
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
			files = append(files, path)
			if cfg.Spec.BasePath != "" {
				if strings.Contains(path, cfg.Spec.BasePath) && strings.HasSuffix(path, ".tgz") {
					baseChart = path
				}
				if strings.Contains(path, cfg.Spec.BasePath) &&
					(strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
					baseValues = append(baseValues, path)
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
				folders = append(folders, path)
			}
		}
		return nil
	})
	return folders, files, baseChart, baseValues, err
}

func processApplicationFoldersForDelete(ctx context.Context, cfg *WorkloadCDConfig, baseChart string, folders []string) error {
	var wg sync.WaitGroup
	var wrkList appspb.WorkloadList
	var wrkPrunedList appspb.WorkloadList

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("checkProject client error", err)
		return err
	}

	// Get all the projects
	projectList, err := client.SystemV3().Project().List(ctx, options.ListOptions{})
	if err != nil {
		log.Println("processApplicationFoldersForDelete failed to get projectList error", err)
		return err
	}

	// Get all the workloads created by the operator
	for _, pr := range projectList.Items {
		// Get all the workloads in the project
		wList, err := client.AppsV3().Workload().List(ctx, options.ListOptions{
			Project: pr.Metadata.Name,
		})
		if err != nil {
			log.Println("processApplicationFoldersForDelete failed to get workload List for project", pr.Metadata.Name, "error", err)
			return err
		}
		for _, w := range wList.Items {
			for k, _ := range w.Metadata.Labels {
				if k == "k8smgmt.io/helm-deployer-tfcd" {
					log.Println("found operator deployed workload", w.Metadata.Name, "project", w.Metadata.Project, "namespace", w.Spec.Namespace)
					wrkList.Items = append(wrkList.Items, w)
				}
			}
		}
	}

	for _, folder := range folders {
		// prune workload list
		var project, namespace, workload string
		var chartPath string

		// get the chart in the folder
		chartPath, _ = getChartInFolder(folder)
		if chartPath == "" && baseChart != "" {
			// get chart from baseChart
			chartPath = baseChart
		}

		valuePaths, _ := getValuesInFolder(folder)

		projectCheck := httprouter.New()
		pattern := strings.TrimPrefix(strings.TrimSuffix(cfg.Spec.RepositoryLocalPath, "/"), ".") + cfg.Spec.PathMatchPattern
		log.Println("folder:", folder, "PathMatchPattern", pattern)
		projectCheck.Handle("POST", pattern, _dummyHandler)
		h, p, _ := projectCheck.Lookup("POST", "/"+folder)
		log.Println("h:", h)

		if h != nil {
			// got a hit for URL
			project = p.ByName("project")
			log.Println("project:", project)

			namespace = p.ByName("namespace")
			log.Println("namespace:", namespace)

			workload = p.ByName("workload")
			log.Println("workload:", workload)
		}

		if workload == "" {
			// use chart name as workload name
			strs := strings.Split(chartPath, "/")
			chartName := strs[len(strs)-1]
			w1 := strings.TrimSuffix(chartName, ".tgz")
			w2 := strings.ReplaceAll(w1, ".", "-")
			workload = strings.ReplaceAll(w2, "_", "-")
			log.Println("workload:", workload)
		}
		for _, w := range wrkList.Items {
			if w.Metadata.Name == workload && w.Spec.Namespace == namespace && w.Metadata.Project == project {
				if chartPath != "" && len(valuePaths) > 0 {
					wrkPrunedList.Items = append(wrkPrunedList.Items, w)
				}
			}
		}
	}

	for _, w := range wrkList.Items {
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
			go deleteApplication(ctx, cfg, w.Metadata.Project, w.Metadata.Name, &wg)
		}
	}

	wg.Wait()
	return nil
}

func deleteApplication(ctx context.Context, cfg *WorkloadCDConfig, project, workload string, wg *sync.WaitGroup) error {
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

	if cfg.Spec.DeleteAction == "delete" {
		uri := fmt.Sprintf("/v2/config/project/%s/workload/%s", pr.ID, workload)
		println("delete uri", uri)
		_, err := auth.AuthAndRequest(uri, "DELETE", nil)
		if err != nil {
			log.Println("delete workload uri", uri, "error", err)
			return err
		}
	} else if cfg.Spec.DeleteAction == "unpublish" {

		uri := fmt.Sprintf("/v2/config/project/%s/workload/%s/unpublish", pr.ID, workload)
		println("unpublish uri", uri)
		_, err := auth.AuthAndRequest(uri, "POST", nil)
		if err != nil {
			log.Println("unpublish workload uri", uri, "error", err)
			return err
		}
	}
	return nil
}

func processApplicationFolders(ctx context.Context, cfg *WorkloadCDConfig, baseChart string, baseValues, folders []string) error {
	var chartPath string
	var wg sync.WaitGroup

	for _, folder := range folders {
		var valuePaths []string

		// process folder and create application
		chartPath = ""

		// get the chart in the folder
		chartPath, _ = getChartInFolder(folder)
		if chartPath == "" && baseChart != "" {
			chartPath = baseChart
		}

		if cfg.Spec.IncludeBaseValue {
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
		if chartPath != "" && len(valuePaths) > 0 {
			wg.Add(1)
			go createApplication(ctx, cfg, folder, chartPath, valuePaths, &wg)
			time.Sleep(time.Duration(2) * time.Second)
		} else {
			log.Println("processApplicationFolders ignore folder ", folder, "  chartPath or valuePaths is empty")
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
func getWorkLoadSpec(cfg *WorkloadCDConfig, project, namespace, workloadName, chartPath, clusterNames, version string, valuePaths []string) string {
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
	spec += "      chartPath:\n"
	spec += "        name: file://" + chartPath + "\n"
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
	if len(cfg.Spec.PlacementLabels) > 0 {
		spec += "    labels:\n"
		for k, v := range cfg.Spec.PlacementLabels {
			spec += "      - key: " + k + "\n"
			if v != "" {
				spec += "        value: " + v + "\n"
			}
		}
	}

	spec += "  version: " + version + "\n"

	return spec

}

func createApplication(ctx context.Context, cfg *WorkloadCDConfig, folder string, chartPath string, valuePaths []string, wg *sync.WaitGroup) error {
	// create application
	var project, namespace, workload string
	var clusterNames string
	var chartVersion string
	var valueVersion string
	var workloadVersion string
	defer wg.Done()

	projectCheck := httprouter.New()
	pattern := strings.TrimPrefix(strings.TrimSuffix(cfg.Spec.RepositoryLocalPath, "/"), ".") + cfg.Spec.PathMatchPattern
	log.Println("folder:", folder, "PathMatchPattern", pattern)
	projectCheck.Handle("POST", pattern, _dummyHandler)
	h, p, _ := projectCheck.Lookup("POST", "/"+folder)
	log.Println("h:", h)

	if h != nil {
		// got a hit for URL
		project = p.ByName("project")
		log.Println("project:", project)

		namespace = p.ByName("namespace")
		log.Println("namespace:", namespace)

		workload = p.ByName("workload")
		log.Println("workload:", workload)
	}

	if workload == "" {
		// use chart name as workload name
		strs := strings.Split(chartPath, "/")
		chartName := strs[len(strs)-1]
		w1 := strings.TrimSuffix(chartName, ".tgz")
		w2 := strings.ReplaceAll(w1, ".", "-")
		workload = strings.ReplaceAll(w2, "_", "-")
		log.Println("workload:", workload)
	}

	if project == "" || namespace == "" || workload == "" {
		log.Println("createApplication: project, namespace or workload is empty ignore folder", folder)
		return fmt.Errorf("createApplication: project, namespace or workload is empty ignore folder %s", folder)
	}

	// check if project exist
	_, clusterList, err := checkProject(ctx, project)
	if err != nil {
		log.Println("createApplication: checkProject error", err)
		status := WorkloadCDStatus{}
		status.RepoFolder = folder
		status.Project = project
		status.Namespace = namespace
		status.WorkloadName = workload
		status.Status.ConditionType = "Failed"
		status.Status.Reason = err.Error()
		cfg.Status = append(cfg.Status, status)
		return err
	}

	if cfg.Spec.ClusterNames == "" && len(cfg.Spec.PlacementLabels) <= 0 {
		// get cluster names from clusterList in the project
		if len(clusterList) <= 0 {
			err = fmt.Errorf("createApplication: no clusters found for project %s", project)
			log.Println(err)
			status := WorkloadCDStatus{}
			status.RepoFolder = folder
			status.Project = project
			status.Namespace = namespace
			status.WorkloadName = workload
			status.Status = &commonpb.Status{}
			status.Status.ConditionType = "Failed"
			status.Status.Reason = err.Error()
			cfg.Status = append(cfg.Status, status)
			return err
		}
		// get cluster names from clusterList in the project
		clusterNames = strings.Join(clusterList, ",")
	}
	if cfg.Spec.ClusterNames != "" {
		clusterNames = cfg.Spec.ClusterNames
	}

	// get chartPath version
	// git log -n1 --oneline --pretty=format:%H
	trimPath := pruneRepolocalPath(cfg.Spec.RepositoryLocalPath, chartPath)
	out, err := runCmd(cfg, cfg.Spec.RepositoryLocalPath, false, "log", "-n1", "--oneline", "--pretty=format:%H", trimPath)
	if err != nil {
		log.Println("failed to runCmd ", err, "trimPath", trimPath)
		chartVersion = RandomString(7)
	} else {
		chartVersion = out[:7]
	}

	// get valuePath version
	for _, valuePath := range valuePaths {
		trimPath = pruneRepolocalPath(cfg.Spec.RepositoryLocalPath, valuePath)
		out, err := runCmd(cfg, cfg.Spec.RepositoryLocalPath, false, "log", "-n1", "--oneline", "--pretty=format:%H", trimPath)
		if err != nil {
			log.Println("failed to runCmd ", err, "trimPath", trimPath)
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
		Name:    workload,
		Project: project,
	})
	if err == nil {
		if wl.Spec.Version == workloadVersion[:7] {
			log.Println("workload version exist NOOP", workloadVersion[:7])
			st, err := getWorkLoadStatus(ctx, cfg, wl, folder, workloadVersion[:7])
			if err == nil {
				cfg.Status = append(cfg.Status, *st)
			}
			return nil
		}
	}

	log.Println("createApplication project:", project)
	workloadSpec := getWorkLoadSpec(cfg, project, namespace, workload, chartPath, clusterNames, workloadVersion[:7], valuePaths)
	log.Println("workloadSpec", "\n---\n", workloadSpec, "\n---")

	err = deployWorkload(ctx, cfg, workloadSpec, folder, workloadVersion[:7])
	if err != nil {
		log.Println("createApplication: deployWorkload error", err)
		status := WorkloadCDStatus{}
		status.RepoFolder = folder
		status.Project = project
		status.Namespace = namespace
		status.WorkloadName = workload
		status.Version = workloadVersion[:7]
		status.Status = &commonpb.Status{}
		status.Status.ConditionType = "Failed"
		status.Status.Reason = err.Error()
		cfg.Status = append(cfg.Status, status)
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
	log.Println("project", project, "clusterNames", clusterNames)
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

	status := WorkloadCDStatus{}
	status.RepoFolder = folder
	status.Project = wl.Metadata.Project
	status.Namespace = wl.Spec.Namespace
	status.WorkloadName = wl.Metadata.Name
	status.Version = version
	// wait for publish
	for {
		time.Sleep(60 * time.Second)
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
	cfg.Status = append(cfg.Status, status)
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
