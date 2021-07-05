package configuration

func GetBool(value *bool, defaultValue bool) bool {
	if value != nil {
		return *value
	}
	return defaultValue
}

func GetInt(value *int, defaultValue int) int {
	if value != nil {
		return *value
	}
	return defaultValue
}

func GetString(value *string, defaultValue string) string {
	if value != nil {
		return *value
	}
	return defaultValue
}

func CopyOf(originalMap map[string]map[string]string) map[string]map[string]string {
	targetMap := make(map[string]map[string]string, len(originalMap))
	for key, value := range originalMap {
		innerMap := make(map[string]string, len(value))
		for k, v := range value {
			innerMap[k] = v
		}
		targetMap[key] = innerMap
	}
	return targetMap
}
