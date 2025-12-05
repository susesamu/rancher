package chartinstall

import (
	"reflect"

	fleet "github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	"github.com/rancher/lasso/pkg/dynamic"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
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
	if err != nil || req == nil {
		return false, err
	}

	userValues := toMap(bundle.Spec.Helm.Values)
	mandatory := extractRequiredValues(req)
	merged := deepMerge(userValues, mandatory)

	if reflect.DeepEqual(userValues, merged) {
		return false, nil
	}

	bundle.Spec.Helm.Values = toGenericMap(merged)
	return true, nil
}

func (i *Injector) findChartRequirement(chart, version string) (*unstructured.Unstructured, error) {
	// Build label selector
	selector := labels.Set{
		"charts.rancher.io/chart":   chart,
		"charts.rancher.io/version": version,
	}.AsSelector()

	// namespace = "" list across all namespaces
	//this can be a future problem if the same charts exists in different namespaces or different versions
	objs, err := i.dynamic.List(chartRequirementGVK, "", selector)
	if err != nil {
		return nil, err
	}

	if len(objs) == 0 {
		return nil, nil
	}
	u := &unstructured.Unstructured{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(objs[0].(*unstructured.Unstructured).Object, u); err != nil {
		return nil, err
	}

	return u, nil
}

func extractRequiredValues(req *unstructured.Unstructured) map[string]any {
	v, _, err := unstructured.NestedMap(req.Object, "spec", "requiredValues")
	if err != nil {
		//TODO(susesamu): it would be good to log something here
		//log.Warnf(...)
	}
	if v == nil {
		return map[string]any{}
	}
	return v
}

// Minimal deep merge with mandatory override
func deepMerge(user map[string]any, mandatory map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range user {
		out[k] = v
	}
	for k, mv := range mandatory {
		if mvMap, ok := mv.(map[string]any); ok {
			if uvMap, ok2 := out[k].(map[string]any); ok2 {
				out[k] = deepMerge(uvMap, mvMap)
				continue
			}
		}
		out[k] = mv // mandatory overrides
	}
	return out
}

func toMap(gm *fleet.GenericMap) map[string]any {
	if gm == nil || gm.Data == nil {
		return map[string]any{}
	}
	return gm.Data
}

func toGenericMap(m map[string]any) *fleet.GenericMap {
	if m == nil {
		return &fleet.GenericMap{Data: map[string]any{}}
	}
	return &fleet.GenericMap{Data: m}
}
