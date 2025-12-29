package api

import (
	suricata "firefighter/core"
	"firefighter/data"

	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/cors"
)

func SetupRouter(db data.Repository, wm *suricata.WindowManager) *gin.Engine {
	r := gin.Default()

	apiGroup := r.Group("/api")
	apiGroup.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "DELETE"},
		AllowHeaders: []string{"Origin", "Content-Type"},
	}))
	{
		// Blocked IPs
		apiGroup.GET("/blocked", getBlocked(db))
		apiGroup.GET("/blocked/by_ip", getBlockedByIPQuery(db))
		apiGroup.POST("/unblock/:ip", unblockIP(db, wm))

		// Whitelist
		apiGroup.GET("/whitelist", getWhitelisted(db))
		apiGroup.POST("/whitelist/:ip", addToWhitelist(db))
		apiGroup.DELETE("/whitelist/:ip", removeFromWhitelist(db))

		// Stats & Analytics
		apiGroup.GET("/stats", getStats(db))
		apiGroup.GET("/stats/hourly", getHourlyAlerts(db))
		apiGroup.GET("/stats/top_ips", getTopIPs(db))
		apiGroup.GET("/stats/categories", getAlertCategories(db))
		apiGroup.GET("/stats/recent_alerts", getRecentAlerts(db))
		apiGroup.GET("/stats/alerts/buckets", getAlertBuckets(db))
		apiGroup.GET("/stats/blocks/buckets", getBlockBuckets(db))
		apiGroup.GET("/stats/alerts/by_ip", getAlertsByIPQuery(db))

		apiGroup.GET("/activity", getActivity(db))
	}

	r.GET("/ws", handleWebSocket)

	r.Static("/assets", "/home/lucas/firefighter/frontend/dist/assets")
	r.StaticFile("/", "/home/lucas/firefighter/frontend/dist/index.html")
	r.NoRoute(func(c *gin.Context) {
		c.File("/home/lucas/firefighter/frontend/dist/index.html")
	})

	return r
}
