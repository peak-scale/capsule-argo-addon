package utils

import (
	"encoding/json"
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

// Check slice if it contains a string
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func YamlToJSON(yamlBytes []byte) ([]byte, error) {
	var yamlObj map[string]interface{}
	err := yaml.Unmarshal(yamlBytes, &yamlObj)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling yaml: %w", err)
	}

	// Convert the YAML object to JSON
	jsonBytes, err := json.Marshal(yamlObj)
	if err != nil {
		return nil, fmt.Errorf("error marshaling to json: %w", err)
	}

	return jsonBytes, nil
}
func Mapify(data interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	v := reflect.ValueOf(data)

	// If the provided data is a pointer, resolve to the underlying value
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return result // Return empty map for nil pointers
		}
		v = v.Elem()
	}

	// Ensure we're working with a struct
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)

			// Skip unexported fields
			if field.PkgPath != "" {
				continue
			}

			value := v.Field(i)
			// Handle different types with recursive or base handling
			switch value.Kind() {
			case reflect.Ptr:
				if !value.IsNil() {
					result[field.Name] = Mapify(value.Interface())
				}
			case reflect.Struct:
				result[field.Name] = Mapify(value.Interface())
			case reflect.Slice:
				var slice []interface{}
				for j := 0; j < value.Len(); j++ {
					item := value.Index(j)
					if item.Kind() == reflect.Struct {
						slice = append(slice, Mapify(item.Interface()))
					} else {
						slice = append(slice, item.Interface())
					}
				}
				result[field.Name] = slice
			case reflect.Map:
				mapResult := make(map[string]interface{})
				for _, key := range value.MapKeys() {
					mapResult[fmt.Sprint(key)] = value.MapIndex(key).Interface()
				}
				result[field.Name] = mapResult
			default:
				result[field.Name] = value.Interface()
			}
		}
	}
	return result
}
