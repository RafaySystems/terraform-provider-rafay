package rafay

import (
	"context"
	"fmt"
	"log"
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
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/RafaySystems/rafay-common/pkg/hub/codec"
	hub_types "github.com/RafaySystems/rafay-common/pkg/hub/conversion/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type CDCredentials struct {
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
	PrivateKey string `json:"privateKey,omitempty"`
	Token      string `json:"token,omitempty"`
}

type WorkloadCDConfigSpec struct {
	Type                string                            `json:"type,omitempty"`
	RepoURL             string                            `json:"repourl,omitempty"`
	RepoBranch          string                            `json:"repobranch,omitempty"`
	Credentials         *CDCredentials                    `json:"credentials,omitempty"`
	Options             *integrationspb.RepositoryOptions `json:"options,omitempty"`
	Insecure            bool                              `json:"insecure,omitempty"`
	PathMatchPattern    string                            `json:"pathMatchPattern,omitempty"` // the path  pattern to extract project name from
	RepositoryLocalPath string                            `json:"repositoryLocalPath,omitempty"`
	BasePath            string                            `json:"basePath,omitempty"` // the path  pattern to extract base chart from
	ScratchPad          string                            `json:"scratchPad,omitempty"`
	ClusterNames        string                            `json:"clusterNames,omitempty"`
}

type WorkloadCDConfig struct {
	ApiVersion string                `json:"apiVersion,omitempty"`
	Kind       string                `json:"kind,omitempty"`
	Metadata   *commonpb.Metadata    `json:"metadata,omitempty"`
	Spec       *WorkloadCDConfigSpec `json:"spec,omitempty"`
}

var (
	hubYAMLCodec = codec.NewYAMLCodec(hub_types.DefaultScheme)
)

var (
	commitSHARegex = regexp.MustCompile("^[0-9A-Fa-f]{40}$")
	sshURLRegex    = regexp.MustCompile("^(ssh://)?([^/:]*?)@[^@]+$")
	httpsURLRegex  = regexp.MustCompile("^(https://).*")
	httpURLRegex   = regexp.MustCompile("^(http://).*")
)

var _dummyHandler = func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {}

