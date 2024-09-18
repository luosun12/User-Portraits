/*
定义gorm格式MySQL数据表信息、各json格式化的接口等结构体
*/

package etc

type Userinfo struct {
	ID       uint       `gorm:"primary_key;auto_increment" json:"id"`
	Username string     `gorm:"type:varchar(16)" json:"username"`
	Password string     `gorm:"type:varchar(255)" json:"password"`
	MacInfo  string     `gorm:"type:varchar(32)" json:"mac_info"`
	Users    []Universe `gorm:"ForeignKey:UserID"`
}

type Admininfo struct {
	ID        uint   `gorm:"primary_key;auto_increment" json:"id"`
	Adminname string `gorm:"type:varchar(16)" json:"adminname"`
	Password  string `gorm:"type:varchar(255)" json:"password"`
}

type ContentType struct {
	ID      uint       `gorm:"primary_key;auto_increment" json:"id"`
	Content string     `gorm:"type:varchar(255)" json:"content-type"`
	Count   uint       `gorm:"type:int;default:1" json:"count"`
	Periods []Universe `gorm:"ForeignKey:ContentID;"`
}

type Universe struct {
	UserID    uint     `gorm:"primary_key;" json:"user_id"`
	Ip        string   `gorm:"type:char" json:"ip"`
	District  string   `gorm:"type:varchar" json:"district"`
	City      string   `gorm:"type:varchar" json:"city"`
	Latitude  float32  `gorm:"type:float" json:"latitude"`
	Longitude float32  `gorm:"type:float" json:"longitude"`
	PeriodID  uint     `gorm:"type:int;" json:"period_id"`
	Date      string   `gorm:"type:char;" json:"date"`
	Count     uint     `gorm:"type:int;default:1" json:"count"`
	Flow      uint     `gorm:"type:int;default:0" json:"flow"`
	Latency   uint     `gorm:"type:int;default:0" json:"latency"`
	ErrCount  uint     `gorm:"type:int;default:0" json:"err_count"`
	User      Userinfo `gorm:"ForeignKey:UserID;references:ID"`
}

type Interests struct {
	UserID    uint `gorm:"primary_key" json:"user_id"`
	ContentID uint `gorm:"primary_key" json:"ct_id"`
	Count     uint `gorm:"type:int;default:1" json:"count"`
}

type Score struct {
	UserID uint     `gorm:"primary_key" json:"user_id"`
	Score  float32  `gorm:"type:float;default:0" json:"score"`
	Date   string   `gorm:"type:char;" json:"date"`
	User   Userinfo `gorm:"ForeignKey:UserID;references:ID"`
}

type BaseStation struct {
	ConnCount  uint    `gorm:"default:1" json:"conn_count"`
	ErrCount   uint    `gorm:"default:0" json:"err_count"`
	Date       string  `json:"date"`
	PeriodID   uint    `json:"period_id"`
	TotalFlow  uint    `json:"total_flow"`
	AveLatency uint    `json:"ave_latency"`
	LossRate   float32 `json:"loss_rate"`
}

// Gorm的特殊方法，指定表名

func (u *Userinfo) TableName() string {
	return "user_info"
}

func (ad *Admininfo) TableName() string { return "admin_info" }

func (ct *ContentType) TableName() string {
	return "content_info"
}

func (uni *Universe) TableName() string {
	return "universe"
}

func (itr *Interests) TableName() string {
	return "content2user"
}

func (sc *Score) TableName() string {
	return "score"
}

func (bs *BaseStation) TableName() string { return "base_station" }

// 查询每日平均分结构体

type AverageScoreInterface struct {
	Date    string  `json:"date"`
	Average float32 `json:"average_score"`
}

// 获取基站区域信息

type StationInterface struct {
	CurrentPeriod uint `json:"current_period_id"`
	StationInfo   struct {
		StationID uint    `json:"station_id"`
		Latitute  float32 `json:"latitute"`
		Longitude float32 `json:"longitude"`
	} `json:"station_info"`
	Status []struct {
		PeriodID        uint    `json:"time_id"`
		ConnCount       uint    `json:"conn_quantity"`
		AverageSpeed    float32 `json:"average_speed"`
		AverageLatency  uint    `json:"average_latency"`
		AverageLossRate float32 `json:"average_packet_loss_rate"`
	} `json:"status"`
}
