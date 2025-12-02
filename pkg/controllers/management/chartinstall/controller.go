package chartinstall

import (
	"context"

	"github.com/rancher/rancher/pkg/types/config"
	"github.com/rancher/rancher/pkg/wrangler"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var chartRequirementGVK = schema.GroupVersionKind{
	Group:   "charts.rancher.io",
	Version: "v1alpha1",
	Kind:    "ChartRequirement",
}

func Register(ctx context.Context, mgmt *config.ManagementContext, wctx *wrangler.Context) {
	wctx.OnLeaderOrDie("chart-install-register", func(ctx context.Context) error {

		bundleController := wctx.Fleet.Bundle()
		bundleCache := bundleController.Cache()

		injector := NewInjector(wctx.Dynamic)
		handler := NewAppHandler(bundleCache, injector)

		// Bundle controller drives reconciliation
		bundleController.OnChange(ctx, "chart-install", handler.OnChange)

		return nil
	})
}
