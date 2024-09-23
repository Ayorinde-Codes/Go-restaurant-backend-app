package routes

import (
	controller "golang-restaurant-backend-app/controllers"

	"github.com/gin-gonic/gin"
)

func OrderItemRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/order-items", controller.GetOrderItems())
	incomingRoutes.GET("/order-items/:order_item_id", controller.GetOrderItem())
	incomingRoutes.GET("/order-items-order/:order_id", controller.GetOrderItemsByOrder())
	incomingRoutes.POST("/order-items", controller.CreateOrderItem())
	incomingRoutes.PATCH("/order-items/:order_item_id", controller.UpdateOrderItem())
}
