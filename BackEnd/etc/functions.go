package etc

import (
	"crypto/md5"
	"fmt"
	"math"
	"net"
	"strconv"
	"time"
)

// 选择基站或universe表名

func ChooseTable(station_id uint, MODE string) string {
	var tablename string
	if MODE == "universe" {
		switch station_id {
		case 1:
			tablename = "universe1"
		case 2:
			tablename = "universe2"
		case 3:
			tablename = "universe3"
		case 4:
			tablename = "universe4"
		}
	} else if MODE == "base_station" {
		switch station_id {
		case 1:
			tablename = "base_station1"
		case 2:
			tablename = "base_station2"
		case 3:
			tablename = "base_station3"
		case 4:
			tablename = "base_station4"
		}
	}
	return tablename
}

func ChooseStationLoc(station_id uint) (float32, float32) {
	switch station_id {
	case 1:
		return StationLocation1[0], StationLocation1[1]
	case 2:
		return StationLocation2[0], StationLocation2[1]
	case 3:
		return StationLocation3[0], StationLocation3[1]
	case 4:
		return StationLocation4[0], StationLocation4[1]
	default:
		return 0, 0
	}
}

//获取日期和时段编码

func GetPeriod(t string) (string, uint, error) {

	if t == "" {
		return "", 0, fmt.Errorf("time is empty")
	}
	var hour = t[11:13]
	var Id, _ = strconv.ParseUint(hour, 10, 64)
	periodId := uint(Id) + 1
	if periodId > 24 || periodId < 1 {
		return "", 0, fmt.Errorf("time is out of range")
	}
	return t[0:10], periodId, nil
}

// 获取当前近二十四小时时段信息,查询时满足：1.lastDate的lastId~24；2.currDate的1~currId

func GetDailyInfo() (string, string, uint, uint, error) {
	lastDate := time.Now().AddDate(0, 0, -1).Format(time.DateOnly)
	t := time.Now().String()
	currDate, currId, err := GetPeriod(t)
	if err != nil {
		return "", "", 0, 0, err
	}
	var lastId uint
	if currId == 24 {
		lastId = 1
	} else {
		lastId = currId + 1
	}
	return lastDate, currDate, lastId, currId, nil
}

// 保留部分小数位，并不改变类型

func RoundToFloat32(f float64, n int) float32 {
	power := math.Pow(10, float64(n))
	return float32(math.Floor(f*power+0.5) / power) // 加0.5后取整，模拟四舍五入
}

// 获取md5编码哈希

func GetMD5Hash(input string) string {
	h := md5.New()
	h.Write([]byte(input))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// 获取本机的WLAN IP

func GetLocalIP() string {
	// 获取所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Error retrieving network interfaces:", err)
		return ""
	}

	for _, iface := range interfaces {
		// 仅处理活动的接口
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
			addrs, err := iface.Addrs()
			if err != nil {
				fmt.Println("Error retrieving addresses:", err)
				continue
			}

			for _, addr := range addrs {
				// 过滤出 IPv4 地址
				if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.To4() != nil {
					// 判断是否为 WLAN 接口
					if iface.Name == "WLAN" { // 根据你的操作系统修改接口名称
						fmt.Println("WLAN IP Address:", ipNet.IP.String())
						return ipNet.IP.String()
					}
				}
			}
		}
	}
	fmt.Println("No WLAN IP address found.")
	return ""
}
