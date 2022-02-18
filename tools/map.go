package tools

// DeepCopy 深拷贝
func DeepCopy(value interface{}) interface{} {
	if valueMap, ok := value.(map[string]interface{}); ok {
		newMap := make(map[string]interface{})
		for k, v := range valueMap {
			newMap[k] = DeepCopy(v)
		}

		return newMap
	} else if valueSlice, ok := value.([]interface{}); ok {
		newSlice := make([]interface{}, len(valueSlice))
		for k, v := range valueSlice {
			newSlice[k] = DeepCopy(v)
		}

		return newSlice
	} else if stringMap, ok := value.(map[string]string); ok {
		newMap := make(map[string]string)
		for k, v := range stringMap {
			newMap[k] = v
		}

		return newMap
	}

	return value
}
