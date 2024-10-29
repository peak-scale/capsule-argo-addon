package errors

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ObjectAlreadyExists represents an error indicating that an object already exists
type ObjectAlreadyExists struct {
	Object client.Object
}

// Error implements the error interface for ObjectAlreadyExists
func (e *ObjectAlreadyExists) Error() string {
	return fmt.Sprintf("object %s/%s of type %T already exists", e.Object.GetNamespace(), e.Object.GetName(), e.Object)
}

// NewObjectAlreadyExistsError creates a new ObjectAlreadyExists error with the provided Kubernetes client object
func NewObjectAlreadyExistsError(obj client.Object) error {
	return &ObjectAlreadyExists{Object: obj}
}
