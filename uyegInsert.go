package main

import (
	"log"
	"reflect"
	"strings"
	"time"

	"./db"
	client "github.com/Heo-youngseo/influxdb1-client/v2"
)

var influxdb = db.Influx{
	Database: conf.INFLUX_DATABASE,
	User:     conf.INFLUX_USER,
	Password: conf.INFLUX_PASSWORD,
}

func influxDataInsert(chInserData chan map[string]interface{}) {
	for {
		select {
		case <-chInserData:
			c := influxDBClient()
			createMetrics(c, chInserData)
		}
	}
}

func influxDBClient() client.Client {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://localhost:8086/",
		Username: influxdb.User,
		Password: influxdb.Password,
		Timeout:  10 * time.Second,
	})
	if err != nil {
		log.Fatalln("Error: ", err)
	}
	return c
}

func createMetrics(c client.Client, chInserData chan map[string]interface{}) {

	for {
		select {
		case data := <-chInserData:
			bp, err := client.NewBatchPoints(client.BatchPointsConfig{
				Database:  influxdb.Database,
				Precision: "ms",
			})

			if err != nil {
				log.Fatalln("Error: ", err)
			}

			values := data["Values"]

			keySec := orderKey(data)
			tempStrSec := strings.Join(keySec[:], ",")
			tempStrSec = strings.Replace(tempStrSec, ",time", "", 1)
			tempStrSec = strings.Replace(tempStrSec, ",ver", "", 1)
			tempStrSec = strings.Replace(tempStrSec, ",gateway", "", 1)
			tempStrSec = strings.Replace(tempStrSec, ",mac", "", 1)
			tempStrSec = strings.Replace(tempStrSec, ",Values", "", 1)
			arrKeySec := strings.Split(tempStrSec, ",")

			switch reflect.TypeOf(values).Kind() {
			case reflect.Slice:
				s := reflect.ValueOf(values)

				keyMilli := orderKey(s.Index(0).Interface().(map[string]interface{}))
				tempStrMilli := strings.Join(keyMilli[:], ",")
				tempStrMilli = strings.Replace(tempStrMilli, "time", "DataSavedTime", 1)
				tempStrMilli = strings.Replace(tempStrMilli, "420", "`420`", 1)

				for i := 0; i < s.Len(); i++ {
					dataMilli := s.Index(i).Interface().(map[string]interface{})
					tags := map[string]string{
						"mac":     data["mac"].(string),
						"gateway": data["gateway"].(string),
					}

					fields := make(map[string]interface{})
					for j := 0; j < len(arrKeySec); j++ {
						fields[arrKeySec[j]] = data[arrKeySec[j]]
					}

					for j := 0; j < len(keyMilli); j++ {
						fields[keyMilli[j]] = dataMilli[keyMilli[j]]
					}

					date := dataMilli["time"].(string)
					t, err := time.Parse("2006-01-02 15:04:05.000", date)
					point, err := client.NewPoint(
						"SmartEOCR",
						tags,
						fields,
						t,
					)

					if err != nil {
						log.Fatalln("Error: ", err)
					}

					bp.AddPoint(point)
				}
			}

			go func() {
				err = c.Write(bp)
				if err != nil {
					log.Fatal(err)
				}
			}()
		}
	}
}
