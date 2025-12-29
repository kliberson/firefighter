package suricata

import (
	"time"
)

type Alert struct {
	Timestamp string `json:"timestamp"`
	FlowID    uint64 `json:"flow_id"`
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
