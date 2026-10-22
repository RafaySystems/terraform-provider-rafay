package resource_mks_cluster

import (
	"fmt"
	"sort"
	"strings"

	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// poolSpec is the terraform-side view of one entry in config.nodepools.
type poolSpec struct {
	Name       string
	Version    string
	HasVersion bool
}

// sameK8sVersion is used only for drift decisions. It is deliberately more
// lenient than edgesrv's: a leading "v" is ignored and, on either side, a
// major.minor version matches any patch of that minor.
func sameK8sVersion(a, b string) bool {
	a, b = strings.TrimPrefix(a, "v"), strings.TrimPrefix(b, "v")
	if a == "" || b == "" {
		return false
	}
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	if len(pa) < 3 || len(pb) < 3 {
		if len(pa) < 2 || len(pb) < 2 {
			return a == b
		}
		return pa[0] == pb[0] && pa[1] == pb[1]
	}
	return a == b
}

func nodepoolsToHub(declared []poolSpec) []*infrapb.MksNodePool {
	out := make([]*infrapb.MksNodePool, 0, len(declared))
	for _, p := range declared {
		np := &infrapb.MksNodePool{Name: p.Name}
		if p.HasVersion {
			np.KubernetesVersion = p.Version
		}
		out = append(out, np)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// nodepoolsFromHub maps server pools back onto the declared ones. Pools that
// exist only on the server are ignored so they never cause drift.
func nodepoolsFromHub(declared []poolSpec, declaredSet bool, server []*infrapb.MksNodePool, clusterVersion string) ([]poolSpec, bool) {
	if !declaredSet {
		return nil, false
	}
	byName := make(map[string]*infrapb.MksNodePool, len(server))
	for _, s := range server {
		byName[s.Name] = s
	}
	out := make([]poolSpec, 0, len(declared))
	for _, d := range declared {
		s, ok := byName[d.Name]
		if !ok {
			continue
		}
		switch {
		case s.KubernetesVersion != "" && d.HasVersion && sameK8sVersion(d.Version, s.KubernetesVersion):
			out = append(out, d)
		case s.KubernetesVersion != "":
			out = append(out, poolSpec{d.Name, s.KubernetesVersion, true})
		case !d.HasVersion || sameK8sVersion(d.Version, clusterVersion):
			out = append(out, d)
		default:
			out = append(out, poolSpec{Name: d.Name})
		}
	}
	return out, true
}

// declaredPoolSpecs converts the terraform nodepools map into poolSpecs.
func declaredPoolSpecs(m types.Map) ([]poolSpec, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := make([]poolSpec, 0, len(m.Elements()))
	for name, el := range m.Elements() {
		np, ok := el.(NodepoolsValue)
		if !ok {
			diags.AddError("Invalid nodepools element",
				fmt.Sprintf("nodepools.%s has unexpected type %T", name, el))
			continue
		}
		out = append(out, poolSpec{
			Name:       name,
			Version:    getStringValue(np.KubernetesVersion),
			HasVersion: !np.KubernetesVersion.IsNull() && !np.KubernetesVersion.IsUnknown(),
		})
	}
	return out, diags
}
