package etc

type Userinfo struct {
	ID       uint       `gorm:"primary_key;auto_increment" json:"id"`
	Username string     `gorm:"type:varchar(16)" json:"username"`
	Password string     `gorm:"type:varchar(255)" json:"password"`
	MacInfo  string     `gorm:"type:varchar(32)" json:"mac_info"`
	Users    []Universe `gorm:"ForeignKey:UserID"`
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
	Latitude  float64  `gorm:"type:float" json:"latitude"`
	Longitude float64  `gorm:"type:float" json:"longitude"`
	PeriodID  uint     `gorm:"type:int;" json:"period_id"`
	Date      string   `gorm:"type:char;" json:"date"`
	Count     uint     `gorm:"type:int;default:1" json:"count"`
	Flow      uint     `gorm:"type:int;default:0" json:"flow"`
	Latency   uint     `gorm:"type:int;default:0" json:"latency"`
	User      Userinfo `gorm:"ForeignKey:UserID;references:ID"`
}

type Interests struct {
	UserID    uint `gorm:"primary_key" json:"user_id"`
	ContentID uint `gorm:"primary_key" json:"ct_id"`
	Count     uint `gorm:"type:int;default:1" json:"count"`
}

type Score struct {
	UserID uint     `gorm:"primary_key" json:"user_id"`
	Score  float64  `gorm:"type:float;default:0" json:"score"`
	Date   string   `gorm:"type:char;" json:"date"`
	User   Userinfo `gorm:"ForeignKey:UserID;references:ID"`
}

// Gorm的特殊方法，指定表名

func (u *Userinfo) TableName() string {
	return "user_info"
}

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

// 查询每日平均分结构体

type AverageScore struct {
	Date    string  `json:"date"`
	Average float64 `json:"average_score"`
}