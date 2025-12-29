package data

import (
	"database/sql"
	"fmt"
	"sort"
	"time"

	_ "modernc.org/sqlite"
)

type DbManager struct {
	db *sql.DB
}

type BlockedIPDetails struct {
	IP            string `json:"ip"`
	Reason        string `json:"reason"`
	Score         int    `json:"score"`
	AlertCount    int    `json:"alert_count"`
	SeverityScore int    `json:"severity_score"`
	UniquePorts   int    `json:"unique_ports"`
	UniqueProtos  int    `json:"unique_protos"`
	UniqueFlows   int    `json:"unique_flows"`
	Categories    string `json:"categories"`
	Details       string `json:"details"`
	Timestamp     int64  `json:"timestamp"`
}

type WhitelistDetails struct {
	IP          string `json:"ip"`
	Description string `json:"description"`
	AddedAt     int64  `json:"added_at"`
}

type AlertDetails struct {
	ID        int       `json:"id"`
	IP        string    `json:"ip"`
	SID       int       `json:"sid"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type Stats struct {
	TotalAlerts  int `json:"total_alerts"`
	TotalBlocked int `json:"total_blocked"`
	UniqueIPs    int `json:"unique_ips"`
}

type HourlyData struct {
	Hour  string `json:"hour"`
	Count int    `json:"count"`
}

type TimeBucket struct {
	Bucket string `json:"bucket"`
	Count  int    `json:"count"`
}

type TopIP struct {
	IP    string `json:"ip"`
	Count int    `json:"count"`
}

type Category struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type ActivityEntry struct {
	Type      string `json:"type"` // "alert", "block", "unblock", "whitelist_add", "whitelist_remove"
	Timestamp int64  `json:"timestamp"`
	IP        string `json:"ip"`
	Details   string `json:"details"` // message/reason/description
	Extra     string `json:"extra"`   // SID dla alertów, Score dla bloków
}

func New(path string) (Repository, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS blocked_ips (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            ip TEXT NOT NULL,
            reason TEXT,
            score INTEGER DEFAULT 0,
            alert_count INTEGER DEFAULT 0,
            severity_score INTEGER DEFAULT 0,
            unique_ports INTEGER DEFAULT 0,
            unique_protos INTEGER DEFAULT 0,
            unique_flows INTEGER DEFAULT 0,
            categories TEXT,
            details TEXT,
            timestamp INTEGER DEFAULT (strftime('%s', 'now')),
            unblock_time INTEGER,
            status TEXT DEFAULT 'blocked' CHECK(status IN ('blocked', 'unblocked'))
        );
        
        CREATE TABLE IF NOT EXISTS alerts (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            ip TEXT NOT NULL,
            sid INTEGER,
            message TEXT,
            timestamp INTEGER DEFAULT (strftime('%s', 'now'))
        );
        
        CREATE TABLE IF NOT EXISTS whitelist (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            ip TEXT NOT NULL UNIQUE,
            description TEXT,
            added_at INTEGER DEFAULT (strftime('%s', 'now'))
        );
        
        CREATE INDEX IF NOT EXISTS idx_blocked_ips_ip ON blocked_ips(ip);
        CREATE INDEX IF NOT EXISTS idx_blocked_ips_status ON blocked_ips(status);
        CREATE INDEX IF NOT EXISTS idx_alerts_ip ON alerts(ip);
        CREATE INDEX IF NOT EXISTS idx_alerts_timestamp ON alerts(timestamp);
    `)

	if err != nil {
		return nil, err
	}

	return &DbManager{db: db}, nil
}

func (s *DbManager) AddBlocked(ip, reason string, score, alertCount, severityScore, uniquePorts, uniqueProtos, uniqueFlows int, categories, details string) error {
	_, err := s.db.Exec(`
        INSERT INTO blocked_ips (ip, reason, score, alert_count, severity_score, unique_ports, unique_protos, unique_flows, categories, details) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `, ip, reason, score, alertCount, severityScore, uniquePorts, uniqueProtos, uniqueFlows, categories, details)
	return err
}

