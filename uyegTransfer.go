package main

import (
	"encoding/json"
	"fmt"

	"./uyeg"
)

func UYeGTransfer(client *uyeg.ModbusClient, tfChan <-chan []interface{}, chInsertData chan map[string]interface{}, chData chan map[string]interface{}) {
	for {
		select {
		case <-client.Done3:
			fmt.Println(fmt.Sprintf("=> %s (%s:%d) 데이터 전송 종료", client.Device.MacId, client.Device.Host, client.Device.Port))
			return
		case data := <-tfChan:
			d := data[0].(map[string]interface{})

			if t, exists := d["time"]; exists {
				bSecT := t.(string)[:len(TimeFormat)-4]
				jsonBytes := client.GetRemapJson(bSecT, data)

				dataSecond := make(map[string]interface{})
				json.Unmarshal(jsonBytes, &dataSecond)

				chInsertData <- dataSecond
				chData <- dataSecond
			}
		}
	}
}
