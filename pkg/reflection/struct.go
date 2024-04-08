package reflection

import (
	"reflect"
)

func GetStringValue(input interface{}, fieldName string) (string, bool) {
	fieldValue, found := GetAnyValue(input, fieldName)
	if !found {
		return "", found
	}
	strValue, ok := fieldValue.(string)
	return strValue, ok
}

func GetAnyValue(input interface{}, fieldName string) (any, bool) {
	tf := reflect.TypeOf(input).Elem()
	vf := reflect.ValueOf(input).Elem()
	if tf.Kind() == reflect.Struct {
		for i := 0; i < tf.NumField(); i++ {
			if tf.Field(i).Name == fieldName {
				return vf.Field(i).Interface(), true
			}
		}
	}
	return nil, false

}