func (s *DbManager) AddAlert(ip string, sid int, message string) error {
	_, err := s.db.Exec(`INSERT INTO alerts (ip, sid, message) VALUES (?, ?, ?)`, ip, sid, message)
	return err
}

func (s *DbManager) GetBlocked() ([]BlockedIPDetails, error) {
	rows, err := s.db.Query(`
        SELECT ip, reason, score, alert_count, severity_score, unique_ports, unique_protos, unique_flows, categories, details, timestamp 
        FROM blocked_ips 
        WHERE status='blocked'
        ORDER BY timestamp DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ips []BlockedIPDetails
	for rows.Next() {
		var ip BlockedIPDetails
		if err := rows.Scan(&ip.IP, &ip.Reason, &ip.Score, &ip.AlertCount, &ip.SeverityScore,
			&ip.UniquePorts, &ip.UniqueProtos, &ip.UniqueFlows,
			&ip.Categories, &ip.Details, &ip.Timestamp); err != nil {
			return nil, err
		}
		ips = append(ips, ip)
	}

	return ips, rows.Err()
}

func (s *DbManager) UnblockIP(ip string) error {
	now := time.Now().Unix()

	result, err := s.db.Exec(`
        UPDATE blocked_ips 
        SET status='unblocked', unblock_time=? 
        WHERE ip=? AND status='blocked'
    `, now, ip)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *DbManager) IsBlocked(ip string) (bool, error) {
	row := s.db.QueryRow(`
        SELECT COUNT(1) 
        FROM blocked_ips 
        WHERE ip=? AND status='blocked'
    `, ip)

	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *DbManager) GetAlertsByIP(ip string, limit int) ([]AlertDetails, error) {
	rows, err := s.db.Query(`
        SELECT id, ip, sid, message, timestamp
        FROM alerts
        WHERE ip=?
        ORDER BY timestamp DESC
        LIMIT ?
    `, ip, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []AlertDetails
	for rows.Next() {
		var a AlertDetails
		var timestamp int64

		err := rows.Scan(&a.ID, &a.IP, &a.SID, &a.Message, &timestamp)
		if err != nil {
			return nil, err
		}

		a.Timestamp = time.Unix(timestamp, 0)
		alerts = append(alerts, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return alerts, nil
}

func (s *DbManager) GetBlockedByIP(ip string) ([]BlockedIPDetails, error) {
	rows, err := s.db.Query(`
        SELECT ip, reason, score, alert_count, severity_score, unique_ports, unique_protos, unique_flows, categories, details, timestamp
        FROM blocked_ips
        WHERE ip = ?
        ORDER BY timestamp DESC
    `, ip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []BlockedIPDetails
	for rows.Next() {
		var b BlockedIPDetails
		if err := rows.Scan(&b.IP, &b.Reason, &b.Score, &b.AlertCount, &b.SeverityScore,
			&b.UniquePorts, &b.UniqueProtos, &b.UniqueFlows,
			&b.Categories, &b.Details, &b.Timestamp); err != nil {
			return nil, err
		}
		blocks = append(blocks, b)
	}
	return blocks, nil
}

func (s *DbManager) AddToWhitelist(ip, description string) error {
	_, err := s.db.Exec(`INSERT OR IGNORE INTO whitelist (ip, description) VALUES (?, ?)`, ip, description)
	return err
}

func (s *DbManager) RemoveFromWhitelist(ip string) error {
	_, err := s.db.Exec(`DELETE FROM whitelist WHERE ip = ?`, ip)
	return err
}

func (s *DbManager) IsWhitelisted(ip string) (bool, error) {
	row := s.db.QueryRow(`SELECT COUNT(1) FROM whitelist WHERE ip = ?`, ip)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *DbManager) GetWhitelistDetails() ([]WhitelistDetails, error) {
	rows, err := s.db.Query(`SELECT ip, description, added_at FROM whitelist ORDER BY added_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []WhitelistDetails
	for rows.Next() {
		var item WhitelistDetails
		if err := rows.Scan(&item.IP, &item.Description, &item.AddedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (s *DbManager) GetStats() (*Stats, error) {
	stats := &Stats{}

	err := s.db.QueryRow("SELECT COUNT(*) FROM alerts").Scan(&stats.TotalAlerts)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow("SELECT COUNT(*) FROM blocked_ips WHERE status='blocked'").Scan(&stats.TotalBlocked)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow("SELECT COUNT(DISTINCT ip) FROM alerts").Scan(&stats.UniqueIPs)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (s *DbManager) GetHourlyAlerts(days int) ([]HourlyData, error) {
	rows, err := s.db.Query(`
        SELECT 
            strftime('%Y-%m-%d %H:00', datetime(timestamp, 'unixepoch')) as hour,
            COUNT(*) as count
        FROM alerts 
        WHERE timestamp > (strftime('%s','now',?))
        GROUP BY hour 
        ORDER BY hour DESC
        LIMIT 168`, fmt.Sprintf("-%d days", days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []HourlyData
	for rows.Next() {
		var d HourlyData
		if err := rows.Scan(&d.Hour, &d.Count); err != nil {
			return nil, err
		}
		data = append(data, d)
	}

	return data, rows.Err()
}

func (s *DbManager) GetTopIPs(limit int) ([]TopIP, error) {
	rows, err := s.db.Query(`
        SELECT ip, COUNT(*) as count 
        FROM alerts 
        GROUP BY ip 
        ORDER BY count DESC 
        LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ips []TopIP
	for rows.Next() {
		var ip TopIP
		if err := rows.Scan(&ip.IP, &ip.Count); err != nil {
			return nil, err
		}
		ips = append(ips, ip)
	}

	return ips, rows.Err()
}

func (s *DbManager) GetAlertCategories(days int) ([]Category, error) {
	rows, err := s.db.Query(`
        SELECT 
            substr(message,1,50) as category,
            COUNT(*) as count
        FROM alerts 
        WHERE timestamp > (strftime('%s','now',?))
        GROUP BY category 
        ORDER BY count DESC 
        LIMIT 10`, fmt.Sprintf("-%d days", days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []Category
	for rows.Next() {
		var cat Category
		if err := rows.Scan(&cat.Name, &cat.Count); err != nil {
			return nil, err
		}
		cats = append(cats, cat)
	}

	return cats, rows.Err()
}

func (s *DbManager) GetRecentAlerts(limit int) ([]AlertDetails, error) {
	rows, err := s.db.Query(`
        SELECT id, ip, sid, message, timestamp
        FROM alerts
        ORDER BY timestamp DESC
        LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []AlertDetails
	for rows.Next() {
		var a AlertDetails
		var timestamp int64
		err := rows.Scan(&a.ID, &a.IP, &a.SID, &a.Message, &timestamp)
		if err != nil {
			return nil, err
		}
		a.Timestamp = time.Unix(timestamp, 0)
		alerts = append(alerts, a)
	}

	return alerts, rows.Err()
}

func (s *DbManager) GetAlertBuckets(days int) ([]TimeBucket, error) {
	var format string
	if days <= 1 {
		format = "%Y-%m-%d %H:00"
	} else {
		format = "%Y-%m-%d"
	}

	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour).Unix()

	rows, err := s.db.Query(fmt.Sprintf(`
        SELECT 
            strftime('%s', datetime(timestamp, 'unixepoch', 'localtime')) AS bucket,
            COUNT(*) AS count
        FROM alerts
        WHERE timestamp > ?
        GROUP BY bucket
        ORDER BY bucket ASC
    `, format), cutoff)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TimeBucket
	for rows.Next() {
		var b TimeBucket
		if err := rows.Scan(&b.Bucket, &b.Count); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *DbManager) GetBlockBuckets(days int) ([]TimeBucket, error) {
	var format string
	if days <= 1 {
		format = "%Y-%m-%d %H:00"
	} else {
		format = "%Y-%m-%d"
	}

	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour).Unix()

	query := fmt.Sprintf(`
        SELECT 
            strftime('%s', datetime(timestamp, 'unixepoch', 'localtime')) AS bucket,
            COUNT(*) AS count
        FROM blocked_ips
        WHERE timestamp > ?
        GROUP BY bucket
        ORDER BY bucket ASC
    `, format)

	rows, err := s.db.Query(query, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TimeBucket
	for rows.Next() {
		var b TimeBucket
		if err := rows.Scan(&b.Bucket, &b.Count); err != nil {
			return nil, err
		}
		out = append(out, b)
	}

	return out, rows.Err()
}

func (s *DbManager) GetActivity(search string, typeFilter string, limit int) ([]ActivityEntry, error) {
	var entries []ActivityEntry

	// 1. Alerty (tylko jeśli typeFilter="" lub "alert")
	if typeFilter == "" || typeFilter == "alert" {
		alertQuery := "SELECT timestamp, ip, message, CAST(sid AS TEXT) FROM alerts WHERE 1=1"
		alertArgs := []interface{}{}

		if search != "" {
			alertQuery += " AND (ip LIKE ? OR message LIKE ?)"
			searchPattern := "%" + search + "%"
			alertArgs = append(alertArgs, searchPattern, searchPattern)
		}

		alertQuery += " ORDER BY timestamp DESC LIMIT ?"
		alertArgs = append(alertArgs, limit)

		alertRows, err := s.db.Query(alertQuery, alertArgs...)
		if err != nil {
			return nil, err
		}
		defer alertRows.Close()

		for alertRows.Next() {
			var entry ActivityEntry
			entry.Type = "alert"
			if err := alertRows.Scan(&entry.Timestamp, &entry.IP, &entry.Details, &entry.Extra); err != nil {
				return nil, err
			}
			entries = append(entries, entry)
		}
	}

	// 2. Blokady (jeśli typeFilter="" lub "block" lub "unblock")
	if typeFilter == "" || typeFilter == "block" || typeFilter == "unblock" {
		blockQuery := "SELECT timestamp, ip, reason, CAST(score AS TEXT), status FROM blocked_ips WHERE 1=1"
		blockArgs := []interface{}{}

		switch typeFilter {
		case "block":
			blockQuery += " AND status='blocked'"
		case "unblock":
			blockQuery += " AND status='unblocked'"
		}

		if search != "" {
			blockQuery += " AND (ip LIKE ? OR reason LIKE ?)"
			searchPattern := "%" + search + "%"
			blockArgs = append(blockArgs, searchPattern, searchPattern)
		}

		blockQuery += " ORDER BY timestamp DESC LIMIT ?"
		blockArgs = append(blockArgs, limit)

		blockRows, err := s.db.Query(blockQuery, blockArgs...)
		if err != nil {
			return nil, err
		}
		defer blockRows.Close()

		for blockRows.Next() {
			var entry ActivityEntry
			var status string
			if err := blockRows.Scan(&entry.Timestamp, &entry.IP, &entry.Details, &entry.Extra, &status); err != nil {
				return nil, err
			}
			if status == "blocked" {
				entry.Type = "block"
			} else {
				entry.Type = "unblock"
			}
			entries = append(entries, entry)
		}
	}

	// 3. Whitelist (tylko jeśli typeFilter="" lub "whitelist_add")
	if typeFilter == "" || typeFilter == "whitelist_add" {
		wlQuery := "SELECT added_at, ip, description FROM whitelist WHERE 1=1"
		wlArgs := []interface{}{}

		if search != "" {
			wlQuery += " AND (ip LIKE ? OR description LIKE ?)"
			searchPattern := "%" + search + "%"
			wlArgs = append(wlArgs, searchPattern, searchPattern)
		}

		wlQuery += " ORDER BY added_at DESC LIMIT ?"
		wlArgs = append(wlArgs, limit)

		wlRows, err := s.db.Query(wlQuery, wlArgs...)
		if err != nil {
			return nil, err
		}
		defer wlRows.Close()

		for wlRows.Next() {
			var entry ActivityEntry
			entry.Type = "whitelist_add"
			entry.Extra = ""
			if err := wlRows.Scan(&entry.Timestamp, &entry.IP, &entry.Details); err != nil {
				return nil, err
			}
			entries = append(entries, entry)
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp > entries[j].Timestamp
	})

	// Ogranicz do limitu (bo teraz mamy więcej niż limit)
	if len(entries) > limit {
		entries = entries[:limit]
	}

	return entries, nil
}

func (s *DbManager) Close() error {
	return s.db.Close()
}
