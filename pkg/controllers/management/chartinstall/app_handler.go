package chartinstall

import (
	fleet "github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	fleetcontrollers "github.com/rancher/rancher/pkg/generated/controllers/fleet.cattle.io/v1alpha1"
)

type AppHandler struct {
	apps     fleetcontrollers.BundleCache
	injector *Injector
}

func NewAppHandler(apps fleetcontrollers.BundleCache, injector *Injector) *AppHandler {
	return &AppHandler{
		apps:     apps,
		injector: injector,
	}
}

// THIS is the correct method
func (h *AppHandler) OnChange(key string, app *fleet.Bundle) (*fleet.Bundle, error) {
	if app == nil {
		return nil, nil
	}

	modified, err := h.injector.Reconcile(app)
	if err != nil {
		return nil, err
	}

	// If no modification, return original
	if !modified {
		return app, nil
	}

	return app, nil
}
