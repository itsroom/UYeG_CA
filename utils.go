package main

import (
	"fmt"
	"time"

	"./db"
	"./uyeg"
)

var Loc, _ = time.LoadLocation("Asia/Seoul")

var TimeFormat = "2006-01-02 15:04:05.000"

func GetEnabledDevices(dbConn *db.DataBase) map[int]uyeg.Device {
	rows, err := dbConn.Conn.Query(`
	SELECT id,GATEWAY_ID, MAC_ID, NAME, HOST, PORT, UNIT_ID, REMAP_VERSION, PROCESS_INTERVAL, RETRY_CYCLE, RETRY_COUNT, RETRY_CONN_FAILED_COUNT
	FROM gateway WHERE ENABLED=True;
	`)

	if err != nil {
		fmt.Println(err.Error())
		return map[int]uyeg.Device{}
	}

	ds := map[int]uyeg.Device{}
	for rows.Next() {
		var device uyeg.Device
		err := rows.Scan(
			&device.Id,
			&device.GatewayId,
			&device.MacId,
			&device.Name,
			&device.Host,
			&device.Port,
			&device.UnitId,
			&device.Version,
			&device.ProcessInterval,
			&device.RetryCycle,
			&device.RetryCount,
			&device.RetryConnFailedCount,
		)

		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		ds[device.Id] = device
	}

	return ds
}
