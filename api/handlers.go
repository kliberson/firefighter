package api

import (
	suricata "firefighter/core"
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

		// 1. Usuń z firewall
		if err := suricata.UnblockIP(ip); err != nil {
			c.JSON(500, gin.H{"error": "Failed to unblock in firewall: " + err.Error()})
			return
		}

		// 2. Usuń z bazy
		if err := db.UnblockIP(ip); err != nil {
			c.JSON(500, gin.H{"error": "Failed to remove from database: " + err.Error()})
			return
		}

		// 3. Powiadom klientów WebSocket
		BroadcastUnblock(ip)

		c.JSON(200, gin.H{"message": "IP unblocked successfully"})
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
		ips, err := db.GetWhitelistDetails()
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve whitelisted IPs"})
			return
		}
		c.JSON(200, gin.H{"whitelisted_ips": ips})
	}
}

func addToWhitelist(db *data.DbManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.Param("ip") // ← Weź IP z URL parametru

		var req struct {
			Description string `json:"description"` // Tylko description z body
		}

		if err := c.BindJSON(&req); err != nil {
			// JSON może być pusty (tylko IP w URL)
			req.Description = ""
		}

		if err := db.AddToWhitelist(ip, req.Description); err != nil {
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
