package alarm_rule

import (
	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_rule/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_rule/repository"
	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_rule/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	// Inyección de dependencias
	do.Provide(injector, func(i *do.Injector) (repository.AlarmRuleRepository, error) {
		db := do.MustInvokeNamed[*gorm.DB](i, constants.DB)
		return repository.NewAlarmRuleRepository(db), nil
	})

	do.Provide(injector, func(i *do.Injector) (service.AlarmRuleService, error) {
		repo := do.MustInvoke[repository.AlarmRuleRepository](i)
		db := do.MustInvokeNamed[*gorm.DB](i, constants.DB)
		return service.NewAlarmRuleService(repo, db), nil
	})

	do.Provide(injector, func(i *do.Injector) (controller.AlarmRuleController, error) {
		svc := do.MustInvoke[service.AlarmRuleService](i)
		return controller.NewAlarmRuleController(i, svc), nil
	})

	alarmRuleController := do.MustInvoke[controller.AlarmRuleController](injector)

	alarmRuleRoutes := server.Group("/api/alarm_rules")
	{
		alarmRuleRoutes.POST("", alarmRuleController.Create)
		alarmRuleRoutes.GET("", alarmRuleController.FindAll)
	}
}
