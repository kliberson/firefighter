package api

import (
	suricata "firefighter/core"
	"firefighter/data"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getBlocked(db data.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		ips, err := db.GetBlocked()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"blocked_ips": ips})
	}
}

func unblockIP(db data.Repository, wm *suricata.WindowManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.Param("ip")

		if err := suricata.UnblockIP(ip); err != nil {
			c.JSON(500, gin.H{"error": "Failed to unblock in firewall: " + err.Error()})
			return
		}

		if err := db.UnblockIP(ip); err != nil {
			c.JSON(500, gin.H{"error": "Failed to remove from database: " + err.Error()})
			return
		}

		wm.RemoveIP(ip)

		BroadcastUnblock(ip)

		c.JSON(200, gin.H{"message": "IP unblocked successfully"})
	}
}

func getWhitelisted(db data.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		ips, err := db.GetWhitelistDetails()
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve whitelisted IPs"})
			return
		}
		c.JSON(200, gin.H{"whitelisted_ips": ips})
	}
}

func addToWhitelist(db data.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.Param("ip")

		var req struct {
			Description string `json:"description"`
		}

		if err := c.BindJSON(&req); err != nil {
			req.Description = ""
		}

		if err := db.AddToWhitelist(ip, req.Description); err != nil {
			c.JSON(500, gin.H{"error": "Failed to add IP to whitelist"})
			return
		}

		c.JSON(200, gin.H{"status": "IP added to whitelist"})
	}
}

func removeFromWhitelist(db data.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.Param("ip")
		if err := db.RemoveFromWhitelist(ip); err != nil {
			c.JSON(500, gin.H{"error": "Failed to remove IP from whitelist"})
			return
		}
		c.JSON(200, gin.H{"status": "IP removed from whitelist"})
	}
}

func getStats(db data.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := db.GetStats()
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve stats"})
			return
		}
		c.JSON(200, stats)
	}
}

func getHourlyAlerts(db data.Repository) gin.HandlerFunc {
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

func getTopIPs(db data.Repository) gin.HandlerFunc {
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

func getAlertCategories(db data.Repository) gin.HandlerFunc {
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

func getRecentAlerts(db data.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "100")
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid limit"})
			return
		}

		alerts, err := db.GetRecentAlerts(limit)
		if err != nil {
			log.Printf("GetRecentAlerts error: %v", err)
			c.JSON(500, gin.H{"error": "Failed to retrieve recent alerts"})
			return
		}

		c.JSON(200, gin.H{"alerts": alerts})
	}
}

func getAlertBuckets(db data.Repository) gin.HandlerFunc {
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

func getBlockBuckets(db data.Repository) gin.HandlerFunc {
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

func getAlertsByIPQuery(db data.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.Query("ip")
		if ip == "" {
			c.JSON(400, gin.H{"error": "ip param required"})
			return
		}

		alerts, err := db.GetAlertsByIP(ip, 100)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"alerts": alerts})
	}
}

func getBlockedByIPQuery(db data.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.Query("ip")
		if ip == "" {
			c.JSON(400, gin.H{"error": "ip param required"})
			return
		}

		blocks, err := db.GetBlockedByIP(ip)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"blocked_ips": blocks})
	}
}

func getActivity(db data.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		search := c.Query("search")
		typeFilter := c.Query("type") // ⬅️ DODAJ TO
		limitStr := c.DefaultQuery("limit", "100")
		limit, _ := strconv.Atoi(limitStr)

		activity, err := db.GetActivity(search, typeFilter, limit) // ⬅️ DODAJ typeFilter
		if err != nil {
			log.Printf("GetActivity error: %v", err)
			c.JSON(500, gin.H{"error": "Failed to retrieve activity"})
			return
		}

		c.JSON(200, gin.H{"activity": activity})
	}
}
