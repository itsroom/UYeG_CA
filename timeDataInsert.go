package main

import (
	"fmt"
	"strings"
	"time"
)

var (
	timeInterval = [5]int64{1000, 10000, 60000, 600000, 3600000}
	timeTable    = [5]string{"CdHisItemSecond", "CdHisItem10Seconds", "CdHisItemMinute", "CdHisItem10Minutes", "CdHisItemHour"}
	dataTime     [6]map[string]float64
)

func timeDataQuery(data map[string]interface{}) {
	date := data["time"].(string)
	if strings.Contains(data["time"].(string), "UTC") {
		return
	}
	t, err := time.Parse("2006-01-02 15:04:05.000", date)
	if err != nil {
		panic(err)
	}
	nowSec := t.UnixNano() / 1000000

	_tagName := tagName

	for i := 0; i < len(timeInterval); i++ {
		myMapHandler := &SyncFloatMap{v: dataTime[i]}

		var (
			insertKey   string
			insertValue string
			Prevmac     string
		)

		for j := 0; j < len(_tagName); j++ {
			tempValue := saveSolutionFunc(dataTime[i][_tagName[j]], data[_tagName[j]], mapMstDevice[_tagName[j]].(map[string]interface{})["SaveSolution"].(string))
			myMapHandler.FloatSet(_tagName[j], tempValue)
			if nowSec%timeInterval[i] == 0 {
				s := strings.Split(_tagName[j], ".")

				if Prevmac == "" {
					Prevmac = s[0]
				}

				if (j == len(_tagName)-1) && dataTime[i][_tagName[j]] > mapMstDevice[_tagName[j]].(map[string]interface{})["Offset"].(float64) {
					insertKey = fmt.Sprintf("%s, %s", insertKey, s[1])
					insertValue = fmt.Sprintf("%s, '%f'", insertValue, dataTime[i][_tagName[j]])
				}

				if s[0] != Prevmac || j == len(_tagName)-1 {
					sqlStr := fmt.Sprintf("INSERT IGNORE INTO %s (`DataSavedTime`,`mac`%s) VALUES ('%s', '%s'%s);", timeTable[i], insertKey, date, Prevmac, insertValue)
					if len(insertKey) != 0 {
						go TimeQueryExec(sqlStr)
					}
					Prevmac = ""
					insertKey = ""
					insertValue = ""
				}

				if j != len(_tagName)-1 && dataTime[i][_tagName[j]] > mapMstDevice[_tagName[j]].(map[string]interface{})["Offset"].(float64) {
					insertKey = fmt.Sprintf("%s, `%s`", insertKey, s[1])
					insertValue = fmt.Sprintf("%s, '%f'", insertValue, dataTime[i][_tagName[j]])
				}
				myMapHandler.FloatSet(_tagName[j], 0)
			}

		}
	}
}

func saveSolutionFunc(curr interface{}, prev interface{}, saveSolution string) float64 {
	currfloat, err := getFloat(curr)
	if err != nil {
		panic(err)
	}
	prevfloat, _ := getFloat(prev)

	if saveSolution == "MAX" {
		if currfloat > prevfloat {
			return currfloat
		}
		return prevfloat

	} else if saveSolution == "MIN" {
		if (currfloat < prevfloat && currfloat != 0) || prevfloat == 0 {
			return currfloat
		}
		return prevfloat

	} else if saveSolution == "AVG" {
		return (currfloat + prevfloat) / 2

	} else if saveSolution == "SUM" {
		return currfloat + prevfloat

	} else if saveSolution == "CUR" {
		return currfloat
	}
	return currfloat
}
