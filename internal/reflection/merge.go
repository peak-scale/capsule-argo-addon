// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package reflection

import (
	"fmt"
	"reflect"

	"dario.cat/mergo"
)

// Merge handles merging two structs, with custom handling for slices to avoid duplicates.
func Merge(target, source interface{}) error {
	// Ensure both inputs are pointers
	targetVal := reflect.ValueOf(target)
	sourceVal := reflect.ValueOf(source)

	// Check if target and source are pointers
	if targetVal.Kind() != reflect.Ptr || sourceVal.Kind() != reflect.Ptr {
		return fmt.Errorf("both target and source must be pointers to structs")
	}

	// Use Elem() to access the struct the pointers refer to
	targetVal = targetVal.Elem()
	sourceVal = sourceVal.Elem()

	// Now targetVal and sourceVal should be structs
	if targetVal.Kind() != reflect.Struct || sourceVal.Kind() != reflect.Struct {
		return fmt.Errorf("both target and source must be pointers to structs")
	}

	// Handle structs
	mergeRecursive(targetVal, sourceVal)

	// Use mergo to handle non-slice fields
	if err := mergo.Merge(target, source); err != nil {
		return err
	}

	return nil
}

func mergeRecursive(targetVal, sourceVal reflect.Value) {
	for i := 0; i < targetVal.NumField(); i++ {
		targetField := targetVal.Field(i)
		sourceField := sourceVal.Field(i)

		switch targetField.Kind() {
		case reflect.Struct:
			// Recurse for nested structs
			mergeRecursive(targetField, sourceField)
		case reflect.Slice:
			// Handle slices to avoid duplicates
			mergeSlices(targetField, sourceField)
		case reflect.Map:
			// Handle maps (optional: add custom logic if needed)
			mergeMaps(targetField, sourceField)
		}
	}
}

// mergeSlices appends unique elements from the source slice to the target slice.
func mergeSlices(targetField, sourceField reflect.Value) {
	uniqueItems := make(map[string]bool)

	// Helper function to generate a unique key for struct elements
	generateKey := func(item reflect.Value) string {
		// Convert the item to a string representation
		return fmt.Sprintf("%#v", item.Interface())
	}

	// Retain all existing items from the target slice
	for i := 0; i < targetField.Len(); i++ {
		targetItem := targetField.Index(i)
		uniqueItems[generateKey(targetItem)] = true
	}

	// Prepare a slice for the merged result
	mergedSlice := reflect.MakeSlice(targetField.Type(), 0, targetField.Len()+sourceField.Len())

	// Add all unique items from the target slice first
	for i := 0; i < targetField.Len(); i++ {
		mergedSlice = reflect.Append(mergedSlice, targetField.Index(i))
	}

	// Append unique items from the source slice to the merged slice
	for i := 0; i < sourceField.Len(); i++ {
		sourceItem := sourceField.Index(i)
		if !uniqueItems[generateKey(sourceItem)] {
			mergedSlice = reflect.Append(mergedSlice, sourceItem)
			uniqueItems[generateKey(sourceItem)] = true
		}
	}

	// Set the merged result back to the target field
	targetField.Set(mergedSlice)
}

// mergeMaps merges maps without overriding existing keys in the target.
func mergeMaps(targetField, sourceField reflect.Value) {
	for _, key := range sourceField.MapKeys() {
		sourceValue := sourceField.MapIndex(key)
		targetField.SetMapIndex(key, sourceValue) // Overwrite target with source
	}
}
