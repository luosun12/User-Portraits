package process

import (
	"UserPortrait/etc"
	"UserPortrait/service"
	"fmt"
	"strings"
	"sync"
	"time"

	"UserPortrait/parsePacket/capture"
)

const connectionTimeout = 60 * time.Second

type ConnectionInfo struct {
	SynTime     time.Time
	SynAckTime  time.Time
	AckTime     time.Time
	Flow        uint
	LossFlag    bool
	StationID   uint
	MAC         string
	SourceIP    string
	DestIP      string
	DestPort    uint16
	Latency     uint
	LastSeqNums map[uint32]struct{} // 存储序列号
	mux         sync.Mutex
}

var connectionMap sync.Map // 替换为 sync.Map

// 处理抓包信息
func processPacket(packet capture.PacketInfo) {
	tcpInfo := packet.TCPInfo
	if tcpInfo == nil {
		return
	}

	// 生成连接的唯一标识符
	connKey := fmt.Sprintf("%s:%d-%s:%d", packet.SourceIP, tcpInfo.SrcPort, packet.DestIP, tcpInfo.DstPort)

	// 从连接表中获取连接
	value, exists := connectionMap.Load(connKey)
	var conn *ConnectionInfo
	if exists {
		conn = value.(*ConnectionInfo)
	} else {
		conn = &ConnectionInfo{
			MAC:         packet.SrcMAC,
			SourceIP:    packet.SourceIP,
			DestIP:      packet.DestIP,
			DestPort:    tcpInfo.DstPort,
			StationID:   1,
			SynTime:     packet.Timestamp,
			LastSeqNums: make(map[uint32]struct{}),
		}
		connectionMap.Store(connKey, conn)
	}

	conn.LossFlag = false
	conn.Latency = 0
	now := packet.Timestamp

	conn.mux.Lock()
	defer conn.mux.Unlock()

	// 根据 TCP 标志位处理连接建立
	if strings.Contains(tcpInfo.Flags, "SYN") && !strings.Contains(tcpInfo.Flags, "ACK") {
		conn.SynTime = now
	} else if strings.Contains(tcpInfo.Flags, "SYN") && strings.Contains(tcpInfo.Flags, "ACK") {
		conn.SynAckTime = now
	} else if strings.Contains(tcpInfo.Flags, "ACK") && !conn.SynAckTime.IsZero() {
		conn.AckTime = now
		conn.Latency = uint(conn.AckTime.Sub(conn.SynTime).Milliseconds())
	}

	// 检查序列号以判断重传
	if _, exists := conn.LastSeqNums[tcpInfo.SeqNum]; exists {
		conn.LossFlag = true
	} else {
		conn.LastSeqNums[tcpInfo.SeqNum] = struct{}{}
	}

	packetDate := packet.Timestamp.Format("2006-01-02 15:04:05")
	fmt.Printf("%v:日期: %s, 连接信息: 基站ID: %d, MAC: %s, IP: %s, 流量: %d字节, 延迟: %d毫秒, 丢包标识: %t\n",
		etc.ParseInfo, packetDate, conn.StationID, conn.MAC, conn.SourceIP, tcpInfo.PayloadSize, conn.Latency, conn.LossFlag)
	err := service.Packet2Universe(conn.StationID, conn.LossFlag, conn.MAC, conn.SourceIP, packetDate, uint(tcpInfo.PayloadSize), conn.Latency)
	if err != nil {
		panic(err)
	}
	err = service.Packet2BaseStation(conn.StationID, conn.LossFlag, packetDate, uint(tcpInfo.PayloadSize), conn.Latency)
	if err != nil {
		panic(err)
	}
}

// 定时清除超时连接
func cleanStaleConnections() {
	ticker := time.NewTicker(1 * time.Minute) // 每分钟清理一次
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		connectionMap.Range(func(key, value interface{}) bool {
			conn := value.(*ConnectionInfo)
			if now.Sub(conn.SynTime) > connectionTimeout {
				connectionMap.Delete(key) // 删除超时连接
			}
			return true
		})
	}
}

// 主抓包处理函数
func CapturePackets() {
	go cleanStaleConnections()

	for {
		packetInfo := <-capture.PacketChannel
		processPacket(packetInfo)
	}
}
