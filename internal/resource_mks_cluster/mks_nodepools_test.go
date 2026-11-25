package resource_mks_cluster

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

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
		{"server version same as declared (no v)", []poolSpec{{"gpu", "1.33.5", true}}, true, []*infrapb.MksNodePool{sv("gpu", "v1.33.5")}, []poolSpec{{"gpu", "1.33.5", true}}, true},
		{"server version same as declared (major.minor)", []poolSpec{{"gpu", "v1.33", true}}, true, []*infrapb.MksNodePool{sv("gpu", "v1.33.5")}, []poolSpec{{"gpu", "v1.33", true}}, true},
		{"server version differs from declared", []poolSpec{{"gpu", "v1.32.1", true}}, true, []*infrapb.MksNodePool{sv("gpu", "v1.33.5")}, []poolSpec{{"gpu", "v1.33.5", true}}, true},
		// declared without a version is unmanaged: stays versionless whatever the server shows
		{"held back, declared without version", []poolSpec{{"gpu", "", false}}, true, []*infrapb.MksNodePool{sv("gpu", "v1.33.5")}, []poolSpec{{"gpu", "", false}}, true},
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

func testPoolsMap(ctx context.Context, pools map[string]attr.Value) types.Map {
	return types.MapValueMust(NodepoolsType{ObjectType: types.ObjectType{AttrTypes: NodepoolsValue{}.AttributeTypes(ctx)}}, pools)
}

func testPool(ctx context.Context, ver attr.Value) NodepoolsValue {
	return NewNodepoolsValueMust(NodepoolsValue{}.AttributeTypes(ctx), map[string]attr.Value{"kubernetes_version": ver})
}

func newTestConfig(ctx context.Context) ConfigValue {
	cfg := NewConfigValueNull()
	nt := NetworkValue{}.AttributeTypes(ctx)
	cni, _ := NewCniValueMust(CniValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"name":    types.StringNull(),
		"version": types.StringNull(),
	}).ToObjectValue(ctx)
	cfg.Network, _ = NewNetworkValueMust(nt, map[string]attr.Value{
		"cni":            cni,
		"ipv6":           types.ObjectNull(Ipv6Value{}.AttributeTypes(ctx)),
		"nameservers":    types.MapNull(types.StringType),
		"pod_subnet":     types.StringNull(),
		"service_subnet": types.StringNull(),
	}).ToObjectValue(ctx)
	return cfg
}

func TestConfigNodepoolsWiring(t *testing.T) {
	ctx := context.Background()

	t.Run("ToHub sorted", func(t *testing.T) {
		cfg := newTestConfig(ctx)
		cfg.Nodepools = testPoolsMap(ctx, map[string]attr.Value{
			"b": testPool(ctx, types.StringValue("v1.33.5")),
			"a": testPool(ctx, types.StringNull()),
		})
		hub, d := cfg.ToHub(ctx)
		if d.HasError() {
			t.Fatalf("diags: %v", d)
		}
		if len(hub.Nodepools) != 2 || hub.Nodepools[0].Name != "a" || hub.Nodepools[0].KubernetesVersion != "" ||
			hub.Nodepools[1].Name != "b" || hub.Nodepools[1].KubernetesVersion != "v1.33.5" {
			t.Fatalf("unexpected hub pools %v", hub.Nodepools)
		}
	})

	t.Run("ToHub null is nil", func(t *testing.T) {
		cfg := newTestConfig(ctx)
		hub, _ := cfg.ToHub(ctx)
		if hub.Nodepools != nil {
			t.Fatalf("want nil, got %v", hub.Nodepools)
		}
	})

	server := &infrapb.MksV3ConfigObject{
		KubernetesVersion: "v1.34.2",
		Network:           &infrapb.MksClusterNetworking{Cni: &infrapb.Cni{}},
		Nodepools:         []*infrapb.MksNodePool{{Name: "gpu"}, {Name: "default"}},
	}

	t.Run("FromHub null prior is null", func(t *testing.T) {
		cfg := newTestConfig(ctx)
		obj, _ := cfg.FromHub(ctx, server)
		got, _ := NewConfigValue(cfg.AttributeTypes(ctx), obj.Attributes())
		if !got.Nodepools.IsNull() {
			t.Fatalf("want null, got %v", got.Nodepools)
		}
	})

	t.Run("FromHub keeps only declared", func(t *testing.T) {
		cfg := newTestConfig(ctx)
		cfg.Nodepools = testPoolsMap(ctx, map[string]attr.Value{"gpu": testPool(ctx, types.StringNull())})
		obj, _ := cfg.FromHub(ctx, server)
		got, _ := NewConfigValue(cfg.AttributeTypes(ctx), obj.Attributes())
		els := got.Nodepools.Elements()
		if len(els) != 1 {
			t.Fatalf("want 1 pool, got %v", got.Nodepools)
		}
		gpu, ok := els["gpu"].(NodepoolsValue)
		if !ok || !gpu.KubernetesVersion.IsNull() {
			t.Fatalf("unexpected gpu pool %v", els["gpu"])
		}
	})
}
