package main

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"sync"
	"syscall"
	"time"

	"./config"
	"./db"
	"./uyeg"
)

var done = make(chan bool, 1)
var ErrChan = make(chan map[string]interface{}, 10)
var conf = config.GetConfiguration()
var dbConn = db.DataBase{
	Host:     conf.MYSQL_HOST,
	Port:     conf.MYSQL_PORT,
	Database: conf.MYSQL_DATABASE,
	User:     conf.MYSQL_USER,
	Password: conf.MYSQL_PASSWORD,
}
var gatewayID = conf.GATEWAYID

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("MAX PROCS", runtime.GOMAXPROCS(0))
	fmt.Println("\n===================")
	fmt.Println("Start Scada Program")
	fmt.Println("===================")

	wg := sync.WaitGroup{}
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	dbConn.Connect()
	defer dbConn.Close()
	wg.Add(1)
	go func() {
		<-sigs

		done <- true

		dbConn.Close()

		fmt.Println("\n===================")
		fmt.Println("Stop Scada Program")
		fmt.Println("===================")
		os.Exit(0)
		wg.Done()
	}()

	go startProgram()

	wg.Wait()
}

func startProgram() {
	addedDs := make(map[int]uyeg.Device)
	connDs := make(map[int]*uyeg.ModbusClient)
	chRawData := make(chan map[string]interface{})
	go preprocessing(chRawData)
	go InsertDB()

	for {
		select {
		case <-done:
			Deletetxt()
			for _, client := range connDs {
				client.Close()
			}
		case derr := <-ErrChan:
			id := derr["Device"].(uyeg.Device).Id
			if derr["Restart"].(bool) {
				connDs[id].Close()
				delete(connDs, id)
				delete(addedDs, id)
			}
		default:
			devices := GetEnabledDevices(&dbConn)
			if reflect.DeepEqual(addedDs, devices) {
				fmt.Println(time.Now().In(Loc).Format(TimeFormat), fmt.Sprintf(" - 모든 장치가 연결됨. (%d 개)", len(addedDs)))
			} else {
				fmt.Println(time.Now().In(Loc).Format(TimeFormat), " - 연결되지 않은 장치 또는 변경된 장치가 있음.")

				for id, device := range addedDs {
					if reflect.DeepEqual(devices[id], device) {
						continue
					}
					fmt.Println("=>", fmt.Sprintf("This device(%s) has been removed from the list.\n", device.MacId))
					connDs[id].Close()
					delete(connDs, id)
					delete(addedDs, id)
				}

				for id, device := range devices {
					if reflect.DeepEqual(addedDs[id], device) {
						continue
					}

					client := new(uyeg.ModbusClient)
					client.Device = device
					client.Done1 = make(chan bool)
					client.Done2 = make(chan bool)
					client.Done3 = make(chan bool)

					addedDs[id] = device
					connDs[id] = client

					go UYeGStartFunc(client, chRawData)

				}

			}

			time.Sleep(1 * time.Second)
		}
	}
}
