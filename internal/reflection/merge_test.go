// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package reflection

import (
	"reflect"
	"testing"
)

type ItemStruct struct {
	Name  string
	Value int
}

type TestStructWithSliceOfStructs struct {
	Tester      bool
	Labels      map[string]string
	Annotations map[string]string
	Finalizers  []string
	Items       []ItemStruct
}

func TestMerge_SuccessWithStructSlices(t *testing.T) {
	target := TestStructWithSliceOfStructs{
		Tester: false,
		Labels: map[string]string{
			"app": "example",
			"env": "prod",
		},
		Annotations: map[string]string{
			"maintainer": "team1",
		},
		Finalizers: []string{
			"finalizer1",
		},
		Items: []ItemStruct{
			{Name: "item1", Value: 10},
			{Name: "item2", Value: 20},
		},
	}

	source := TestStructWithSliceOfStructs{
		Tester: true,
		Labels: map[string]string{
			"env":   "dev",
			"owner": "team2",
		},
		Annotations: map[string]string{
			"maintainer": "team2",
		},
		Finalizers: []string{
			"finalizer1", "finalizer2",
		},
		Items: []ItemStruct{
			{Name: "item2", Value: 20}, // Should not be duplicated
			{Name: "item3", Value: 30}, // New item to be added
		},
	}

	expected := TestStructWithSliceOfStructs{
		Tester: true,
		Labels: map[string]string{
			"app":   "example",
			"env":   "dev",   // 'prod' from target retained
			"owner": "team2", // 'owner' from source added
		},
		Annotations: map[string]string{
			"maintainer": "team2", // 'maintainer' from source overrides
		},
		Finalizers: []string{
			"finalizer1", "finalizer2", // 'finalizer2' from source added
		},
		Items: []ItemStruct{
			{Name: "item1", Value: 10}, // Original value retained
			{Name: "item2", Value: 20}, // Duplicates avoided
			{Name: "item3", Value: 30}, // New item added
		},
	}

	// Test the Merge function
	err := Merge(&target, &source)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Use reflect.DeepEqual to compare the merged struct with the expected result
	if !reflect.DeepEqual(target, expected) {
		t.Errorf("expected %+v, got %+v", expected, target)
	}
}
