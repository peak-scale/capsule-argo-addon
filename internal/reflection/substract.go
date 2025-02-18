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
	// If either is a pointer, ensure both are pointers before recursing.
	if targetVal.Kind() == reflect.Ptr || sourceVal.Kind() == reflect.Ptr {
		// Only proceed if both are pointers.
		if targetVal.Kind() == reflect.Ptr && sourceVal.Kind() == reflect.Ptr {
			if !targetVal.IsNil() && !sourceVal.IsNil() {
				subtractRecursive(targetVal.Elem(), sourceVal.Elem())
			}
		}

		return
	}

	// Make sure we are working with structs.
	if targetVal.Kind() != reflect.Struct || sourceVal.Kind() != reflect.Struct {
		return
	}

	for i := range targetVal.NumField() {
		targetField := targetVal.Field(i)
		sourceField := sourceVal.Field(i)

		// Handle pointer fields.
		if targetField.Kind() == reflect.Ptr {
			// Ensure that the source field is also a pointer.
			if sourceField.Kind() != reflect.Ptr {
				// If types are mismatched, you might choose to skip or handle differently.
				continue
			}

			if !targetField.IsNil() && !sourceField.IsNil() {
				subtractRecursive(targetField.Elem(), sourceField.Elem())
			}

			continue
		}

		switch targetField.Kind() {
		case reflect.Struct:
			subtractRecursive(targetField, sourceField)
		case reflect.Slice:
			subtractSlices(targetField, sourceField)
		case reflect.Map:
			subtractMaps(targetField, sourceField)
		default:
			// For primitives, if they are equal, zero out the target.
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
