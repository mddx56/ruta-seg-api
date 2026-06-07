package health

import (
	"github.com/Caknoooo/go-gin-clean-starter/modules/health/controller"
	"github.com/samber/do"
)

func ProvideHealthController(i *do.Injector) (controller.HealthController, error) {
	return controller.NewHealthController(), nil
}
