package routes

import (
	controller "golang-restaurant-backend-app/controllers"

	"github.com/gin-gonic/gin"
)

func TableRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/table", controller.GetTable())
	incomingRoutes.GET("/table/:table_id", controller.GetTable())
	incomingRoutes.POST("/table", controller.CreateTable())
	incomingRoutes.PATCH("/table/:table_id", controller.UpdateTable())
}
