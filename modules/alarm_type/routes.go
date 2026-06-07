package alarm_type

import (
	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_type/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_type/repository"
	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_type/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	// Inyección de dependencias con samber/do
	do.Provide(injector, func(i *do.Injector) (repository.AlarmTypeRepository, error) {
		db := do.MustInvokeNamed[*gorm.DB](i, constants.DB)
		return repository.NewAlarmTypeRepository(db), nil
	})

	do.Provide(injector, func(i *do.Injector) (service.AlarmTypeService, error) {
		repo := do.MustInvoke[repository.AlarmTypeRepository](i)
		db := do.MustInvokeNamed[*gorm.DB](i, constants.DB)
		return service.NewAlarmTypeService(repo, db), nil
	})

	do.Provide(injector, func(i *do.Injector) (controller.AlarmTypeController, error) {
		svc := do.MustInvoke[service.AlarmTypeService](i)
		return controller.NewAlarmTypeController(i, svc), nil
	})

	alarmTypeController := do.MustInvoke[controller.AlarmTypeController](injector)

	alarmTypeRoutes := server.Group("/api/alarm_types")
	{
		alarmTypeRoutes.POST("", alarmTypeController.Create)
		alarmTypeRoutes.GET("", alarmTypeController.FindAll)
	}
}
