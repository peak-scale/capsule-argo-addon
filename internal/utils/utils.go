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

func Mapify(input interface{}) interface{} {
	switch v := input.(type) {
	case map[interface{}]interface{}:
		// If it's a map with interface keys, convert to map[string]interface{}
		m := make(map[string]interface{})
		for key, value := range v {
			m[fmt.Sprintf("%v", key)] = Mapify(value) // Recursively mapify the value
		}
		return m
	case map[string]interface{}:
		// If it's already a map[string]interface{}, recursively process it
		for key, value := range v {
			v[key] = Mapify(value)
		}
		return v
	case []interface{}:
		// If it's a slice of interface{}, recursively process the elements
		for i, value := range v {
			v[i] = Mapify(value)
		}
		return v
	default:
		// For all other types, return as-is
		return v
	}
}

func ConvertStructToMap(data interface{}) map[string]interface{} {
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
				result[field.Name] = ConvertStructToMap(value.Interface())
			} else {
				result[field.Name] = value.Interface()
			}
		}
	}
	return result
}

func Flatten(input map[string]interface{}, prefix string) map[string]interface{} {
	flatMap := make(map[string]interface{})
	for key, value := range input {
		compoundKey := key
		if prefix != "" {
			compoundKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			// Recursively flatten nested maps
			nested := Flatten(v, compoundKey)
			for nestedKey, nestedValue := range nested {
				flatMap[nestedKey] = nestedValue
			}
		default:
			// Add non-map values to the flattened map
			flatMap[compoundKey] = value
		}
	}
	return flatMap
}
