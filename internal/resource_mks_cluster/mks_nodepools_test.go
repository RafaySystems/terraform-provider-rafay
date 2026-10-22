package resource_mks_cluster

import (
	"reflect"
	"testing"

	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
)

func TestSameK8sVersion(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"v1.34.2", "1.34.2", true},
		{"v1.34.2", "v1.34.2", true},
		{"v1.34", "v1.34.2", true},
		{"v1.34.2", "v1.34", true},
		{"v1.34", "1.34.9", true},
		{"v1.34.2", "v1.34.3", false},
		{"v1.33", "v1.34.2", false},
		{"v1.33.5", "v1.34.2", false},
		{"", "v1.34.2", false},
	}
	for _, c := range cases {
		if got := sameK8sVersion(c.a, c.b); got != c.want {
			t.Errorf("sameK8sVersion(%q,%q)=%v want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestNodepoolsToHub(t *testing.T) {
	got := nodepoolsToHub([]poolSpec{{"b", "v1.33.5", true}, {"a", "", false}})
	want := []*infrapb.MksNodePool{{Name: "a"}, {Name: "b", KubernetesVersion: "v1.33.5"}}
	if len(got) != len(want) {
		t.Fatalf("len %d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Name != want[i].Name || got[i].KubernetesVersion != want[i].KubernetesVersion {
			t.Errorf("[%d] got %v want %v", i, got[i], want[i])
		}
	}
}

func TestNodepoolsFromHub(t *testing.T) {
	sp := func(n string) *infrapb.MksNodePool { return &infrapb.MksNodePool{Name: n} }
	sv := func(n, v string) *infrapb.MksNodePool { return &infrapb.MksNodePool{Name: n, KubernetesVersion: v} }
	cases := []struct {
		name        string
		declared    []poolSpec
		declaredSet bool
		server      []*infrapb.MksNodePool
		want        []poolSpec
		keep        bool
	}{
		{"undeclared", nil, false, []*infrapb.MksNodePool{sv("gpu", "v1.33.5"), sp("default")}, nil, false},
		{"server-only ignored", []poolSpec{{"gpu", "", false}}, true, []*infrapb.MksNodePool{sp("gpu"), sp("default")}, []poolSpec{{"gpu", "", false}}, true},
		{"R69 equal full", []poolSpec{{"gpu", "v1.34.2", true}}, true, []*infrapb.MksNodePool{sp("gpu")}, []poolSpec{{"gpu", "v1.34.2", true}}, true},
		{"R69 major.minor", []poolSpec{{"gpu", "v1.34", true}}, true, []*infrapb.MksNodePool{sp("gpu")}, []poolSpec{{"gpu", "v1.34", true}}, true},
		{"declared differs from cp", []poolSpec{{"gpu", "v1.33.5", true}}, true, []*infrapb.MksNodePool{sp("gpu")}, []poolSpec{{"gpu", "", false}}, true},
		{"held back", []poolSpec{{"gpu", "", false}}, true, []*infrapb.MksNodePool{sv("gpu", "v1.33.5")}, []poolSpec{{"gpu", "v1.33.5", true}}, true},
		{"declared missing on server", []poolSpec{{"gpu", "", false}, {"old", "", false}}, true, []*infrapb.MksNodePool{sp("gpu")}, []poolSpec{{"gpu", "", false}}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, keep := nodepoolsFromHub(c.declared, c.declaredSet, c.server, "v1.34.2")
			if keep != c.keep {
				t.Fatalf("keep=%v want %v", keep, c.keep)
			}
			if keep && !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %v want %v", got, c.want)
			}
		})
	}
}
