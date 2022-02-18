package tools

import (
	"sort"
)

// 获取deviceName，需要按从前到后，哪里有空位置就取哪个
// allDevice 表示全部的deviceName列表
// usedDevice 表示当前已占用的deviceName，获取的时候要排除这些
// need 表示需要获取几个deviceName
func GetDeviceName(allDevice map[string]int, usedDevice map[string]interface{}, need int) []string {

	sortedIdxList := make([]int, 0)
	idxMap := make(map[int]string)
	for k, v := range allDevice {
		sortedIdxList = append(sortedIdxList, v)
		idxMap[v] = k
	}
	sort.Ints(sortedIdxList)

	list := make([]string, 0)
	for i, idx := range sortedIdxList {
		if i == 0 {
			continue
		}
		dev := idxMap[idx]
		if usedDevice[dev] == nil {
			list = append(list, dev)
		}
		if len(list) == need {
			break
		}
	}
	return list
}
