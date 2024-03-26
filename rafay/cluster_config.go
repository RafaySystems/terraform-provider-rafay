package rafay

type SystemComponentsPlacement struct {
	NodeSelector      map[string]string  `yaml:"nodeSelector,omitempty"`
	Tolerations       []*Tolerations     `yaml:"tolerations,omitempty"`
	DaemonsetOverride *DaemonsetOverride `yaml:"daemonSetOverride,omitempty"`
}

type Tolerations struct {
	Key               string `yaml:"key,omitempty"`
	Operator          string `yaml:"operator,omitempty"`
	Value             string `yaml:"value,omitempty"`
	Effect            string `yaml:"effect,omitempty"`
	TolerationSeconds *int   `yaml:"tolerationSeconds,omitempty"`
}

type DaemonsetOverride struct {
	NodeSelectionEnabled *bool          `yaml:"nodeSelectionEnabled,omitempty"`
	Tolerations          []*Tolerations `yaml:"tolerations,omitempty"`
}

type V1ClusterSharing struct {
	Enabled  *bool                      `yaml:"enabled,omitempty"`
	Projects []*V1ClusterSharingProject `yaml:"projects,omitempty"`
}

type V1ClusterSharingProject struct {
	Name string `yaml:"name,omitempty"`
}
