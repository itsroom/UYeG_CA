package main

import (
	"fmt"
	//"log"
	"math"
	//"os"
	"strings"
	"time"

	"./uyeg"
)

func UYeGProcessing(client *uyeg.ModbusClient, collChan <-chan map[string]interface{}, tfChan chan<- []interface{}) {

	var queue ItemQueue
	if queue.items == nil {
		queue = ItemQueue{}
		queue.New()
	}

	go QueueProcess(client, &queue, tfChan)

	for {
		select {
		case <-client.Done2:
			fmt.Println(fmt.Sprintf("=> %s (%s:%d) 데이터 처리 종료", client.Device.MacId, client.Device.Host, client.Device.Port))
			return
		case data := <-collChan:
			queue.Enqueue(data)

		}
		time.Sleep(1 * time.Millisecond)
	}
}

func QueueProcess(client *uyeg.ModbusClient, queue *ItemQueue, tfChan chan<- []interface{}) {
	syncMap := SyncMap{v: make(map[string]interface{})}
	ds := make([]interface{}, 0, 100)
	var lastData map[string]interface{}

	for {
		for len(queue.items) > 0 {
			data := (*queue).Dequeue()
			t := (*data).(map[string]interface{})["time"].(time.Time).Truncate(time.Duration(client.Device.ProcessInterval) * time.Millisecond).Format(TimeFormat)
			if v := syncMap.Get(t); v != nil {
				tv := make(map[string]interface{})
				for k, v := range v.(map[string]interface{}) {
					var tmp float64
					if strings.Contains(k, "time") {
						tv[k] = v
						continue
					} else if strings.Contains(k, "Volt") {
						tmp = math.Min(v.(float64), (*data).(map[string]interface{})[k].(float64))
					} else {
						tmp = math.Max(v.(float64), (*data).(map[string]interface{})[k].(float64))
					}
					tv[k] = tmp
				}
				syncMap.Set(t, tv)
			} else {
				(*data).(map[string]interface{})["time"] = t
				syncMap.Set(t, (*data).(map[string]interface{}))

				tmillisecond := t[len(t)-4:]
				t2, _ := time.Parse(TimeFormat[:len(TimeFormat)-4], t)
				if tmillisecond == ".000" && syncMap.Size() >= 10 {
					sMap := syncMap.GetMap()
					bSecT, _ := time.Parse(TimeFormat[:len(TimeFormat)-4], t2.Add(-1 * time.Second).Format(TimeFormat)[:len(TimeFormat)-4])

					for i := 0; i < 1000/client.Device.ProcessInterval; i++ {
						vT := bSecT.Add(time.Duration(i*client.Device.ProcessInterval) * time.Millisecond).Format(TimeFormat)
						if val, exists := sMap[vT]; exists == true {
							value := val.(map[string]interface{})
							value["status"] = true
							ds = append(ds, value)
							lastData = val.(map[string]interface{})
							syncMap.Delete(vT)

						} else {
							if lastData != nil {
								ld := CopyMap(lastData)
								ld["time"] = vT
								ld["status"] = false
								ds = append(ds, ld)
							}

							fmt.Println(" No Data ", vT)

							derr := make(map[string]interface{})
							derr["Device"] = client.Device
							derr["Error"] = fmt.Sprintf(" No Data ", vT)
							derr["Restart"] = false

							ErrChan <- derr
						}
					}

					tfChan <- ds
					ds = ds[:0]
				}
			}
			time.Sleep(1 * time.Millisecond)
		}
		time.Sleep(1 * time.Millisecond)
	}
}
