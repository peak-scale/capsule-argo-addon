package argo

import (
	"github.com/peak-scale/capsule-argo-addon/internal/utils"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

func ArgoPolicyName(tenant *capsulev1beta2.Tenant) string {
	return "policy." + utils.TenantProjectName(tenant) + ".csv"
}
