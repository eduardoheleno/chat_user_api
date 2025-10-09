package routes

import (
	"nossochat_api/controllers"
	"nossochat_api/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetUserRoutes(router *gin.Engine, db *gorm.DB) {
	userController := controller.NewUserController(db)

	userGroup := router.Group("/user")
	userGroup.POST("/create", userController.Store)
	userGroup.POST("/login", userController.Login)
	userGroup.GET("/search", middlewares.ProtectRoute(), userController.Search)
}
