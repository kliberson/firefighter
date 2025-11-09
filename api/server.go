package api

import (
	"firefighter/data"

	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/cors"
)

func SetupRouter(db *data.DbManager) *gin.Engine {
	r := gin.Default()

	r.Use(cors.Default())

	api := r.Group("/api")
	{
		api.GET("/blocked", getBlocked(db))
		api.POST("/unblock/:ip", unblockIP(db))
		api.GET("/alerts/:ip", getAlerts(db))
		api.GET("/whitelist", getWhitelisted(db))
		api.POST("/whitelist/:ip", addToWhitelist(db))
		api.DELETE("/whitelist/:ip", removeFromWhitelist(db))
	}

	r.GET("/ws", handleWebsocket)

	r.Static("/", "./frontend/dist")

	return r
}
