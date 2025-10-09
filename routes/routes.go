package routes

import (
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func InitRoutes(db *gorm.DB, rabbit *amqp.Channel, redis *redis.Client) *gin.Engine {
	router := gin.Default()

	SetUserRoutes(router, db)
	SetContactRoutes(router, db, rabbit, redis)

	return router
}
