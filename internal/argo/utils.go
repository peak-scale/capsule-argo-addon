// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package argo

import (
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

func ArgoPolicyName(tenant *capsulev1beta2.Tenant) string {
	return "policy." + meta.TenantProjectName(tenant) + ".csv"
}
