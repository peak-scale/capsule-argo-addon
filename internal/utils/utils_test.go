// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsString(t *testing.T) {
	tests := []struct {
		slice    []string
		element  string
		expected bool
	}{
		{[]string{"apple", "banana", "cherry"}, "banana", true},
		{[]string{"apple", "banana", "cherry"}, "mango", false},
		{[]string{}, "banana", false},
	}

	for _, test := range tests {
		result := ContainsString(test.slice, test.element)
		assert.Equal(t, test.expected, result, "ContainsString failed")
	}
}

func TestYamlToJSON(t *testing.T) {
	yamlData := `
name: John
age: 30
address:
  city: New York
  zipcode: "10001"
`
	expectedJSON := `{"address":{"city":"New York","zipcode":"10001"},"age":30,"name":"John"}`

	jsonBytes, err := YamlToJSON([]byte(yamlData))
	assert.NoError(t, err, "Expected no error in YamlToJSON")
	assert.JSONEq(t, expectedJSON, string(jsonBytes), "Expected JSON output to match")
}

func TestMapify(t *testing.T) {
	type Address struct {
		City    string
		ZipCode string
	}

	type Person struct {
		Name    string
		Age     int
		Address Address
	}

	// Create a struct for testing
	person := Person{
		Name: "Alice",
		Age:  28,
		Address: Address{
			City:    "Wonderland",
			ZipCode: "12345",
		},
	}

	expectedMap := map[string]interface{}{
		"Name": "Alice",
		"Age":  28,
		"Address": map[string]interface{}{
			"City":    "Wonderland",
			"ZipCode": "12345",
		},
	}

	// Test Mapify function
	result := Mapify(person)
	assert.Equal(t, expectedMap, result, "Expected Mapify to return the correct map")
}
