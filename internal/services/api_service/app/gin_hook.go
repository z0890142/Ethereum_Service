package app

import (
	"Ethereum_Service/internal/services/api_service/controller"

	"github.com/gin-gonic/gin"
)

var defaultController *controller.Controller

func initCtrl(app *Application, r *gin.Engine) (*controller.Controller, error) {

	defaultController = controller.NewController()

	r.GET("/transaction/:txHash", defaultController.GetTransaction)
	r.GET("/blocks/:id", defaultController.GetBlock)

	return defaultController, nil
}

func InitGinApplicationHook(app *Application) error {
	if app.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	gin.EnableJsonDecoderUseNumber()
	r := gin.New()
	r.Use(gin.Recovery())

	initCtrl(app, r)

	return nil
}

func DestroyGinApplicationHook(app *Application) error {
	defaultController.Shutdown()
	return nil
}
