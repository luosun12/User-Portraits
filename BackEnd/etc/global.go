package etc

var (
	Periods = map[int]string{1: "0~1", 2: "1~2", 3: "2~3", 4: "3~4", 5: "4~5", 6: "5~6", 7: "6~7", 8: "7~8",
		9: "8~9", 10: "9~10", 11: "10~11", 12: "11~12", 13: "12~13", 14: "13~14", 15: "14~15", 16: "15~16",
		17: "16~17", 18: "17~18", 19: "18~19", 20: "19~20", 21: "20~21", 22: "21~22", 23: "22~23", 24: "23~24"}
)

var (
	UniverseChannel = make(chan Universe, 100)
	StationChannel  = make(chan BaseStation, 100)
)

var (
	StationLocation1 = []float32{39.9042, 116.4074}
	StationLocation2 = []float32{39.9042, 116.4074}
	StationLocation3 = []float32{39.9042, 116.4074}
	StationLocation4 = []float32{39.9042, 116.4074}
)

// 终端字体颜色
const (
	Red     = "\033[31m" // 红色
	Green   = "\033[32m" // 绿色
	Yellow  = "\033[33m" // 黄色
	Blue    = "\033[34m" // 蓝色
	Magenta = "\033[35m" // 紫色
	Cyan    = "\033[36m" // 青色
	Reset   = "\033[0m"  // 重置
)

const (
	LoginErr    = Red + "[Login Error]:" + Reset
	RegisterErr = Red + "[Register Error]:" + Reset
	ParseInfo   = Cyan + "[Parse Info]:" + Reset
)
