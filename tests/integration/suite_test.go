package integration

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

func Test(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Palette CLI Suite")
}

var _ = ginkgo.BeforeSuite(func() {
	ginkgo.By("preparing the test environment")

	done := make(chan interface{})
	go func() {
		// user test code to run asynchronously
		close(done) // signifies the code is done
	}()
	gomega.Eventually(done, 60).Should(gomega.BeClosed())
})

var _ = ginkgo.AfterSuite(func() {
	ginkgo.By("tearing down the test environment")
})
