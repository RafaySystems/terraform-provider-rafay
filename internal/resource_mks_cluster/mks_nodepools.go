package resource_mks_cluster

import (
	"sort"
	"strings"

	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
)

// poolSpec is the terraform-side view of one entry in config.nodepools.
type poolSpec struct {
	Name       string
	Version    string
	HasVersion bool
}

// sameK8sVersion mirrors edgesrv: a leading "v" is ignored, and a version
// with only major.minor matches any patch of that minor.
func sameK8sVersion(a, b string) bool {
	a, b = strings.TrimPrefix(a, "v"), strings.TrimPrefix(b, "v")
	if a == "" || b == "" {
		return false
	}
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	if len(pa) < 3 || len(pb) < 3 {
		n := 2
		if len(pa) < n || len(pb) < n {
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