var WorkloadCDRepositorySchema = &schema.Resource{
	Description: "Workload CD Repository  definition",
	Schema: map[string]*schema.Schema{
		"always_run": &schema.Schema{
			Description: "name of the resource",
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
				"scratch_pad": &schema.Schema{
					Description: "folder to create temporary file",
					Optional:    true,
					Type:        schema.TypeString,
				},
				"base_path": &schema.Schema{
					Description: "repository local path",
					Optional:    true,
					Type:        schema.TypeString,
				},
				"repo_local_path": &schema.Schema{
					Description: "repository local path",
					Optional:    true,
					Type:        schema.TypeString,
				},
				"path_match_pattern": &schema.Schema{
					Description: "project/namespace/workload name path match pattern",
					Optional:    true,
					Type:        schema.TypeString,
				},
				"cluster_names": &schema.Schema{
					Description: "cluster names ',' separated",
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
	},
}

func resourceWorkloadCDOperator() *schema.Resource {
	modSchema := WorkloadCDRepositorySchema.Schema
	modSchema["impersonate"] = &schema.Schema{
		Description: "impersonate user",
		Optional:    true,
		Type:        schema.TypeString,
	}
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

func resourceWorkloadCDOperatorCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resourceWorkloadCDOperator create starts")
	diags := resourceWorkloadCDOperatorUpsert(ctx, d, m)

	return diags
}

func resourceWorkloadCDOperatorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resourceWorkloadCDOperator create starts")
	var diags diag.Diagnostics

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

	if workloadCDConfig.Spec.PathMatchPattern != "" {
		projectCheck := httprouter.New()
		projectCheck.Handle("POST", workloadCDConfig.Spec.PathMatchPattern, _dummyHandler)

		h, p, _ := projectCheck.Lookup("POST", "/project1/ns1/w1/values.yaml")
		log.Println("h:", h)

		if h != nil {
			// got a hit for URL
			project := p.ByName("project")
			log.Println("project:", project)
		}
	}

	output, err := cloneRepo(workloadCDConfig)
	if err != nil {
		log.Println("cloneRepo error", err)
	}
	log.Println("cloneRepo output", output)

	folders, files, baseChart, err := walkRepo(workloadCDConfig)
	if err != nil {
		log.Println("getRepoFiles error", err)
	}
	log.Println("cloneRepo files", files)
	log.Println("baseChart", baseChart)
	log.Println("folders", folders)

	processApplicationFolders(ctx, workloadCDConfig, baseChart, folders)
	d.SetId(workloadCDConfig.Metadata.Name)
	return diags
}

func expandWorkloadCDConfig(in *schema.ResourceData) (*WorkloadCDConfig, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand addon empty input")
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

	if v, ok := in["base_path"].(string); ok && len(v) > 0 {
		obj.BasePath = v
	}

	if v, ok := in["scratch_pad"].(string); ok && len(v) > 0 {
		obj.ScratchPad = v
	}
	// if v, ok := in["namespace_path_prefix"].(string); ok && len(v) > 0 {
	// 	obj.NamespacePathPrefix = v
	// }

	// if v, ok := in["workload_path_prefix"].(string); ok && len(v) > 0 {
	// 	obj.WorkloadNamePathPrefix = v
	// }

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
func runCmd(workloadCdCfg *WorkloadCDConfig, cmdDir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	return runCmdOutput(workloadCdCfg, cmdDir, cmd)
}

func runCmdOutput(workloadCdCfg *WorkloadCDConfig, cmdDir string, cmd *exec.Cmd) (string, error) {
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
	log.Println("runCmdOutput", cmd, "out", out, "err", err)
	return out, err
}

func cloneRepo(workloadCdCfg *WorkloadCDConfig) ([]string, error) {
	log.Printf("cloneRepo starts")
	repo_url := workloadCdCfg.Spec.RepoURL
	//repo_branch := workloadCdCfg.Spec.RepoBranch
	user := workloadCdCfg.Spec.Credentials.Username
	password := workloadCdCfg.Spec.Credentials.Password
	token := workloadCdCfg.Spec.Credentials.Token
	path := workloadCdCfg.Spec.RepositoryLocalPath

	//git -C ./repo pull || git clone https://stephan-rafay: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@github.com/stephan-rafay/test-tfcd.git ./repo
	out, err := runCmd(workloadCdCfg, workloadCdCfg.Spec.RepositoryLocalPath, "-C", path, "pull", "||")
	if err != nil {
		var url string
		// if the repo doesn't exist, we need to clone it
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
		out, err := runCmd(workloadCdCfg, ".", "clone", url, path)
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
func walkRepo(cfg *WorkloadCDConfig) ([]string, []string, string, error) {
	var files []string
	var folders []string
	var baseChart string

	root := cfg.Spec.RepositoryLocalPath
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".git") {
			return nil
		}
		if !info.IsDir() {
			files = append(files, path)
			if strings.Contains(path, cfg.Spec.BasePath) && strings.HasSuffix(path, ".tgz") {
				baseChart = path
			}
		} else {
			var isLeaf = true
			filepath.Walk(path, func(path1 string, info1 os.FileInfo, err1 error) error {
				//log.Println("path1", path1, "info", info1)
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
	return folders, files, baseChart, err
}

func processApplicationFolders(ctx context.Context, cfg *WorkloadCDConfig, baseChart string, folders []string) error {
	var chartPath string
	var wg sync.WaitGroup

	if baseChart != "" {
		chartPath = baseChart
	}

	for _, folder := range folders {
		// process folder and create application
		if chartPath == "" {
			// get the chart in the folder
			chartPath, _ = getChartInFolder(folder)
		}
		valuePaths, _ := getValuesInFolder(folder)
		wg.Add(1)
		// create application
		go createApplication(ctx, cfg, folder, chartPath, valuePaths, &wg)
		time.Sleep(time.Duration(2) * time.Second)
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

func getWorkLoadValues(project, namespace, workloadName, chartPath, clusterNames string, valuePaths []string) string {
	var vPth string

	log.Println("getWorkLoadValues project:", project)
	// get the workload values
	valueStr := "WorkloadName: " + workloadName + "\n"
	valueStr += "Project: " + project + "\n"
	valueStr += "ChartPath: " + chartPath + "\n"
	valueStr += "Namespace: " + namespace + "\n"
	for _, valuePath := range valuePaths {
		vPth += "- name: " + valuePath + "\n"
	}
	valueStr += "ValuesPaths:\n" + vPth
	valueStr += "ClusterNames: " + clusterNames + "\n"

	log.Println("getWorkLoadValues", valueStr)
	return valueStr
}

/*
`kind: Workload
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
	version: {{ .Version }}
*/
func getWorkLoadSpec(project, namespace, workloadName, chartPath, clusterNames string, valuePaths []string) string {
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
	spec += "    selector: rafay.dev/clusterName in (" + clusterNames + ")\n"
	spec += "  version: 1.0.0\n"

	return spec

}

func createTempFile(cfg *WorkloadCDConfig, prefix, workload string, b []byte) (string, error) {
	var root string
	if cfg.Spec.ScratchPad == "" {
		root = os.TempDir()
	} else {
		root = cfg.Spec.ScratchPad
	}
	dir, err := os.MkdirTemp(root, "workloadcd-"+workload)
	if err != nil {
		log.Println("createTempFile error", err)
		return "", err
	}
	filepath := dir + "/" + prefix + ".yaml"
	log.Println("createTempFile", filepath)
	return filepath, os.WriteFile(filepath, b, 0644)
}

func createApplication(ctx context.Context, cfg *WorkloadCDConfig, folder string, chartPath string, valuePaths []string, wg *sync.WaitGroup) error {
	// create application
	// use the template to create the application
	var project, namespace, workload string
	defer wg.Done()

	if cfg.Spec.PathMatchPattern != "" {
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
	}

	// check if project exist
	err := checkProject(ctx, project)
	if err != nil {
		log.Println("createApplication checkProject error", err)
		return err
	}

	if workload == "" {
		// use chart name as workload name
		strs := strings.Split(chartPath, "/")
		chartName := strs[len(strs)-1]
		workload = strings.TrimSuffix(chartName, ".tgz")
		log.Println("workload:", workload)
	}

	log.Println("createApplication project:", project)
	workloadSpec := getWorkLoadSpec(project, namespace, workload, chartPath, cfg.Spec.ClusterNames, valuePaths)
	log.Println("workloadSpec", "\n---\n", workloadSpec, "\n---")

	deployWorkload(ctx, cfg, workloadSpec)
	return nil
}

func checkProject(ctx context.Context, project string) error {
	// check if project exist

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("checkProject client error", err)
		return err
	}

	_, err = client.SystemV3().Project().Get(ctx, options.GetOptions{
		Name: project,
	})
	if err != nil {
		if strings.Contains(err.Error(), "code 404") {
			log.Println("Resource Read ", "error", err)
			return err
		}
	}
	return nil
}

func deployWorkload(ctx context.Context, cfg *WorkloadCDConfig, workloadSpec string) error {
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
	return nil
}
