package data

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type DbManager struct {
	db *sql.DB
}

// DO SPRAWDZENIA CZY POTRZEBNE
// EDIT: BYŁO POTRZEBNE
type BlockedIPDetails struct {
	IP        string `json:"ip"`
	Reason    string `json:"reason"`
	Timestamp int64  `json:"timestamp"`
}

type WhitelistDetails struct {
	IP          string `json:"ip"`
	Description string `json:"description"`
	AddedAt     int64  `json:"added_at"`
}

// Struktura dla alertu
type AlertDetails struct {
	ID        int
	IP        string
	SID       int
	Message   string
	Timestamp time.Time
}

func New(path string) (*DbManager, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS blocked_ips (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            ip TEXT NOT NULL,
            reason TEXT,
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

func (s *DbManager) AddBlocked(ip, reason string) error {
	_, err := s.db.Exec(`INSERT INTO blocked_ips (ip, reason) VALUES (?, ?)`, ip, reason)
	return err
}

func (s *DbManager) AddAlert(ip string, sid int, message string) error {
	_, err := s.db.Exec(`INSERT INTO alerts (ip, sid, message) VALUES (?, ?, ?)`, ip, sid, message)
	return err
}

func (s *DbManager) GetBlocked() ([]BlockedIPDetails, error) {
	rows, err := s.db.Query(`SELECT ip, timestamp, reason 
							 FROM blocked_ips 
							 WHERE status='blocked'
							 ORDER BY timestamp DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ips []BlockedIPDetails
	for rows.Next() {
		var ip BlockedIPDetails
		if err := rows.Scan(&ip.IP, &ip.Timestamp, &ip.Reason); err != nil {
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

	// Sprawdź czy coś zaktualizowano
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows // IP nie było zablokowane
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

// DO SPRAWDZENIA CZY POTRZEBNE

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

// WHITELIST METHODS
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

func (s *DbManager) Close() error {
	return s.db.Close()
}
