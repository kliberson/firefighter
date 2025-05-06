package sniffer

import (
	"bytes"
	"firefighter/internal/analyzer"
	"firefighter/internal/packetdata"
	"fmt"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func Start(interfaceName string) error {
	handle, err := pcap.OpenLive(interfaceName, 1600, true, pcap.BlockForever)
	if err != nil {
		return fmt.Errorf("nie mogę otworzyć interfejsu: %v", err)
	}
	defer handle.Close()

	// tymczasowe blokowanie ssh dla testów
	if err := handle.SetBPFFilter("not port 22"); err != nil {
		return fmt.Errorf("błąd ustawiania filtra BPF: %w", err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for packet := range packetSource.Packets() {
		info := ExtractPacketInfo(packet)
		if info != nil {
			// Logowanie do testów
			fmt.Printf("[%s] %s:%s → %s:%s (%s)\n",
				info.Timestamp.Format("15:04:05"),
				info.SrcIP, info.SrcPort,
				info.DstIP, info.DstPort,
				info.Protocol,
			)

			// Przekazanie do analizera
			analyzer.AnalyzePacket(info)
		}
	}
	return nil
}

func ExtractPacketInfo(packet gopacket.Packet) *packetdata.PacketInfo {
	ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	transportLayer := packet.TransportLayer()

	var (
		srcMAC, dstMAC   string
		srcIP, dstIP     string
		ttl, ipID, ipLen int
		srcPort, dstPort string
		tcpFlags         []string
	)

	// MAC
	if ethernetLayer != nil {
		eth := ethernetLayer.(*layers.Ethernet)
		srcMAC = eth.SrcMAC.String()
		dstMAC = eth.DstMAC.String()
	}

	// IP
	if ipLayer != nil {
		ip := ipLayer.(*layers.IPv4)
		srcIP = ip.SrcIP.String()
		dstIP = ip.DstIP.String()
		ttl = int(ip.TTL)
		ipID = int(ip.Id)
		ipLen = int(ip.Length)
	}

	// Transport (TCP/UDP)
	if transportLayer != nil {
		switch t := transportLayer.(type) {
		case *layers.TCP:
			srcPort = t.SrcPort.String()
			dstPort = t.DstPort.String()

			if t.SYN {
				tcpFlags = append(tcpFlags, "SYN")
			}
			if t.ACK {
				tcpFlags = append(tcpFlags, "ACK")
			}
			if t.FIN {
				tcpFlags = append(tcpFlags, "FIN")
			}
			if t.RST {
				tcpFlags = append(tcpFlags, "RST")
			}
		case *layers.UDP:
			srcPort = t.SrcPort.String()
			dstPort = t.DstPort.String()
		}
	}

	// Application layer
	appLayer := packet.ApplicationLayer()
	var protocol string
	var payloadPreview string

	if appLayer != nil {
		payload := appLayer.Payload()
		if len(payload) > 0 {
			if bytes.HasPrefix(payload, []byte("HTTP/")) || bytes.HasPrefix(payload, []byte("GET ")) {
				protocol = "HTTP"
			} else if bytes.HasPrefix(payload, []byte{0x16, 0x03}) {
				protocol = "TLS/SSL"
			} else {
				protocol = "Other"
			}

			// // Krótki podgląd (tekstowy, jeśli możliwe)
			// maxLen := 32
			// previewLen := min(len(payload), maxLen)
			// isText := true
			// for i := range previewLen {
			// 	if payload[i] < 32 || payload[i] > 126 {
			// 		isText = false
			// 		break
			// 	}
			// }
			// if isText {
			// 	payloadPreview = string(payload[:previewLen])
			// } else {
			// 	payloadPreview = fmt.Sprintf("%x", payload[:previewLen])
			// }
		}
		// Logowanie danych
		fmt.Printf("MAC: %s → %s | IP: %s → %s | %s → %s\n", srcMAC, dstMAC, srcIP, dstIP, srcPort, dstPort)
		fmt.Println(strings.Repeat("-", 60))
	}

	return &packetdata.PacketInfo{
		Timestamp:      packet.Metadata().Timestamp,
		SrcMAC:         srcMAC,
		DstMAC:         dstMAC,
		SrcIP:          srcIP,
		DstIP:          dstIP,
		SrcPort:        srcPort,
		DstPort:        dstPort,
		TTL:            ttl,
		IPID:           ipID,
		IPLen:          ipLen,
		TCPFlags:       tcpFlags,
		Protocol:       protocol,
		PayloadPreview: payloadPreview,
	}
}
