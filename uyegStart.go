package main

import (
	"fmt"

	"./uyeg"
)

func UYeGStartFunc(client *uyeg.ModbusClient, chRawData chan map[string]interface{}) {
	defer func() {
		v := recover()

		if v != nil {
			derr := make(map[string]interface{})
			derr["Device"] = client.Device
			derr["Error"] = v
			derr["Restart"] = true

			ErrChan <- derr
		}
	}()

	if !client.Connect() {
		derr := make(map[string]interface{})
		derr["Device"] = client.Device
		derr["Error"] = fmt.Sprintf("%s(%s): Connection failed", client.Device.Name, client.Device.MacId)
		derr["Restart"] = false

		ErrChan <- derr
	}

	collChan := make(chan map[string]interface{}, 20)
	tfChan := make(chan []interface{}, 20)
	chInsertData := make(chan map[string]interface{})
	chData := make(chan map[string]interface{})

	go influxDataInsert(chInsertData)
	go UYeGTransfer(client, tfChan, chInsertData, chData)
	go UYeGProcessing(client, collChan, tfChan)
	go UYeGDataCollection(client, collChan)

	for {
		select {
		case data := <-chData:
			chRawData <- data
		}
	}
}
