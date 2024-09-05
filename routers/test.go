package main

import (
	"UserPortrait/service"
	"encoding/json"
	"fmt"
)

func main() {
	//err := service.Packet2Universe("newmac", "192.168.7.6", "2020-01-01 22:01:00", 1, 1)
	//if err != nil {
	//	println(err)
	//}
	locinfo, _ := service.GetLocation("8.130.125.140")
	loc, _ := json.MarshalIndent(locinfo, "", "  ")
	fmt.Println(string(loc))
}
