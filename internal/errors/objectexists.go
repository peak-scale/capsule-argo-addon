// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ObjectAlreadyExists represents an error indicating that an object already exists.
type ObjectAlreadyExistsError struct {
	Object client.Object
}

// Error implements the error interface for ObjectAlreadyExists.
func (e *ObjectAlreadyExistsError) Error() string {
	return fmt.Sprintf("object %s/%s of type %T already exists", e.Object.GetNamespace(), e.Object.GetName(), e.Object)
}

// NewObjectAlreadyExistsError creates a new ObjectAlreadyExists error with the provided Kubernetes client object.
func NewObjectAlreadyExistsError(obj client.Object) error {
	return &ObjectAlreadyExistsError{Object: obj}
}
