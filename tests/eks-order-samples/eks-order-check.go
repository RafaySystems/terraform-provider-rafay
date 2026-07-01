package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/go-yaml/yaml"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: eks-order-check <yaml>")
		os.Exit(2)
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	var cluster rafay.EKSCluster
	if err := dec.Decode(&cluster); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var cfg rafay.EKSClusterConfig
	if err := dec.Decode(&cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	managed := make([]string, 0)
	for _, ng := range cfg.ManagedNodeGroups {
		if ng == nil {
			continue
		}
		managed = append(managed, ng.Name)
	}
	nodes := make([]string, 0)
	for _, ng := range cfg.NodeGroups {
		if ng == nil {
			continue
		}
		nodes = append(nodes, ng.Name)
	}

	fmt.Printf("managed=%s\n", strings.Join(managed, ","))
	fmt.Printf("nodes=%s\n", strings.Join(nodes, ","))

	printManagedList("taints-ng", cfg.ManagedNodeGroups, listManagedTaints)
	printManagedList("attach-ng", cfg.ManagedNodeGroups, listManagedAttachIDs)
	printManagedList("instance-ng", cfg.ManagedNodeGroups, listManagedInstanceTypes)
	printManagedList("suspend-ng", cfg.ManagedNodeGroups, listManagedSuspendProcesses)
	printNodeList("taints-ng", cfg.NodeGroups, listNodeTaints)
	printNodeList("attach-ng", cfg.NodeGroups, listNodeAttachIDs)
	printNodeList("instance-ng", cfg.NodeGroups, listNodeInstanceTypes)
	printNodeList("suspend-ng", cfg.NodeGroups, listNodeSuspendProcesses)
	printNodeList("lb-ng", cfg.NodeGroups, listNodeClassicLBs)
	printNodeList("tg-ng", cfg.NodeGroups, listNodeTargetGroups)
}

type managedListFn func(*rafay.ManagedNodeGroup) []string

type nodeListFn func(*rafay.NodeGroup) []string

func printManagedList(name string, managed []*rafay.ManagedNodeGroup, fn managedListFn) {
	if fn == nil {
		return
	}
	for _, ng := range managed {
		if ng != nil && ng.Name == name {
			fmt.Printf("managed.%s=%s\n", name, strings.Join(fn(ng), ","))
			return
		}
	}
}

func printNodeList(name string, nodes []*rafay.NodeGroup, fn nodeListFn) {
	if fn == nil {
		return
	}
	for _, ng := range nodes {
		if ng != nil && ng.Name == name {
			fmt.Printf("node.%s=%s\n", name, strings.Join(fn(ng), ","))
			return
		}
	}
}

func listManagedTaints(ng *rafay.ManagedNodeGroup) []string {
	if ng == nil {
		return nil
	}
	out := make([]string, 0)
	for _, t := range ng.Taints {
		out = append(out, fmt.Sprintf("%s|%s|%s", t.Key, t.Effect, t.Value))
	}
	return out
}

func listManagedAttachIDs(ng *rafay.ManagedNodeGroup) []string {
	if ng == nil || ng.SecurityGroups == nil {
		return nil
	}
	return ng.SecurityGroups.AttachIDs
}

func listManagedInstanceTypes(ng *rafay.ManagedNodeGroup) []string {
	if ng == nil {
		return nil
	}
	return ng.InstanceTypes
}

func listManagedSuspendProcesses(ng *rafay.ManagedNodeGroup) []string {
	if ng == nil {
		return nil
	}
	return ng.ASGSuspendProcesses
}

func listNodeTaints(ng *rafay.NodeGroup) []string {
	if ng == nil {
		return nil
	}
	out := make([]string, 0)
	for _, t := range ng.Taints {
		out = append(out, fmt.Sprintf("%s|%s|%s", t.Key, t.Effect, t.Value))
	}
	return out
}

func listNodeAttachIDs(ng *rafay.NodeGroup) []string {
	if ng == nil || ng.SecurityGroups == nil {
		return nil
	}
	return ng.SecurityGroups.AttachIDs
}

func listNodeInstanceTypes(ng *rafay.NodeGroup) []string {
	if ng == nil || ng.InstancesDistribution == nil {
		return nil
	}
	return ng.InstancesDistribution.InstanceTypes
}

func listNodeSuspendProcesses(ng *rafay.NodeGroup) []string {
	if ng == nil {
		return nil
	}
	return ng.ASGSuspendProcesses
}

func listNodeClassicLBs(ng *rafay.NodeGroup) []string {
	if ng == nil {
		return nil
	}
	return ng.ClassicLoadBalancerNames
}

func listNodeTargetGroups(ng *rafay.NodeGroup) []string {
	if ng == nil {
		return nil
	}
	return ng.TargetGroupARNs
}
