package mapStruct

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func MapStruct(from interface{}, to interface{}) {
	fromValue := reflect.ValueOf(from)
	toValue := reflect.ValueOf(to).Elem()

	for i := 0; i < fromValue.NumField(); i++ {
		fromField := fromValue.Field(i)
		toField := toValue.FieldByName(fromValue.Type().Field(i).Name)

		if toField.IsValid() && toField.CanSet() {
			toFieldType := toField.Type()

			if fromField.Type() == toFieldType {
				toField.Set(fromField)
			} else if fromField.Kind() == reflect.Struct && toFieldType.Kind() == reflect.Struct {
				MapStruct(fromField.Interface(), toField.Addr().Interface())
			} else if fromField.Kind() == reflect.Slice && toFieldType.Kind() == reflect.Slice {
				mapSlice(fromField, toField)
			} else {
				mapField(fromField, toField)
			}
		}
	}
}

func mapField(fromField reflect.Value, toField reflect.Value) {
	// Dereference the pointers if they are pointers
	if fromField.Kind() == reflect.Ptr {
		fromField = fromField.Elem()
	}

	if toField.Kind() == reflect.Ptr {
		toField = toField.Elem()
	}

	if !fromField.IsValid() || !toField.IsValid() {
		return
	}

	toFieldType := toField.Type()
	fromFieldType := fromField.Type()

	switch toFieldType.Kind() {
	case reflect.String:
		if fromFieldType.Kind() == reflect.Int || fromFieldType.Kind() == reflect.Int8 || fromFieldType.Kind() == reflect.Int16 || fromFieldType.Kind() == reflect.Int32 || fromFieldType.Kind() == reflect.Int64 {
			toField.SetString(strconv.FormatInt(fromField.Int(), 10))
		} else if fromFieldType.Kind() == reflect.Uint || fromFieldType.Kind() == reflect.Uint8 || fromFieldType.Kind() == reflect.Uint16 || fromFieldType.Kind() == reflect.Uint32 || fromFieldType.Kind() == reflect.Uint64 {
			toField.SetString(strconv.FormatUint(fromField.Uint(), 10))
		} else if fromFieldType.Kind() == reflect.Float32 || fromFieldType.Kind() == reflect.Float64 {
			toField.SetString(strconv.FormatFloat(fromField.Float(), 'f', -1, 64))
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if fromFieldType.Kind() == reflect.String {
			fromVal := fromField.String()
			regex := regexp.MustCompile(`\.0+$|,`)
			fromVal = regex.ReplaceAllString(fromVal, "")
			intVal, err := strconv.ParseInt(fromVal, 10, toFieldType.Bits())
			if err == nil {
				toField.SetInt(intVal)
			}
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if fromFieldType.Kind() == reflect.String {
			fromVal := fromField.String()
			fromVal = strings.Replace(fromVal, ",", "", -1)
			floatVal, err := strconv.ParseFloat(fromVal, toFieldType.Bits())
			if err == nil {
				uintVal := uint64(floatVal)
				toField.SetUint(uintVal)
			}
		} else if fromFieldType.Kind() == reflect.Interface {
			// Convert the interface value to uint64
			var uintVal uint64
			switch interfaceType := fromField.Interface().(type) {
			case int, int8, int16, int32, int64:
				uintVal = uint64(interfaceType.(int64))
			case uint, uint8, uint16, uint32, uint64:
				uintVal = uint64(interfaceType.(uint64))
			case float32:
				uintVal = uint64(interfaceType)
			case float64:
				uintVal = uint64(interfaceType)
			case string:
				floatVal, err := strconv.ParseFloat(interfaceType, toFieldType.Bits())
				if err == nil {
					uintVal = uint64(floatVal)
				}
			}
			toField.SetUint(uintVal)
		}
	case reflect.Float32, reflect.Float64:
		if fromFieldType.Kind() == reflect.String {
			fromVal := fromField.String()
			fromVal = strings.Replace(fromVal, ",", "", -1)
			floatVal, err := strconv.ParseFloat(fromVal, toFieldType.Bits())
			if err == nil {
				toField.SetFloat(floatVal)
			}
		}
	}
}

func mapSlice(fromField reflect.Value, toField reflect.Value) {
	fromElemType := fromField.Type().Elem()
	toElemType := toField.Type().Elem()

	// Specific case for converting an array of strings to an array of uint
	if fromElemType.Kind() == reflect.String && toElemType.Kind() == reflect.Uint {
		for i := 0; i < fromField.Len(); i++ {
			strVal := fromField.Index(i).String()
			uintVal, err := strconv.ParseUint(strVal, 10, toElemType.Bits())
			if err == nil {
				toField.Set(reflect.Append(toField, reflect.ValueOf(uintVal).Convert(toElemType)))
			}
		}
		return
	}

	// Specific case for converting an array of strings to an array of float32
	if fromElemType.Kind() == reflect.String && toElemType.Kind() == reflect.Float32 {
		for i := 0; i < fromField.Len(); i++ {
			strVal := fromField.Index(i).String()
			floatVal, err := strconv.ParseFloat(strVal, toElemType.Bits())
			if err == nil {
				toField.Set(reflect.Append(toField, reflect.ValueOf(float32(floatVal)).Convert(toElemType)))
			}
		}
		return
	}

	// General case for converting slices
	if fromElemType == toElemType {
		toField.Set(fromField)
		return
	}

	if fromElemType.Kind() == reflect.Struct && toElemType.Kind() == reflect.Struct {
		for i := 0; i < fromField.Len(); i++ {
			fromElem := fromField.Index(i)
			toElem := reflect.New(toElemType).Elem()
			MapStruct(fromElem.Interface(), toElem.Addr().Interface())
			toField.Set(reflect.Append(toField, toElem))
		}
		return
	}
}
