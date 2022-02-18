package reflections

func GetInt(obj interface{}, name string) (int, error) {
	if v, err := GetField(obj, name); err != nil {
		return 0, err
	} else {
		return v.(int), nil
	}
}

func GetString(obj interface{}, name string) (string, error) {
	if v, err := GetField(obj, name); err != nil {
		return "", err
	} else {
		return v.(string), nil
	}
}
