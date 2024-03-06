package special

import (
	"reflect"

	"github.com/volcengine/volcengine-go-sdk/volcengine/response"
)

func iotResponse(response response.VolcengineResponse, i interface{}) interface{} {
	_, ok1 := reflect.TypeOf(i).Elem().FieldByName("ResponseMetadata")
	_, ok2 := reflect.TypeOf(i).Elem().FieldByName("Result")
	if ok1 && ok2 {
		return response
	}
	return response.Result
}
