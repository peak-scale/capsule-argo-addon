// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	addonsv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/argo"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	capsulerbac "github.com/projectcapsule/capsule/pkg/api/rbac"
	capsulerules "github.com/projectcapsule/capsule/pkg/api/rules"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Argo RBAC permission sources", Serial, Label("rbac-permissions"), func() {
	const timeout = 2 * time.Minute

	var (
		tenant    *capsulev1beta2.Tenant
		selector  map[string]string
		configKey client.ObjectKey
	)

	// Each example owns its resources and waits for finalizers before finishing.
	create := func(object client.Object) {
		Expect(k8sClient.Create(context.Background(), object)).To(Succeed())
		DeferCleanup(func() {
			Expect(client.IgnoreNotFound(k8sClient.Delete(context.Background(), object))).To(Succeed())
			Eventually(func() bool {
				return apierrors.IsNotFound(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(object), object))
			}, timeout, defaultPollInterval).Should(BeTrue(), "resource %s should be deleted", object.GetName())
		})
	}

	updateTenant := func(update func(*capsulev1beta2.Tenant)) {
		Eventually(func() error {
			current := &capsulev1beta2.Tenant{}
			if err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(tenant), current); err != nil {
				return err
			}
			update(current)
			return k8sClient.Update(context.Background(), current)
		}, timeout, defaultPollInterval).Should(Succeed())
	}

	createOwner := func(subject rbacv1.Subject) *capsulev1beta2.TenantOwner {
		owner := &capsulev1beta2.TenantOwner{
			ObjectMeta: metav1.ObjectMeta{Name: tenant.Name + "-owner", Labels: selector},
			Spec: capsulev1beta2.TenantOwnerSpec{
				CoreOwnerSpec: capsulerbac.CoreOwnerSpec{
					UserSpec: capsulerbac.UserSpec{
						Kind: capsulerbac.OwnerKind(subject.Kind), Name: subject.Name,
					},
					ClusterRoles: []string{"edit"},
				},
			},
		}
		create(owner)
		return owner
	}

	expectStatusOwner := func(subject rbacv1.Subject, roles ...string) {
		Eventually(func(g Gomega) {
			current := &capsulev1beta2.Tenant{}
			g.Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(tenant), current)).To(Succeed())
			// This identity must only enter the addon through Capsule's resolved status.
			g.Expect(current.Spec.Owners).NotTo(ContainElement(HaveField("Name", subject.Name)))
			owner, found := current.Status.Owners.FindOwner(subject.Name, capsulerbac.OwnerKind(subject.Kind))
			g.Expect(found).To(Equal(len(roles) > 0), "resolved owner %s", subject.Name)
			g.Expect(owner.ClusterRoles).To(ConsistOf(roles))
		}, timeout, defaultPollInterval).Should(Succeed())
	}

	bindings := func(role string, subjects ...string) []string {
		var lines []string
		for _, subject := range subjects {
			lines = append(lines,
				fmt.Sprintf("g, %s, role:%s:%s", subject, tenant.Name, role),
				fmt.Sprintf("g, %s, caa:role:%s:read-only", subject, tenant.Name),
			)
			if role == "editor" {
				lines = append(lines, fmt.Sprintf("g, %s, caa:role:%s:owner", subject, tenant.Name))
			}
		}
		return lines
	}

	expectBindings := func(expected ...string) {
		Eventually(func(g Gomega) {
			configMap := &corev1.ConfigMap{}
			g.Expect(k8sClient.Get(context.Background(), configKey, configMap)).To(Succeed())
			policy, exists := configMap.Data[argo.ArgoPolicyName(tenant)]
			g.Expect(exists).To(BeTrue(), "tenant policy must exist, even when no subjects have access")
			reader := csv.NewReader(strings.NewReader(policy))
			reader.FieldsPerRecord = -1
			rows, err := reader.ReadAll()
			g.Expect(err).NotTo(HaveOccurred())
			var actual []string
			for _, row := range rows {
				if strings.TrimSpace(row[0]) != "g" {
					continue
				}
				for i := range row {
					row[i] = strings.TrimSpace(row[i])
				}
				actual = append(actual, strings.Join(row, ", "))
			}
			// Comparing the complete multiset also rejects duplicate and ServiceAccount grants.
			g.Expect(actual).To(ConsistOf(expected))
		}, timeout, defaultPollInterval).Should(Succeed())
	}

	BeforeEach(func() {
		name := "rbac-permissions-" + rand.String(8)
		selector = e2eLabels("e2e_rbac_permissions")
		selector["app.kubernetes.io/instance"] = name
		settings := &addonsv1alpha1.ArgoAddon{}
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: e2eConfigName()}, settings)).To(Succeed())
		configKey = client.ObjectKey{Name: settings.Spec.Argo.RBACConfigMap, Namespace: settings.Spec.Argo.Namespace}

		create(&addonsv1alpha1.ArgoTranslator{
			ObjectMeta: metav1.ObjectMeta{Name: name, Labels: selector},
			Spec: addonsv1alpha1.ArgoTranslatorSpec{
				Selector: &metav1.LabelSelector{MatchLabels: selector},
				ProjectRoles: []addonsv1alpha1.ArgocdProjectRolesTranslator{
					{
						Name: "editor", ClusterRoles: []string{"edit"}, Owner: true,
						Policies: []addonsv1alpha1.ArgocdPolicyDefinition{{
							Resource: "applications", Action: []string{"get", "sync"}, Verb: "allow", Path: "*",
						}},
					},
					{
						Name: "viewer", ClusterRoles: []string{"view"},
						Policies: []addonsv1alpha1.ArgocdPolicyDefinition{{
							Resource: "applications", Action: []string{"get"}, Verb: "allow", Path: "*",
						}},
					},
				},
			},
		})
		tenant = &capsulev1beta2.Tenant{
			ObjectMeta: metav1.ObjectMeta{
				Name: name, Labels: selector,
				Annotations: map[string]string{"argo.addons.projectcapsule.dev/decouple": "false"},
			},
			Spec: capsulev1beta2.TenantSpec{
				Permissions: capsulev1beta2.Permissions{
					MatchOwners: []*metav1.LabelSelector{{MatchLabels: selector}},
				},
			},
		}
		create(tenant)
		expectBindings()
	})

	DescribeTable("reflects status-only owners and revokes their previous roles",
		func(kind string) {
			subject := rbacv1.Subject{Kind: kind, Name: tenant.Name + "-resolved"}
			By("letting Capsule resolve an external TenantOwner into status.owners")
			owner := createOwner(subject)
			expectStatusOwner(subject, "edit")
			expectBindings(bindings("editor", subject.Name)...)

			By("updating only the TenantOwner and observing the new status-based grants")
			Eventually(func() error {
				current := &capsulev1beta2.TenantOwner{}
				if err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(owner), current); err != nil {
					return err
				}
				current.Spec.ClusterRoles = []string{"view"}
				return k8sClient.Update(context.Background(), current)
			}, timeout, defaultPollInterval).Should(Succeed())
			expectStatusOwner(subject, "view")
			expectBindings(bindings("viewer", subject.Name)...)

			By("removing the TenantOwner and revoking all of its Argo bindings")
			Expect(k8sClient.Delete(context.Background(), owner)).To(Succeed())
			expectStatusOwner(subject)
			expectBindings()
		},
		Entry("User", rbacv1.UserKind),
		Entry("Group", rbacv1.GroupKind),
	)

	It("reflects users and groups from every rule and revokes changed or removed bindings", func() {
		editUser, editGroup := tenant.Name+"-edit-user", tenant.Name+"-edit-group"
		viewUser, viewGroup := tenant.Name+"-view-user", tenant.Name+"-view-group"
		By("adding bindings in two rules with different namespace selectors")
		updateTenant(func(current *capsulev1beta2.Tenant) {
			current.Spec.Rules = []*capsulerules.NamespaceRuleBodyTenant{
				{
					NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"environment": "dev"}},
					Permissions: capsulerules.NamespaceRulePermissionBody{
						Bindings: []capsulerbac.AdditionalRoleBindingsSpec{{ClusterRoleName: "edit", Subjects: []rbacv1.Subject{
							{Kind: rbacv1.UserKind, Name: editUser}, {Kind: rbacv1.GroupKind, Name: editGroup},
							{Kind: rbacv1.ServiceAccountKind, Name: "robot", Namespace: "default"},
						}}},
					},
				},
				{
					NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"environment": "prod"}},
					Permissions: capsulerules.NamespaceRulePermissionBody{
						Bindings: []capsulerbac.AdditionalRoleBindingsSpec{{ClusterRoleName: "view", Subjects: []rbacv1.Subject{
							{Kind: rbacv1.UserKind, Name: viewUser}, {Kind: rbacv1.GroupKind, Name: viewGroup},
						}}},
					},
				},
			}
		})
		expectBindings(append(bindings("editor", editUser, editGroup), bindings("viewer", viewUser, viewGroup)...)...)

		By("removing the first rule while preserving grants from the second rule")
		updateTenant(func(current *capsulev1beta2.Tenant) { current.Spec.Rules = current.Spec.Rules[1:] })
		expectBindings(bindings("viewer", viewUser, viewGroup)...)

		By("changing the remaining binding's ClusterRole")
		updateTenant(func(current *capsulev1beta2.Tenant) {
			current.Spec.Rules[0].Permissions.Bindings[0].ClusterRoleName = "edit"
		})
		expectBindings(bindings("editor", viewUser, viewGroup)...)

		By("removing all rule bindings")
		updateTenant(func(current *capsulev1beta2.Tenant) { current.Spec.Rules = nil })
		expectBindings()
	})

	It("merges status owners, legacy bindings, and rule bindings without duplicates", func() {
		shared := rbacv1.Subject{Kind: rbacv1.UserKind, Name: tenant.Name + "-shared"}
		legacy := rbacv1.Subject{Kind: rbacv1.GroupKind, Name: tenant.Name + "-legacy"}
		ruleOnly := rbacv1.Subject{Kind: rbacv1.UserKind, Name: tenant.Name + "-rule"}
		owner := createOwner(shared)
		expectStatusOwner(shared, "edit")

		By("granting the same role through all three sources and multiple rules")
		updateTenant(func(current *capsulev1beta2.Tenant) {
			current.Spec.AdditionalRoleBindings = []capsulerbac.AdditionalRoleBindingsSpec{{
				ClusterRoleName: "edit", Subjects: []rbacv1.Subject{shared, legacy},
			}}
			current.Spec.Rules = []*capsulerules.NamespaceRuleBodyTenant{
				{Permissions: capsulerules.NamespaceRulePermissionBody{Bindings: []capsulerbac.AdditionalRoleBindingsSpec{{
					ClusterRoleName: "edit", Subjects: []rbacv1.Subject{shared, ruleOnly},
				}}}},
				{Permissions: capsulerules.NamespaceRulePermissionBody{Bindings: []capsulerbac.AdditionalRoleBindingsSpec{{
					ClusterRoleName: "edit", Subjects: []rbacv1.Subject{shared},
				}}}},
			}
		})
		expectBindings(bindings("editor", shared.Name, legacy.Name, ruleOnly.Name)...)

		By("removing rule bindings while keeping status and legacy grants")
		updateTenant(func(current *capsulev1beta2.Tenant) { current.Spec.Rules = nil })
		expectBindings(bindings("editor", shared.Name, legacy.Name)...)

		By("removing the status owner while keeping its legacy grant")
		Expect(k8sClient.Delete(context.Background(), owner)).To(Succeed())
		expectStatusOwner(shared)
		expectBindings(bindings("editor", shared.Name, legacy.Name)...)

		By("revoking the final legacy grants")
		updateTenant(func(current *capsulev1beta2.Tenant) { current.Spec.AdditionalRoleBindings = nil })
		expectBindings()
	})
})
