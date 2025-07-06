// TODO: 实现空间、时间、流量分布的接口函数；在创建Universe后作为hook调用
package Controllers

import (
	"UserPortrait/configs"
	"UserPortrait/functions"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// LocInfo 位置信息结构体
type LocInfo struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Result  struct {
		Ip       string `json:"ip"`
		Location struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"location"`
		AdInfo struct {
			Nation     string `json:"nation"`
			NationCode int    `json:"nation_code"`
			Province   string `json:"province"`
			City       string `json:"city"`
			District   string `json:"district"`
			Adcode     int    `json:"adcode"`
		} `json:"ad_info"`
	} `json:"result"`
}

// GetLocation 调用腾讯地图API获取IP地址的位置信息
func GetLocation(ip string) (LocInfo, error) {
	url := fmt.Sprintf("/ws/location/v1/ip?ip=%s&key=%s", ip, configs.TencentMapKey)
	sig := functions.GetMD5Hash(url + configs.TencentSK)
	requestURL := fmt.Sprintf("https://apis.map.qq.com%s&sig=%s", url, sig)
	method := "GET"

	payload := strings.NewReader("")
	req, err := http.NewRequest(method, requestURL, payload)

	if err != nil {
		errinfo := fmt.Errorf("err in GetLocation/NewRequest:%v", err)
		return LocInfo{}, errinfo
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		errinfo := fmt.Errorf("err in GetLocation/client:%v", err)
		return LocInfo{}, errinfo
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		errinfo := fmt.Errorf("err in GetLocation/readall:%v", err)
		return LocInfo{}, errinfo
	}
	var locinfo = LocInfo{}
	err = json.Unmarshal(body, &locinfo)
	return locinfo, err
}
