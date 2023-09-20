package rafay

type SystemComponentsPlacement struct {
	NodeSelector      map[string]string  `yaml:"nodeSelector,omitempty"`
	Tolerations       []*Tolerations     `yaml:"tolerations,omitempty"`
	DaemonsetOverride *DaemonsetOverride `yaml:"daemonsetOverride,omitempty"`
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
