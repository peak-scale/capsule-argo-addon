// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0
//

package reflection

import "reflect"

func Subtract(target, source interface{}) {
	subtractRecursive(reflect.ValueOf(target).Elem(), reflect.ValueOf(source).Elem())
}

//nolint:exhaustive
func subtractRecursive(targetVal, sourceVal reflect.Value) {
	for i := range targetVal.NumField() {
		targetField := targetVal.Field(i)
		sourceField := sourceVal.Field(i)

		// Handle different types
		switch targetField.Kind() {
		case reflect.Struct:
			// Recurse for nested structs
			subtractRecursive(targetField, sourceField)
		case reflect.Slice:
			// Handle slices
			subtractSlices(targetField, sourceField)
		case reflect.Map:
			// Handle maps
			subtractMaps(targetField, sourceField)
		default:
			// Handle primitive types
			if reflect.DeepEqual(targetField.Interface(), sourceField.Interface()) {
				targetField.Set(reflect.Zero(targetField.Type()))
			}
		}
	}
}

func subtractSlices(targetField, sourceField reflect.Value) {
	resultSlice := reflect.MakeSlice(targetField.Type(), 0, targetField.Len())

	for i := range targetField.Len() {
		targetItem := targetField.Index(i)

		found := false

		for j := range sourceField.Len() {
			sourceItem := sourceField.Index(j)
			if reflect.DeepEqual(targetItem.Interface(), sourceItem.Interface()) {
				found = true

				break
			}
		}

		// Only keep items that are not in the source slice
		if !found {
			resultSlice = reflect.Append(resultSlice, targetItem)
		}
	}

	targetField.Set(resultSlice)
}

func subtractMaps(targetField, sourceField reflect.Value) {
	for _, key := range sourceField.MapKeys() {
		targetValue := targetField.MapIndex(key)
		sourceValue := sourceField.MapIndex(key)

		// Remove matching map entries
		if reflect.DeepEqual(targetValue.Interface(), sourceValue.Interface()) {
			targetField.SetMapIndex(key, reflect.Value{})
		}
	}
}
