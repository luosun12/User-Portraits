// TODO: 实现空间、时间、流量分布的接口函数；在创建Universe后作为hook调用
package Controllers

import (
	"UserPortrait/configs"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// 定义结构体，作为位置信息的接收器
type LocInfo struct {
	Code string `json:"code"`
	Data struct {
		Continent string `json:"continent"`
		Country   string `json:"country"`
		Zipcode   string `json:"zipcode"`
		Timezone  string `json:"timezone"`
		Accuracy  string `json:"accuracy"`
		Owner     string `json:"owner"`
		Isp       string `json:"isp"`
		Source    string `json:"source"`
		Areacode  string `json:"areacode"`
		Adcode    string `json:"adcode"`
		Asnumber  string `json:"asnumber"`
		Lat       string `json:"lat"`
		Lng       string `json:"lng"`
		Radius    string `json:"radius"`
		Prov      string `json:"prov"`
		City      string `json:"city"`
		District  string `json:"district"`
	} `json:"data"`
	Charge   bool   `json:"charge"`
	Msg      string `json:"msg"`
	Ip       string `json:"ip"`
	Coordsys string `json:"coordsys"`
}

// 调用IP2Location API获取位置信息
func GetLocation(ip string) (LocInfo, error) {
	url := "https://eolink.o.apispace.com/ipguishu/ip/geo/v1/district?ip=" + ip
	method := "GET"

	payload := strings.NewReader("")
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		errinfo := fmt.Errorf("err in GetLocation/NewRequest:%v", err)
		return LocInfo{}, errinfo
	}
	req.Header.Add("X-APISpace-Token", configs.APISpaceKey)

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
