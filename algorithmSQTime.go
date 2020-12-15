package main

import (
	"fmt"
	"time"
)

func (input *SQTimeInput) SetSQTimeInput(_tagName string) {
	input.offset = mapMstDevice[_tagName].(map[string]interface{})["Offset"].(float64)
	input.resetPeriod = mapMstDevice[_tagName].(map[string]interface{})["ResetPeriod"].(float64) * 60000
	input.endCount = mapMstDevice[_tagName].(map[string]interface{})["EndCount"].(float64)
}

var sqtime map[string]*SQTimeProcessVariable

func (pv *SQTimeProcessVariable) sqtimeExtraction(date string, _tagValue float64, _tagName string) {
	input := new(SQTimeInput)
	input.SetSQTimeInput(_tagName)
	t, err := time.Parse("2006-01-02 15:04:05.000", date)
	if err != nil {
		panic(err)
	}

	_nowTime := t.UnixNano() / int64(time.Millisecond)

	if _tagValue > input.offset {
		pv.endCountStart = 0
		pv.waveformEnd = 0

		if pv.waveformStart == 0 {
			pv.waveformStart = _nowTime
		}

		if input.resetPeriod > 0 {
			if pv.resetStart == 0 {
				pv.resetStart = _nowTime
			}

			if pv.resetStart+int64(input.resetPeriod) < _nowTime {
				if pv.sumValues != 0 {
					pv.avgValues = pv.sumValues / float64(pv.cnt)
					strExtaction := fmt.Sprintf("Avg/%s/%f/%f/%s/", date, pv.avgValues, pv.sumValues, _tagName)
					outputExtraction(strExtaction)
					pv.maxPerValue, pv.maxPerTotalValue, pv.maxPerTime = timeAvgExtraction(pv.avgValues, pv.sumValues, _nowTime, pv.maxPerValue, pv.maxPerTotalValue, pv.maxPerTime, _tagName, date)
				}

				pv.waveformStart = 0
				pv.resetStart = 0
				pv.sumValues = 0
				pv.avgValues = 0
				pv.cnt = 0
			}
		}
		pv.sumValues += _tagValue
		pv.cnt++

	} else {
		if pv.endCountStart == 0 {
			pv.endCountStart = _nowTime
			pv.waveformEnd = _nowTime
		}

		if pv.endCountStart+int64(input.endCount) > _nowTime {
			return
		} else {
			pv.maxPerValue, pv.maxPerTotalValue, pv.maxPerTime = timeAvgExtraction(pv.avgValues, pv.sumValues, _nowTime, pv.maxPerValue, pv.maxPerTotalValue, pv.maxPerTime, _tagName, date)
		}

		if pv.sumValues != 0 {
			pv.avgValues = pv.sumValues / float64(pv.cnt)
			strExtaction := fmt.Sprintf("Avg/%s/%f/%f/%s/", date, pv.avgValues, pv.sumValues, _tagName)
			outputExtraction(strExtaction)
		}

		pv.waveformStart = 0
		pv.resetStart = 0
		pv.sumValues = 0
		pv.avgValues = 0
		pv.cnt = 0
	}
}

func timeAvgExtraction(maxValue float64, maxTotalValue float64, _nowTime int64, maxPerValue [4]float64, maxPerTotalValue [4]float64, maxPerTime [4]int64, _tagName string, date string) ([4]float64, [4]float64, [4]int64) {
	divTime := []int64{10000, 60000, 600000, 3600000}
	divTimeStr := []string{"10Seconds", "Minute", "10Minutes", "Hour"}

	for i := 0; i < len(divTime); i++ {
		t := time.Unix(0, maxPerTime[i]*int64(time.Millisecond))
		var t2 string
		if i == 0 {
			t1 := t.Add(-9 * time.Hour)
			t2 = t1.Add(1 * time.Minute).Format("2006-01-02 15:04:05.000")
		} else {
			t2 = t.Add(-9 * time.Hour).Format("2006-01-02 15:04:05.000")
		}

		if maxValue == 0 && maxPerTime[i] == 0 {
			continue
		}
		if maxPerTime[i] > _nowTime || maxPerTime[i] == 0 {
			if maxPerTime[i] == 0 {
				maxPerTime[i] = (_nowTime/divTime[i] + 1) * divTime[i]
			}

			if maxPerValue[i] < maxValue {
				maxPerValue[i] = maxValue
				maxPerTotalValue[i] = maxTotalValue
			}

		} else if maxPerTime[i] == _nowTime {
			if maxPerValue[i] < maxValue {
				maxPerValue[i] = maxValue
				maxPerTotalValue[i] = maxTotalValue
			}
			strExtaction := fmt.Sprintf("Avg/%s/%f/%f/%s/%s", t2, maxPerValue[i], maxPerTotalValue[i], _tagName, divTimeStr[i])
			outputExtraction(strExtaction)
			maxPerValue[i] = 0
			maxPerTotalValue[i] = 0
			maxPerTime[i] = 0

		} else {
			strExtaction := fmt.Sprintf("Avg/%s/%f/%f/%s/%s", t2, maxPerValue[i], maxPerTotalValue[i], _tagName, divTimeStr[i])
			outputExtraction(strExtaction)
			if maxValue != 0 {
				maxPerValue[i] = maxValue
				maxPerTotalValue[i] = maxTotalValue
				maxPerTime[i] = (_nowTime/divTime[i] + 1) * divTime[i]
			} else {
				maxPerValue[i] = 0
				maxPerTotalValue[i] = 0
				maxPerTime[i] = 0
			}
		}
	}
	return maxPerValue, maxPerTotalValue, maxPerTime
}
