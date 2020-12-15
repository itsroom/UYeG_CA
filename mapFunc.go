package main

import (
	"sync"
)

type SyncMap struct {
	v   map[string]interface{}
	mux sync.RWMutex
}

type SyncFloatMap struct {
	v   map[string]float64
	mux sync.RWMutex
}

type SyncArrMap struct {
	v   [6]map[string]float64
	mux sync.RWMutex
}

func (sm *SyncMap) Get(key string) interface{} {
	var value interface{}
	sm.mux.RLock()
	value = sm.v[key]
	sm.mux.RUnlock()
	return value
}

func (sm *SyncArrMap) Select(key string) bool {
	var found bool
	sm.mux.Lock()
	_, found = sm.v[0][key]
	sm.mux.Unlock()
	return found
}

func (sm *SyncMap) Set(key string, value interface{}) {
	sm.mux.Lock()
	sm.v[key] = value
	sm.mux.Unlock()
}

func (sm *SyncFloatMap) FloatSet(key string, value float64) {
	sm.mux.Lock()
	sm.v[key] = value
	sm.mux.Unlock()
}

func (sm *SyncMap) Delete(key string) {
	sm.mux.Lock()
	delete(sm.v, key)
	sm.mux.Unlock()
}

func (sm *SyncMap) GetMap() map[string]interface{} {
	sm.mux.RLock()
	value := CopyMap(sm.v)
	sm.mux.RUnlock()
	return value
}

func (sm *SyncMap) MoveMap(copyMap map[string]interface{}) {
	sm.mux.Lock()
	mapMstDevice = copyMap
	sm.mux.Unlock()
}

func CopyMap(originalMap map[string]interface{}) map[string]interface{} {
	newMap := make(map[string]interface{})
	for key, value := range originalMap {
		newMap[key] = value
	}
	return newMap
}

func (sm *SyncMap) Size() int {
	value := len(sm.v)
	return value
}
