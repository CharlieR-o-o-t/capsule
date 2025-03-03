//go:build e2e

// Copyright 2020-2023 Project Capsule Authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"github.com/projectcapsule/capsule/pkg/api"
	"github.com/projectcapsule/capsule/pkg/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("checking namespace metadata considers tenant spec", func() {
	tnt := &capsulev1beta2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
		Spec: capsulev1beta2.TenantSpec{
			Owners: capsulev1beta2.OwnerListSpec{
				{
					Name: "ruby",
					Kind: "User",
				},
			},
			NodeSelector: map[string]string{
				"foo": "bar",
			},
			NamespaceOptions: &capsulev1beta2.NamespaceOptions{
				AdditionalMetadata: &api.AdditionalMetadataSpec{
					Labels: map[string]string{
						"label1key": "label1value",
					},
					Annotations: map[string]string{
						"annotation1key": "annotation1value",
					},
				},
			},
		},
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo-namespace",
		},
	}

	JustBeforeEach(func() {
		EventuallyCreation(func() error {
			tnt.ResourceVersion = ""
			return k8sClient.Create(context.TODO(), tnt)
		}).Should(Succeed())
	})
	JustAfterEach(func() {
		Expect(k8sClient.Delete(context.TODO(), tnt)).Should(Succeed())
		Expect(k8sClient.Delete(context.TODO(), ns)).Should(Succeed())
	})

	It("tenant namespace metadata check", func() {
		By("checking namespace annotations", func() {
			EventuallyCreation(func() error {
				return k8sClient.Create(context.TODO(), ns)
			}).Should(Succeed())

			nsExpected := &corev1.Namespace{}
			Eventually(func() error {
				return k8sClient.Get(context.TODO(), types.NamespacedName{Name: ns.GetName()}, nsExpected)
			}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
			Expect(nsExpected.GetAnnotations()).Should(Equal(tnt.Spec.NamespaceOptions.AdditionalMetadata.Annotations))
		})
		By("checking namespace labels", func() {
			EventuallyCreation(func() error {
				return k8sClient.Create(context.TODO(), ns)
			}).Should(Succeed())

			nsExpected := &corev1.Namespace{}
			Eventually(func() error {
				return k8sClient.Get(context.TODO(), types.NamespacedName{Name: ns.GetName()}, nsExpected)
			}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
			Expect(nsExpected.GetLabels()).Should(Equal(tnt.Spec.NamespaceOptions.AdditionalMetadata.Labels))
		})
		By("checking namespace node-selector annotation", func() {
			EventuallyCreation(func() error {
				return k8sClient.Create(context.TODO(), ns)
			}).Should(Succeed())

			nsExpected := &corev1.Namespace{}
			nsAnnotations := utils.BuildNodeSelector(tnt, ns.GetAnnotations())

			Eventually(func() error {
				return k8sClient.Get(context.TODO(), types.NamespacedName{Name: ns.GetName()}, nsExpected)
			}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
			Expect(nsExpected.Annotations[utils.NodeSelectorAnnotation]).Should(Equal(nsAnnotations[utils.NodeSelectorAnnotation]))
		})
		By("checking tenant metadata annotation override", func() {
			ns.Annotations = map[string]string{
				"annotation1key": "annotation1value-override",
			}

			EventuallyCreation(func() error {
				return k8sClient.Create(context.TODO(), ns)
			}).Should(Succeed())

			nsExpected := &corev1.Namespace{}

			Eventually(func() error {
				return k8sClient.Get(context.TODO(), types.NamespacedName{Name: ns.GetName()}, nsExpected)
			}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
			Expect(nsExpected.Annotations["annotation1key"]).Should(Equal(tnt.Spec.NamespaceOptions.AdditionalMetadata.Annotations["annotation1key"]))
		})
	})
})
