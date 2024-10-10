package utils

import (
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

func ArgoPolicyName(tenant *capsulev1beta2.Tenant) string {
	return "policy." + tenant.Name + ".csv"
}
