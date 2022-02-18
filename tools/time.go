package tools

import (
	"time"
)

func SplitDay(startTime, endTime time.Time) *[]string {
	list := []string{}
	for startTime.Before(endTime) {
		list = append(list, startTime.Format("2006-01-02"))
		startTime = startTime.Add(24 * time.Hour)
	}
	return &list
}

func SplitHour(startTime, endTime time.Time) *[]string {
	list := []string{}
	for startTime.Before(endTime) {
		list = append(list, startTime.Format("2006-01-02 15"))
		startTime = startTime.Add(time.Hour)
	}
	return &list
}

func CheckTimeFormat(layout, str string) error {
	_, err := time.Parse(layout, str)
	if err != nil {
		return err
	}
	return nil
}

func IsEarlier(layout, start, end string) bool {
	var startTime, endTime *time.Time
	var err1, err2 error
	if result1, err1 := time.Parse(layout, start); err1 == nil {
		startTime = &result1
	}
	if result2, err2 := time.Parse(layout, end); err2 == nil {
		endTime = &result2
	}
	if err1 != nil || err2 != nil {
		return false
	}
	if startTime == nil || endTime == nil {
		return false
	}
	if startTime.Before(*endTime) {
		return true
	} else {
		return false
	}
}
