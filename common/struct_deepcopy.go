package common

import "reflect"

// StructDeepCopy 定义       未完
func StructDeepCopy(srcStruct interface{}, desStruct interface{}) {
	if srcStruct == nil || desStruct == nil {
		return
	}

	input := reflect.ValueOf(srcStruct)
	input = removePtr(input)
	output := reflect.ValueOf(desStruct)
	output = removePtr(output)
	structValueDeepCopy(input, output)
	return
}

// StructValueDeepCopy 定义
func structValueDeepCopy(input reflect.Value, output reflect.Value) {
	inputType := input.Type()
	length := inputType.NumField()
	for i := 0; i < length; i++ {
		inputField := input.Field(i)
		outputField := output.FieldByName(inputType.Field(i).Name)

		if !outputField.IsValid() {
			continue
		}

		if !outputField.CanSet() {
			continue
		}
		inputFieldType := inputField.Type()
		outputFieldType := outputField.Type()
		inputFieldTypeKind := inputFieldType.Kind()
		outputFieldTypeKind := outputFieldType.Kind()
		if inputFieldTypeKind == reflect.Struct && outputFieldTypeKind == reflect.Struct {
			structValueDeepCopy(inputField, outputField)
			continue
		}
		if inputFieldType == outputFieldType {
			outputField.Set(inputField)
			continue
		}
	}
	return
}

func removePtr(input reflect.Value) (output reflect.Value) {
	if input.Type().Kind() == reflect.Ptr {
		output = input.Elem()
		output = removePtr(output)
	} else {
		output = input
	}
	return output
}
