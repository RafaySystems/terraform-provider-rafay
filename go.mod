module github.com/RafaySystems/terraform-provider-rafay

go 1.16

require (
	github.com/RafaySystems/rafay-common v1.10.0-beta8
	//github.com/RafaySystems/rafay-common v1.10.0-tf-schema-beta1
	github.com/RafaySystems/rctl v1.11.0-beta1
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/hashicorp/terraform-plugin-docs v0.4.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.7.1
	github.com/tidwall/gjson v1.9.3 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b

)

replace (
	//github.com/RafaySystems/rctl => github.com/RafaySystems/rctl v1.5.14
	//github.com/RafaySystems/terraform-provider-rafay/rafay => ../rafay
	//github.com/RafaySystems/rctl => ../rctl
	github.com/RafaySystems/rafay-common => ../rafay-common

	k8s.io/api => k8s.io/api v0.23.1
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.23.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.23.1
	k8s.io/apiserver => k8s.io/apiserver v0.23.1
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.23.1
	k8s.io/client-go => k8s.io/client-go v0.23.1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.23.1
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.23.1
	k8s.io/code-generator => k8s.io/code-generator v0.23.1
	k8s.io/component-base => k8s.io/component-base v0.23.1
	k8s.io/cri-api => k8s.io/cri-api v0.23.1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.23.1
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.23.1
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.23.1
	// this if required for to prevent kustomize from breaking
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20211115234752-e816edb12b65
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.23.1
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.23.1
	k8s.io/kubectl => k8s.io/kubectl v0.23.1
	k8s.io/kubelet => k8s.io/kubelet v0.23.1
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.23.1
	k8s.io/metrics => k8s.io/metrics v0.23.1
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.23.1
)
