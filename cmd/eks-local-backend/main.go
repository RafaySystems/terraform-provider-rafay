package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/go-yaml/yaml"
)

type store struct {
	Clusters map[string]storedCluster `json:"clusters"`
}

type storedCluster struct {
	Cluster   rafay.EKSCluster       `json:"cluster"`
	Config    rafay.EKSClusterConfig `json:"config"`
	UpdatedAt time.Time              `json:"updated_at"`
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "apply":
		applyCmd(os.Args[2:])
	case "read":
		readCmd(os.Args[2:])
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  eks-local-backend apply --file <cluster.yaml> [--store <store.json>]")
	fmt.Fprintln(os.Stderr, "  eks-local-backend read --name <cluster> [--project <project>] [--store <store.json>]")
	os.Exit(2)
}

func applyCmd(args []string) {
	fs := flag.NewFlagSet("apply", flag.ExitOnError)
	storePath := fs.String("store", defaultStorePath(), "Path to JSON store")
	filePath := fs.String("file", "", "Path to YAML file containing cluster and cluster_config docs")
	fs.Parse(args)

	if *filePath == "" {
		fmt.Fprintln(os.Stderr, "apply requires --file")
		os.Exit(2)
	}

	cluster, config, err := readClusterYAML(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read YAML: %v\n", err)
		os.Exit(1)
	}
	canonicalizeEKSClusterConfig(config)

	st, err := loadStore(*storePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load store: %v\n", err)
		os.Exit(1)
	}
	key := clusterKey(cluster)
	st.Clusters[key] = storedCluster{
		Cluster:   *cluster,
		Config:    *config,
		UpdatedAt: time.Now().UTC(),
	}
	if err := saveStore(*storePath, st); err != nil {
		fmt.Fprintf(os.Stderr, "save store: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("applied cluster %s\n", key)
}

func readCmd(args []string) {
	fs := flag.NewFlagSet("read", flag.ExitOnError)
	storePath := fs.String("store", defaultStorePath(), "Path to JSON store")
	name := fs.String("name", "", "Cluster name")
	project := fs.String("project", "", "Project name")
	fs.Parse(args)

	if *name == "" {
		fmt.Fprintln(os.Stderr, "read requires --name")
		os.Exit(2)
	}

	st, err := loadStore(*storePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load store: %v\n", err)
		os.Exit(1)
	}

	key := *name
	if *project != "" {
		key = fmt.Sprintf("%s/%s", *project, *name)
	}
	cluster, ok := st.Clusters[key]
	if !ok {
		fmt.Fprintf(os.Stderr, "cluster not found: %s\n", key)
		os.Exit(1)
	}

	enc := yaml.NewEncoder(os.Stdout)
	if err := enc.Encode(&cluster.Cluster); err != nil {
		fmt.Fprintf(os.Stderr, "write cluster YAML: %v\n", err)
		os.Exit(1)
	}
	if err := enc.Encode(&cluster.Config); err != nil {
		fmt.Fprintf(os.Stderr, "write config YAML: %v\n", err)
		os.Exit(1)
	}
}

func defaultStorePath() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ".eks-local-backend.json"
	}
	return filepath.Join(cwd, ".eks-local-backend.json")
}

func readClusterYAML(path string) (*rafay.EKSCluster, *rafay.EKSClusterConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	cluster := &rafay.EKSCluster{}
	if err := decoder.Decode(cluster); err != nil {
		return nil, nil, err
	}
	config := &rafay.EKSClusterConfig{}
	if err := decoder.Decode(config); err != nil {
		return nil, nil, err
	}
	return cluster, config, nil
}

func loadStore(path string) (store, error) {
	st := store{Clusters: map[string]storedCluster{}}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return st, nil
		}
		return st, err
	}
	if len(data) == 0 {
		return st, nil
	}
	if err := json.Unmarshal(data, &st); err != nil {
		return st, err
	}
	if st.Clusters == nil {
		st.Clusters = map[string]storedCluster{}
	}
	return st, nil
}

func saveStore(path string, st store) error {
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func clusterKey(cluster *rafay.EKSCluster) string {
	if cluster == nil || cluster.Metadata == nil {
		return "unknown"
	}
	if cluster.Metadata.Project == "" {
		return cluster.Metadata.Name
	}
	return fmt.Sprintf("%s/%s", cluster.Metadata.Project, cluster.Metadata.Name)
}

func canonicalizeEKSClusterConfig(cfg *rafay.EKSClusterConfig) {
	if cfg == nil {
		return
	}
	if len(cfg.NodeGroups) > 0 {
		sort.SliceStable(cfg.NodeGroups, func(i, j int) bool {
			return nodeGroupName(cfg.NodeGroups[i]) < nodeGroupName(cfg.NodeGroups[j])
		})
	}
	if len(cfg.ManagedNodeGroups) > 0 {
		sort.SliceStable(cfg.ManagedNodeGroups, func(i, j int) bool {
			return managedNodeGroupName(cfg.ManagedNodeGroups[i]) < managedNodeGroupName(cfg.ManagedNodeGroups[j])
		})
	}
}

func nodeGroupName(ng *rafay.NodeGroup) string {
	if ng == nil {
		return ""
	}
	return ng.Name
}

func managedNodeGroupName(ng *rafay.ManagedNodeGroup) string {
	if ng == nil {
		return ""
	}
	return ng.Name
}
