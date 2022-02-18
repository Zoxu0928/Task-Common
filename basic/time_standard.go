package basic

// 定义时间类型

import (
	"database/sql/driver"
	"encoding/xml"
	"time"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

type TimeStandard time.Time

func NewTimeStandard(str string) (TimeStandard, error) {
	t := TimeStandard{}
	var err error
	if str != "" && str != "null" {
		err = t.UnmarshalJSON([]byte(`"` + str + `"`))
	}
	return t, err
}

// Json 时间类型反序列化
func (t *TimeStandard) UnmarshalJSON(data []byte) error {
	now, time_err := unmarshalJSONToJson(data)
	if time_err != nil {
		return time_err
	}
	if !now.IsZero() {
		now = now.Local()
		*t = TimeStandard(now)
	}
	return nil
}

// Json 时间类型序列化
func (t TimeStandard) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("\"\""), nil
	}
	b := make([]byte, 0, len(TIME_FORMART)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, TIME_FORMART)
	b = append(b, '"')
	return b, nil
}

func (t TimeStandard) String() string {
	if t.IsZero() {
		return ""
	}
	// fmt.Println("----", t.Time().UTC().Format("2006-01-02T15:04:05Z")) 转换为UTC时间
	return t.Time().Format(TIME_FORMART)
}

func (t TimeStandard) Time() time.Time {
	return time.Time(t)
}

func (t TimeStandard) IsZero() bool {
	return time.Time(t).IsZero()
}

func (t TimeStandard) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	e.EncodeElement(t.String(), start)
	return nil
}

// 数据库字段反序列化
func (t *TimeStandard) Scan(val interface{}) error {
	if val == nil {
		return nil
	}
	str := string(val.([]byte))
	if str == "\"\"" || str == "null" || str == "0000-00-00 00:00:00" {
		return nil
	}
	now, err := time.ParseInLocation(TIME_FORMART, str, time.Local)
	if err != nil {
		return err
	}
	*t = TimeStandard(now)
	return nil
}

// 数据库字段序列化
func (t TimeStandard) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.String(), nil
}
