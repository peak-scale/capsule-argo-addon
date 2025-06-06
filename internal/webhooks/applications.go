// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"
	"encoding/json"
	"net/http"

	argocdapi "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

// MutatingWebhook handles mutating webhook requests.
type ApplicationWebhook struct {
	Decoder admission.Decoder
	Client  client.Client
	Log     logr.Logger
}

// Handle processes the admission request and adds a label if necessary.
func (mw *ApplicationWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	mw.Log.V(7).Info("Received Request")
	// Only consider namespaced objects
	if req.Namespace == "" {
		return admission.Allowed("not namespaced object")
	}

	// Decode the object
	app := &argocdapi.Application{}
	if err := mw.Decoder.Decode(req, app); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	mw.Log.V(7).Info("looking up tenant for namespace", "namespace", app.GetNamespace())

	tntList := capsulev1beta2.TenantList{}
	if err := mw.Client.List(ctx, &tntList, client.MatchingFields{".status.namespaces": app.GetNamespace()}); err != nil {
		admission.Errored(http.StatusInternalServerError, err)
	}

	mw.Log.V(7).Info("retrieved tenants", "tenants", tntList)

	if len(tntList.Items) == 0 {
		return admission.Allowed("no tenant object")
	}

	tenant := tntList.Items[0]

	mw.Log.V(7).Info("matching tenant", "name", tenant.Name)

	// Only if Tenant is translated
	if !controllerutil.ContainsFinalizer(&tenant, meta.ControllerFinalizer) {
		return admission.Allowed("tenant not translated")
	}

	// Add the label if not present
	if app.Spec.Project == tenant.Name {
		mw.Log.V(7).Info("project already set to tenant")

		return admission.Allowed("tenant already set correctly")
	}

	// Overwrite Project
	app.Spec.Project = tenant.Name

	// Marshal the object back to JSON
	marshaledObj, err := json.Marshal(app)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledObj)
}
