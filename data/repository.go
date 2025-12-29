package data

type Repository interface {
	AddAlert(ip string, sid int, message string) error
	GetAlertsByIP(ip string, limit int) ([]AlertDetails, error)
	GetRecentAlerts(limit int) ([]AlertDetails, error)
	GetAlertBuckets(days int) ([]TimeBucket, error)

	AddBlocked(ip, reason string, score, alertCount, severityScore, uniquePorts, uniqueProtos, uniqueFlows int, categories, details string) error
	GetBlocked() ([]BlockedIPDetails, error)
	GetBlockedByIP(ip string) ([]BlockedIPDetails, error)
	UnblockIP(ip string) error
	IsBlocked(ip string) (bool, error)
	GetBlockBuckets(days int) ([]TimeBucket, error)

	AddToWhitelist(ip, description string) error
	RemoveFromWhitelist(ip string) error
	IsWhitelisted(ip string) (bool, error)
	GetWhitelistDetails() ([]WhitelistDetails, error)

	GetStats() (*Stats, error)
	GetHourlyAlerts(days int) ([]HourlyData, error)
	GetTopIPs(limit int) ([]TopIP, error)
	GetAlertCategories(days int) ([]Category, error)

	GetActivity(search string, typeFilter string, limit int) ([]ActivityEntry, error)

	Close() error
}

var _ Repository = (*DbManager)(nil)
