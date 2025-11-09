package api

import (
	"firefighter/data"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getBlocked(db *data.DbManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		ips, err := db.GetBlocked()
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve blocked IPs"})
			return
		}
		c.JSON(200, gin.H{"blocked_ips": ips})
	}
}

func unblockIP(db *data.DbManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.Param("ip")
		err := db.UnblockIP(ip)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"message": "unblocked"})
	}
}

// DO ZAIMPLEMENTOWANIA PO STRONIE BAZY DANYCH

func getAlerts(db *data.DbManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.Param("ip")
		if ip == "" {
			c.JSON(400, gin.H{"error": "Missing IP parameter"})
			return
		}

		limitStr := c.DefaultQuery("limit", "50") // domyślnie 50
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid limit parameter"})
			return
		}

		alerts, err := db.GetAlertsByIP(ip, limit)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve alerts"})
			return
		}

		c.JSON(200, gin.H{"alerts": alerts})
	}
}

func getWhitelisted(db *data.DbManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		ips, err := db.GetWhitelist()
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve whitelisted IPs"})
			return
		}
		c.JSON(200, gin.H{"whitelisted_ips": ips})
	}
}

func addToWhitelist(db *data.DbManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			IP          string `json:"ip"`
			Description string `json:"description"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}
		if err := db.AddToWhitelist(req.IP, req.Description); err != nil {
			c.JSON(500, gin.H{"error": "Failed to add IP to whitelist"})
			return
		}
		c.JSON(200, gin.H{"status": "IP added to whitelist"})
	}
}

func removeFromWhitelist(db *data.DbManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.Param("ip")
		if err := db.RemoveFromWhitelist(ip); err != nil {
			c.JSON(500, gin.H{"error": "Failed to remove IP from whitelist"})
			return
		}
		c.JSON(200, gin.H{"status": "IP removed from whitelist"})
	}
}
