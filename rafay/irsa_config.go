package rafay

import (
	commonpb "github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
)

type IRSA struct {
	Kind     string             `yaml:"kind,omitempty"`
	Metadata *commonpb.Metadata `yaml:"metadata,omitempty"`
	Spec     *IRSASpec          `yaml:"spec,omitempty"`
}

type IRSASpec struct {
	ClusterName         string            `yaml:"clusterName,omitempty"`
	Namespace           string            `yaml:"namespace,omitempty"`
	PermissionsBoundary string            `yaml:"permissionBoundary,omitempty"`
	RoleOnly            *bool             `yaml:"roleOnly,omitempty"`
	Tags                map[string]string `yaml:"tags,omitempty"`
	PolicyARNs          []string          `yaml:"policyARNs,omitempty"`
	//def need to do something different here with map[string]interface{}
	PolicyDocument string `yaml:"policyDocument,omitempty"`
}

type PolicyDocument struct {
	Version   string          `yaml:"Version,omitempty"`
	Statement PolicyStatement `yaml:"Statement,omitempty"`
}
type PolicyStatement struct {
	Sid      string   `yaml:"Sid,omitempty"`
	Effect   string   `yaml:"Effect,omitempty"`
	Action   []string `yaml:"Action,omitempty"`
	Resource []string `yaml:"Resource,omitempty"`
}
