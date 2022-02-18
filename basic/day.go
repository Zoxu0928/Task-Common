package basic

// 定义日期类型

import (
	"database/sql/driver"
	"encoding/xml"
	"github.com/Zoxu0928/task-common/logger"
	"time"
)

const (
	DAY_FORMART = "2006-01-02"
)

type Day time.Time

func NewDay(str string) (Day, error) {
	t := Day{}
	var err error
	if str != "" && str != "null" {
		err = t.UnmarshalJSON([]byte(`"` + str + `"`))
	}
	return t, err
}

// Json 时间类型反序列化
func (t *Day) UnmarshalJSON(data []byte) (err error) {
	str := string(data)
	if str == "\"\"" || str == "null" {
		return
	}
	now, err := time.ParseInLocation(`"`+DAY_FORMART+`"`, str, time.Local)
	if err != nil {
		logger.Error("day type convert error. %s", err)
	}
	*t = Day(now)
	return
}

// Json 时间类型序列化
func (t Day) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("\"\""), nil
	}
	b := make([]byte, 0, len(DAY_FORMART)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, DAY_FORMART)
	b = append(b, '"')
	return b, nil
}

func (t Day) String() string {
	if t.IsZero() {
		return ""
	}
	return time.Time(t).Format(DAY_FORMART)
}

func (t Day) Time() time.Time {
	return time.Time(t)
}

func (t Day) IsZero() bool {
	return time.Time(t).IsZero()
}

func (t Day) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	e.EncodeElement(t.String(), start)
	return nil
}

// 数据库字段反序列化
func (t *Day) Scan(val interface{}) (err error) {
	if val == nil {
		return
	}
	str := string(val.([]byte))
	if str == "\"\"" || str == "null" || str == "0000-00-00" {
		return
	}
	now, err := time.ParseInLocation(DAY_FORMART, str, time.Local)
	if err != nil {
		logger.Error("day type convert error. %s", err)
	}
	*t = Day(now)
	return
}

// 数据库字段序列化
func (t Day) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.String(), nil
}
