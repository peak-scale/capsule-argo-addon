package argo

import (
	"testing"

	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Test ValidateCSV function
func TestValidateCSV(t *testing.T) {
	validCSV := "p, role:admin, applications, get, *, allow"
	invalidCSV := "p role:admin applications get * allow"

	assert.NoError(t, ValidateCSV(validCSV), "Expected no error for valid CSV")
	assert.Error(t, ValidateCSV(invalidCSV), "Expected an error for invalid CSV")
}

// Test PolicyString function
func TestPolicyString(t *testing.T) {
	policy := "test-policy"
	tenant := "test-tenant"
	argopolicy := v1alpha1.ArgocdPolicyDefinition{
		Resource: "applications",
		Action:   []string{"get", "update"},
		Verb:     "allow",
		Path:     "*",
	}

	expectedResult := "p, test-policy,applications,get,test-tenant/*,allow\n" +
		"p, test-policy,applications,update,test-tenant/*,allow\n"
	result := PolicyString(policy, tenant, argopolicy)

	assert.Equal(t, expectedResult, result, "PolicyString should return correct policy string")
}

// Test BindingString function
func TestBindingString(t *testing.T) {
	subject := v1.Subject{Name: "test-user"}
	role := "admin"
	expectedResult := "g, test-user, admin\n"

	result := BindingString(subject, role)
	assert.Equal(t, expectedResult, result, "BindingString should return correct binding string")
}

// Test DefaultPolicies function
func TestDefaultPolicies(t *testing.T) {
	tenant := &capsulev1beta2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-tenant",
		},
	}

	expectedResult := "p, role:test-tenant:read-only,projects,get,test-tenant,allow\n" +
		"p, role:test-tenant:owner,projects,update,test-tenant,allow\n" +
		"p, role:test-tenant:read-only,clusters,get,test-tenant/*,allow\n" +
		"p, role:test-tenant:owner,clusters,update,test-tenant/*,allow\n"

	result := DefaultPolicies(tenant, true)
	assert.Equal(t, expectedResult, result, "DefaultPolicies should return correct default policies")
}

func TestDefaultPoliciesNoProxy(t *testing.T) {
	tenant := &capsulev1beta2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-tenant",
		},
	}

	expectedResult := "p, role:test-tenant:read-only,projects,get,test-tenant,allow\n" +
		"p, role:test-tenant:owner,projects,update,test-tenant,allow\n"

	result := DefaultPolicies(tenant, false)
	assert.Equal(t, expectedResult, result, "DefaultPolicies should return correct default policies")
}
