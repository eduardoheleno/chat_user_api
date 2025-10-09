package routes

import (
	controller "nossochat_api/controllers"
	"nossochat_api/middlewares"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func SetContactRoutes(
	router *gin.Engine,
	db *gorm.DB,
	rabbit *amqp.Channel,
	redis *redis.Client,
) {
	contactController := controller.NewContactController(db, rabbit, redis)

	contactGroup := router.Group("/contact")
	contactGroup.Use(middlewares.ProtectRoute())
	contactGroup.POST("/create", contactController.Create)
	contactGroup.POST("/accept-invite/:idContact", contactController.AcceptInvite)
}
