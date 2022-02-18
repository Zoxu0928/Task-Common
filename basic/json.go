/*
 * Copyright (c) 2020 Hui Dong <donghui6@jd.com>
 */

package basic

import "encoding/json"

func ForceMarshalToStr(obj interface{}) string {
	v, _ := json.Marshal(obj)
	if v == nil {
		return ""
	}
	return string(v)
}

func ForceMarshalToPrettyStr(obj interface{}) string {
	v, _ := json.MarshalIndent(obj, "", "    ")
	return string(v)
}
