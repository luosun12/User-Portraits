package capture

import (
	"UserPortrait/configs"
	"UserPortrait/functions"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type PacketInfo struct {
	Device       string
	Timestamp    time.Time
	SrcMAC       string
	DstMAC       string
	EthType      string
	SourceIP     string
	DestIP       string
	Protocol     string
	TCPInfo      *TCPInfo
	UDPInfo      *UDPInfo
	HTTPInfo     *HTTPInfo
	PacketLength int
}

var PacketChannel = make(chan PacketInfo, 100)
var packetHashSet = make(map[string]struct{})
var snapshotLen int32 = 1024
var promiscuous bool = true
var timeout time.Duration = 30 * time.Second
var filter string = "tcp or udp or (ip and (port 80 or port 443))"
var excludeIP = []string{
	configs.DBHost[:13],
	functions.GetLocalIP(),
}

func capturePackets(device string) {
	handle, err := pcap.OpenLive(device, snapshotLen, promiscuous, timeout)
	if err != nil {
		log.Fatalf("Error opening device %s: %v", device, err)
	}
	defer handle.Close()

	if err := handle.SetBPFFilter(filter); err != nil {
		log.Fatalf("Error setting BPF filter: %v", err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		srcMAC, dstMAC, ethType := extractEthernetInfo(packet)
		if srcMAC == "" || dstMAC == "" {
			continue
		}

		srcIP, dstIP, protocol := extractIPInfo(packet)
		if srcIP == "" || dstIP == "" || protocol == "" || (srcIP == excludeIP[0] || dstIP == excludeIP[0] || srcIP == excludeIP[1]) {
			continue
		}

		tcpInfo := extractTCPInfo(packet)
		udpInfo := extractUDPInfo(packet)
		httpInfo := parseHTTP(packet)

		packetHash := getPacketHash(packet)
		if _, exists := packetHashSet[packetHash]; exists {
			continue
		}

		packetHashSet[packetHash] = struct{}{}

		packetInfo := PacketInfo{
			Device:       device,
			Timestamp:    packet.Metadata().Timestamp,
			SrcMAC:       srcMAC,
			DstMAC:       dstMAC,
			EthType:      ethType,
			SourceIP:     srcIP,
			DestIP:       dstIP,
			Protocol:     protocol,
			TCPInfo:      tcpInfo,
			UDPInfo:      udpInfo,
			HTTPInfo:     httpInfo,
			PacketLength: len(packet.Data()),
		}

		PacketChannel <- packetInfo
	}
}

func getPacketHash(packet gopacket.Packet) string {
	hash := md5.Sum(packet.Data())
	return hex.EncodeToString(hash[:])
}

func extractEthernetInfo(packet gopacket.Packet) (string, string, string) {
	ethLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethLayer != nil {
		eth, _ := ethLayer.(*layers.Ethernet)
		srcMAC := eth.SrcMAC.String()
		dstMAC := eth.DstMAC.String()
		if srcMAC == "" || srcMAC == "00:00:00:00:00:00" || dstMAC == "" || dstMAC == "00:00:00:00:00:00" {
			return "", "", ""
		}
		return srcMAC, dstMAC, fmt.Sprintf("0x%X", eth.EthernetType)
	}
	return "", "", ""
}

func extractIPInfo(packet gopacket.Packet) (string, string, string) {
	ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
	if ipv4Layer != nil {
		ip, _ := ipv4Layer.(*layers.IPv4)
		return ip.SrcIP.String(), ip.DstIP.String(), ip.Protocol.String()
	}

	ipv6Layer := packet.Layer(layers.LayerTypeIPv6)
	if ipv6Layer != nil {
		ip, _ := ipv6Layer.(*layers.IPv6)
		return ip.SrcIP.String(), ip.DstIP.String(), ip.NextHeader.String()
	}

	return "", "", ""
}

type TCPInfo struct {
	SrcPort, DstPort            uint16
	SeqNum, AckNum, PayloadSize uint32
	Flags                       string
}

func extractTCPInfo(packet gopacket.Packet) *TCPInfo {
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		tcp, _ := tcpLayer.(*layers.TCP)
		flags := []string{}
		if tcp.FIN {
			flags = append(flags, "FIN")
		}
		if tcp.SYN {
			flags = append(flags, "SYN")
		}
		if tcp.RST {
			flags = append(flags, "RST")
		}
		if tcp.PSH {
			flags = append(flags, "PSH")
		}
		if tcp.ACK {
			flags = append(flags, "ACK")
		}
		if tcp.URG {
			flags = append(flags, "URG")
		}
		return &TCPInfo{
			SrcPort:     uint16(tcp.SrcPort),
			DstPort:     uint16(tcp.DstPort),
			SeqNum:      tcp.Seq,
			AckNum:      tcp.Ack,
			PayloadSize: uint32(len(tcp.Payload)),
			Flags:       strings.Join(flags, "|"),
		}
	}
	return nil
}

type UDPInfo struct {
	SrcPort, DstPort    uint16
	Length, PayloadSize uint16
}

func extractUDPInfo(packet gopacket.Packet) *UDPInfo {
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer != nil {
		udp, _ := udpLayer.(*layers.UDP)
		return &UDPInfo{
			SrcPort:     uint16(udp.SrcPort),
			DstPort:     uint16(udp.DstPort),
			Length:      udp.Length,
			PayloadSize: uint16(len(udp.Payload)),
		}
	}
	return nil
}

type HTTPInfo struct {
	RequestMethod string
	StatusCode    int
	ContentType   string
	UserAgent     string
}

func parseHTTP(packet gopacket.Packet) *HTTPInfo {
	appLayer := packet.ApplicationLayer()
	if appLayer != nil && strings.Contains(string(appLayer.Payload()), "HTTP") {
		payload := string(appLayer.Payload())
		lines := strings.Split(payload, "\n")
		httpInfo := &HTTPInfo{}

		for _, line := range lines {
			if strings.HasPrefix(line, "GET") || strings.HasPrefix(line, "POST") {
				httpInfo.RequestMethod = strings.Fields(line)[0]
			} else if strings.HasPrefix(line, "HTTP") {
				statusCodeStr := strings.Fields(line)[1]
				httpInfo.StatusCode, _ = strconv.Atoi(statusCodeStr)
			}

			if strings.HasPrefix(line, "Content-Type:") {
				httpInfo.ContentType = strings.TrimSpace(strings.TrimPrefix(line, "Content-Type:"))
			}
			if strings.HasPrefix(line, "User-Agent:") {
				httpInfo.UserAgent = strings.TrimSpace(strings.TrimPrefix(line, "User-Agent:"))
			}
		}
		return httpInfo
	}
	return nil
}

func Tcpd() {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
	}

	for _, device := range devices {
		go capturePackets(device.Name)
	}
}
