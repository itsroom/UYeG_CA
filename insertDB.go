package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func InsertDB() {
	for {
		timeInsert()
		algorithmInsert()

		time.Sleep(1 * time.Second)
	}
}

func TimeQueryExec(sql string) {

	f, err := os.OpenFile("time.txt", os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		log.Println(err)
	}
	_, err = fmt.Fprintln(f, sql)
	if err != nil {
		log.Println(err)
		f.Close()
	}
	err = f.Close()
	if err != nil {
		log.Println(err)
	}

	f.Close()
}

func AlgorithmQueryExec(sql string) {

	f, err := os.OpenFile("algorithm.txt", os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		log.Println(err)
	}
	_, err = fmt.Fprintln(f, sql)
	if err != nil {
		log.Println(err)
		f.Close()
	}
	err = f.Close()
	if err != nil {
		log.Println(err)
	}
	f.Close()
}

func timeInsert() {
	bytes, err := ioutil.ReadFile("time.txt")
	if err != nil {
		log.Fatal(err)
	}
	if bytes != nil {
		s := string(bytes)
		if s != "" {
			startTime := time.Now()
			go dbConn.NotResultQueryExec(s)
			elapsedTime := time.Since(startTime)
			NTime := fmt.Sprint(time.Now())
			if elapsedTime >= time.Second*10 {
				fmt.Println(NTime[:21], "time 실행시간 : ", elapsedTime)
				err = ioutil.WriteFile("log.txt", []byte(s), 0644)
				if err != nil {
					panic(err)

				}
			}
			d := []byte("")
			err = ioutil.WriteFile("time.txt", d, 0644)
			if err != nil {
				panic(err)
			}
		}
	}
}

func Deletetxt() {
	d := []byte("")
	err := ioutil.WriteFile("time.txt", d, 0644)
	_ = ioutil.WriteFile("algorithm.txt", d, 0644)
	if err != nil {
		panic(err)
	}
}

func algorithmInsert() {
	bytes, err := ioutil.ReadFile("algorithm.txt")
	if err != nil {
		log.Fatal(err)
	}
	if bytes != nil {
		s := string(bytes)
		if s != "" {
			go dbConn.NotResultQueryExec(s)
			d := []byte("")
			err = ioutil.WriteFile("algorithm.txt", d, 0644)
			if err != nil {
				panic(err)
			}
		}
	}
}
