// package main

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"net/http"
// 	"net/url"
// 	"os"
// 	"time"
// )

// type PrometheusQueryResult struct {
// 	Status string `json:"status"`
// 	Data   struct {
// 		ResultType string `json:"resultType"`
// 		Result     []struct {
// 			Metric map[string]string `json:"metric"`
// 			Value  []interface{}     `json:"values"`
// 		} `json:"result"`
// 	} `json:"data"`
// }

// func main() {
// 	//err := service.Packet2Universe(1, true, "testformultistation", "1.2.3.4", "2024-10-28 03:01:00", 1, 1)
// 	//if err != nil {
// 	//	fmt.Println(err)
// 	//}
// 	//err = service.Packet2Universe(1, true, "testformultistation", "1.2.3.4", "2024-10-27 18:01:00", 1, 1)
// 	//if err != nil {
// 	//	fmt.Println(err)
// 	//}
// 	//err = service.Packet2Universe(1, true, "testformultistation", "1.2.3.4", "2024-10-27 23:01:00", 1, 1)
// 	//if err != nil {
// 	//	fmt.Println(err)
// 	//}

// 	//err = service.Packet2BaseStation(1, false, "2024-09-18 02:01:00", 1, 1)
// 	//if err != nil {
// 	//	fmt.Println(err)
// 	//}
// 	//err = service.Packet2BaseStation(1, false, "2024-09-18 01:01:00", 1, 1)
// 	//if err != nil {
// 	//	fmt.Println(err)
// 	//}
// 	//s := Controllers.SqlController{}
// 	//p1, p2, lat, lng, err := s.TransferLocationInfo("8.130.125.140")
// 	//if err != nil {
// 	//	fmt.Println(err)
// 	//}
// 	//fmt.Println(p1, p2, lat, lng)
// 	// PrometheusQueryResult represents the structure of the Prometheus query result.

// 	// Prometheus server address and port
// 	prometheusServer := "http://192.168.225.133:9090" // 修改为你的Prometheus服务器地址

// 	// Query parameters
// 	query := "node_network_speed_bytes"

// 	step := "15s"
// 	// Build the query URL
// 	queryURL, err := url.Parse(prometheusServer)
// 	if err != nil {
// 		fmt.Printf("Error parsing Prometheus server URL: %v\n", err)
// 		os.Exit(1)
// 	}
// 	queryURL.Path = "/api/v1/query_range"
// 	params := queryURL.Query()
// 	//params.Set("query", query)
// 	params.Set("start", time.Now().Add(-time.Second*30).Format(time.RFC3339)) // 设置开始时间（这里为了简单使用当前时间减去1小时，但通常应该使用固定的时间范围字符串，如"1h"）
// 	params.Set("end", time.Now().Format(time.RFC3339))                        // 设置结束时间（当前时间）
// 	// 注意：上面的start和end参数实际上在Prometheus的/api/v1/query中不是必需的，因为可以通过query字符串中的时间范围选择器（如[1h]）来指定。
// 	// 但为了演示如何构建URL参数，这里还是包含了它们。在实际查询中，应该只使用query参数中的时间范围选择器。
// 	// 因此，下面我们将注释掉start和end参数的设置，并在query字符串中直接使用[1h]。
// 	// params.Set("start", timeRangeStart.Format(time.RFC3339))
// 	// params.Set("end", timeRangeEnd.Format(time.RFC3339))
// 	//queryURL.RawQuery = params.Encode()

// 	// Correct way to specify time range in the query string
// 	//queryWithTimeRange := fmt.Sprintf("%s[%s]", query, timeRange)
// 	params.Set("query", query)
// 	params.Set("step", step)
// 	queryURL.RawQuery = params.Encode()
// 	fmt.Println("Query URL:", queryURL.String())

// 	// Send the query request
// 	resp, err := http.Get(queryURL.String())
// 	if err != nil {
// 		fmt.Printf("Error querying Prometheus: %v\n", err)
// 		os.Exit(1)
// 	}
// 	defer resp.Body.Close()

// 	// Read the response body
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		fmt.Printf("Error reading Prometheus response: %v\n", err)
// 		os.Exit(1)
// 	}

// 	// Parse the response body
// 	var result PrometheusQueryResult
// 	err = json.Unmarshal(body, &result)
// 	if err != nil {
// 		fmt.Printf("Error unmarshaling Prometheus response: %v\n", err)
// 		os.Exit(1)
// 	}

// 	// Print the query result
// 	if result.Status == "success" {
// 		fmt.Println("Query Result:")
// 		for _, r := range result.Data.Result {
// 			fmt.Printf("Metric: %v, Value: %v\n, Length: %d\n", r.Metric, r.Value, len(r.Value))
// 		}
// 	} else {
// 		fmt.Printf("Query failed: %s\n", result)
// 	}
// }
