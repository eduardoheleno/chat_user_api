package main

import (
	"nossochat_api/database"
	"nossochat_api/routes"
	util "nossochat_api/utils"

	"github.com/redis/go-redis/v9"
)

func main() {
	db := database.InitDatabase()
	rabbit := util.NewChannel()
	redis := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
		Password: "",
		DB: 0,
	})

	router := routes.InitRoutes(db, rabbit, redis)
	router.Run()
}
