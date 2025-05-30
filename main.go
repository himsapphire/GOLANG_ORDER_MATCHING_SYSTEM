package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/himsapphire/GOLANG_ORDER_MATCHING_SYSTEM/db"
	"github.com/himsapphire/GOLANG_ORDER_MATCHING_SYSTEM/api"
)

func main() {
    // Initialize DB
    dsn := "root:password@tcp(127.0.0.1:3306)/order_matching?parseTime=true"
	

    db.InitDB(dsn)

    r := gin.Default()

    // Sample health check route
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })
	r.POST("/orders", api.CreateOrder)
	r.DELETE("/orders/:id", api.CancelOrder)
	r.GET("/orderbook", api.GetOrderBook)
	r.GET("/trades", api.ListTrades)
	r.GET("/orders/:id", api.GetOrderStatus)





    log.Println("Server running on http://localhost:6080")
    r.Run(":6080")
}
