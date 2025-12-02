package chartinstall

import (
	fleet "github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	"github.com/rancher/lasso/pkg/dynamic"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
)

type Injector struct {
	dynamic *dynamic.Controller
}

func NewInjector(dynamic *dynamic.Controller) *Injector {
	return &Injector{
		dynamic: dynamic,
	}
}

// Core logic deciding if values must be injected
func (i *Injector) Reconcile(bundle *fleet.Bundle) (bool, error) {
	chartName := bundle.Spec.Helm.Chart
	version := bundle.Spec.Helm.Version

	if chartName == "" || version == "" {
		return false, nil
	}

	// Lookup dynamic ChartRequirement
	req, err := i.findChartRequirement(chartName, version)
	if err != nil {
		return false, err
	}
	if req == nil {
		return false, nil
	}

	userValues := toMap(bundle.Spec.Helm.Values)
	mandatory := extractRequiredValues(req)

	merged := deepMerge(userValues, mandatory)

	bundle.Spec.Helm.Values = toGenericMap(merged)
	return true, nil
}

func (i *Injector) findChartRequirement(chart, version string) (*unstructured.Unstructured, error) {
	// Build label selector
	selector := labels.Set{
		"charts.rancher.io/chart":   chart,
		"charts.rancher.io/version": version,
	}.AsSelector()

	// namespace = "" should list across all namespaces
	objs, err := i.dynamic.List(chartRequirementGVK, "", selector)
	if err != nil {
		return nil, err
	}

	if len(objs) == 0 {
		return nil, nil
	}
	u, ok := objs[0].(*unstructured.Unstructured)
	if !ok {
		return nil, nil // or return fmt.Errorf("unexpected type: %T", objs[0])
	}

	return u, nil
}

func extractRequiredValues(req *unstructured.Unstructured) map[string]interface{} {
	v, _, _ := unstructured.NestedMap(req.Object, "spec", "requiredValues")
	if v == nil {
		return map[string]interface{}{}
	}
	return v
}

// Minimal deep merge with mandatory override
func deepMerge(user map[string]interface{}, mandatory map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range user {
		out[k] = v
	}
	for k, mv := range mandatory {
		if mvMap, ok := mv.(map[string]interface{}); ok {
			if uvMap, ok2 := out[k].(map[string]interface{}); ok2 {
				out[k] = deepMerge(uvMap, mvMap)
				continue
			}
		}
		out[k] = mv // mandatory overrides
	}
	return out
}

func toMap(gm *fleet.GenericMap) map[string]interface{} {
	if gm == nil || gm.Data == nil {
		return map[string]interface{}{}
	}
	return gm.Data
}

func toGenericMap(m map[string]interface{}) *fleet.GenericMap {
	if m == nil {
		return &fleet.GenericMap{Data: map[string]interface{}{}}
	}
	return &fleet.GenericMap{Data: m}
}
