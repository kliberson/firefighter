package api

import (
	"firefighter/data"

	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/cors"
)

func SetupRouter(db *data.DbManager) *gin.Engine {
	r := gin.Default()

	apiGroup := r.Group("/api")
	apiGroup.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "DELETE"},
		AllowHeaders: []string{"Origin", "Content-Type"},
	}))
	{
		apiGroup.GET("/blocked", getBlocked(db))
		apiGroup.POST("/unblock/:ip", unblockIP(db))
		apiGroup.GET("/alerts/:ip", getAlerts(db))
		apiGroup.GET("/whitelist", getWhitelisted(db))
		apiGroup.POST("/whitelist/:ip", addToWhitelist(db))
		apiGroup.DELETE("/whitelist/:ip", removeFromWhitelist(db))
	}

	r.GET("/ws", handleWebSocket)

	r.Static("/assets", "/home/lucas/firefighter/frontend/dist/assets")
	r.StaticFile("/", "/home/lucas/firefighter/frontend/dist/index.html")
	r.NoRoute(func(c *gin.Context) {
		c.File("/home/lucas/firefighter/frontend/dist/index.html")
	})

	return r
}
