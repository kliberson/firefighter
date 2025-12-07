package api

import (
	suricata "firefighter/core"
	"firefighter/data"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getBlocked(db *data.DbManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("🔍 DEBUG: Calling GetBlocked()...")
		ips, err := db.GetBlocked()
		if err != nil {
			log.Printf("💥 ERROR GetBlocked: %v", err) // ← TO POKAŻE BŁĄD!
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		log.Printf("✅ GetBlocked returned %d IPs", len(ips))
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

func getStats(db *data.DbManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := db.GetStats()
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve stats"})
			return
		}
		c.JSON(200, stats)
	}
}

func getHourlyAlerts(db *data.DbManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		daysStr := c.DefaultQuery("days", "7")
		days, _ := strconv.Atoi(daysStr)
		data, err := db.GetHourlyAlerts(days)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve hourly data"})
			return
		}
		c.JSON(200, gin.H{"data": data})
	}
}

func getTopIPs(db *data.DbManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "10")
		limit, _ := strconv.Atoi(limitStr)
		data, err := db.GetTopIPs(limit)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve top IPs"})
			return
		}
		c.JSON(200, gin.H{"data": data})
	}
}

func getAlertCategories(db *data.DbManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		daysStr := c.DefaultQuery("days", "7")
		days, _ := strconv.Atoi(daysStr)
		data, err := db.GetAlertCategories(days)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve categories"})
			return
		}
		c.JSON(200, gin.H{"data": data})
	}
}

func getRecentAlerts(db *data.DbManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "100")
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid limit"})
			return
		}

		alerts, err := db.GetRecentAlerts(limit)
		if err != nil {
			log.Printf("💥 GetRecentAlerts error: %v", err) // <-- to jest kluczowe
			c.JSON(500, gin.H{"error": "Failed to retrieve recent alerts"})
			return
		}

		c.JSON(200, gin.H{"alerts": alerts})
	}
}

func getAlertBuckets(db *data.DbManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		daysStr := c.DefaultQuery("days", "7")
		days, _ := strconv.Atoi(daysStr)
		data, err := db.GetAlertBuckets(days)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve alert buckets"})
			return
		}
		c.JSON(200, gin.H{"data": data})
	}
}

func getBlockBuckets(db *data.DbManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		daysStr := c.DefaultQuery("days", "7")
		days, _ := strconv.Atoi(daysStr)
		data, err := db.GetBlockBuckets(days)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve block buckets"})
			return
		}
		c.JSON(200, gin.H{"data": data})
	}
}
