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

func (h *AppHandler) OnChange(key string, app *fleet.Bundle) (*fleet.Bundle, error) {
	if app == nil {
		return nil, nil
	}

	//Need to check this if its a secret change
	// if secret.Type == "helm.sh/release.v1" {
	// 	// decode chart name + values from secret.Data["release"]
	// }

	app = app.DeepCopy()
	modified, err := h.injector.Reconcile(app)
	if err != nil {
		return nil, err
	}

	// this is doing nothing, it's just to debug/be explicit, I may remove later
	if modified {
		return app, nil
	}

	return app, nil
}
