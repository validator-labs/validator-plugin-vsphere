package integration

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/validator-labs/validator-plugin-vsphere/tests/integration/common"
	"github.com/validator-labs/validator-plugin-vsphere/tests/integration/tags"
	"github.com/validator-labs/validator-plugin-vsphere/tests/utils/test"
)

var _ = ginkgo.Describe("Palette CLI Integration Test Suite", func() {

	ginkgo.Context("Executing Palette CLI integration tests", func() {
		ginkgo.It("should not error", func() {
			testCtx := test.NewTestContext()
			err := test.Flow(testCtx).
				Test(common.NewSingleFuncTest("tags-test", tags.Execute)).
				TearDown().
				Audit()
			gomega.Expect(err).To(gomega.Not(gomega.HaveOccurred()))
		})
	})
})
