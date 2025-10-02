package suricata

import (
	"container/list"
	"time"
)

type Alert struct {
	Timestamp string `json:"timestamp"`
	EventType string `json:"event_type"`
	SrcIP     string `json:"src_ip"`
	SrcPort   int    `json:"src_port"`
	DstIP     string `json:"dest_ip"`
	DstPort   int    `json:"dest_port"`
	Proto     string `json:"proto"`
	Alert     struct {
		Signature   string `json:"signature"`
		Category    string `json:"category"`
		Severity    int    `json:"severity"`
		SignatureID int    `json:"signature_id"`
	} `json:"alert"`

	ParsedTime time.Time `json:"-"`
}

// Last alerts from SlidingWindow
type SlidingWindow struct {
	Duration time.Duration
	Events   *list.List
}

// Manages multiple SlidingWindows by source IP
type WindowManager struct {
	Duration time.Duration
	Windows  map[string]*SlidingWindow
}
