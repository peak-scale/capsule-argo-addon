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
	var yamlObj interface{}
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

			// If the field is a struct, recursively convert it
			if value.Kind() == reflect.Struct {
				result[field.Name] = Mapify(value.Interface())
			} else {
				result[field.Name] = value.Interface()
			}
		}
	}
	return result
}
