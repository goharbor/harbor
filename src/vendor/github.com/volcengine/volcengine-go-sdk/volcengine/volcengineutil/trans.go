package volcengineutil

import "strings"

func ParameterToMap(body string, sensitive []string, enable bool) map[string]interface{} {
	if !enable {
		return nil
	}
	result := make(map[string]interface{})
	params := strings.Split(body, "&")
	for _, param := range params {
		values := strings.Split(param, "=")
		if values[0] == "Action" || values[0] == "Version" {
			continue
		}
		v := values[1]
		if sensitive != nil && len(sensitive) > 0 {
			for _, s := range sensitive {
				if strings.Contains(values[0], s) {
					v = "****"
					break
				}
			}
		}
		result[values[0]] = v
	}
	return result
}

func BodyToMap(input map[string]interface{}, sensitive []string, enable bool) map[string]interface{} {
	if !enable {
		return nil
	}
	result := make(map[string]interface{})
loop:
	for k, v := range input {
		if len(sensitive) > 0 {
			for _, s := range sensitive {
				if strings.Contains(k, s) {
					v = "****"
					result[k] = v
					continue loop
				}
			}
		}
		var (
			next    map[string]interface{}
			nextPtr *map[string]interface{}
			ok      bool
		)

		if next, ok = v.(map[string]interface{}); ok {
			result[k] = BodyToMap(next, sensitive, enable)
		} else if nextPtr, ok = v.(*map[string]interface{}); ok {
			result[k] = BodyToMap(*nextPtr, sensitive, enable)
		} else {
			result[k] = v
		}
	}
	return result
}
