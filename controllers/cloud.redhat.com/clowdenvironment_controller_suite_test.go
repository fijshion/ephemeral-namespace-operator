package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
)

var _ = Describe("Clowdenvironment controller basic update", func() {
	const (
		timeout  = time.Second * 90
		duration = time.Second * 90
		interval = time.Millisecond * 250
	)

	Context("When a clowdenvironment is created", func() {
		It("Should update the namespace annotations when ready if owned by the pool", func() {
			By("Checking the clowdenvironment conditions")
			ctx := context.Background()
			nsList := core.NamespaceList{}

			Eventually(func() bool {
				err := k8sClient.List(ctx, &nsList)
				Expect(err).NotTo(HaveOccurred())

				if len(nsList.Items) == 0 {
					return false
				}

				for _, ns := range nsList.Items {
					if isOwnedByPool(ctx, k8sClient, ns.Name) {
						a := ns.GetAnnotations()
						if val, ok := a["status"]; !ok || val != "ready" {
							if val != "deleting" {
								return false
							}
						}
					}
				}

				return true
			}, timeout, interval).Should(BeTrue())
		})

		It("Should ignore envs not owned by the pool", func() {
			By("Checking namespace ownerRef in event filter")
			ctx := context.Background()

			ns := core.Namespace{}
			ns.Name = "no-owner"

			err := k8sClient.Create(ctx, &ns)
			Expect(err).NotTo(HaveOccurred())

			Expect(isOwnedByPool(ctx, k8sClient, ns.Name)).To(Equal(false))
		})
	})
})
