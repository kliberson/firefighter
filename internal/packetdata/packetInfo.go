package packetdata

import "time"

type PacketInfo struct {
	Timestamp        time.Time
	SrcMAC, DstMAC   string
	SrcIP, DstIP     string
	SrcPort, DstPort string
	TTL, IPID, IPLen int
	TCPFlags         []string
	Protocol         string
	PayloadPreview   string
}
