package controller

import (
	"context"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/validator-labs/validator-plugin-vsphere/api/v1alpha1"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/vcsim"
	vapi "github.com/validator-labs/validator/api/v1alpha1"
	vres "github.com/validator-labs/validator/pkg/validationresult"
	//+kubebuilder:scaffold:imports
)

const (
	vsphereValidatorName = "vsphere-validator"
	username             = "admin@vsphere.local"
)

var _ = Describe("VsphereValidator controller", Ordered, func() {

	BeforeEach(func() {
		// toggle true/false to enable/disable the VsphereValidator controller specs
		if false {
			Skip("skipping")
		}
	})

	val := &v1alpha1.VsphereValidator{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vsphereValidatorName,
			Namespace: validatorNamespace,
		},
		Spec: v1alpha1.VsphereValidatorSpec{
			Auth: v1alpha1.VsphereAuth{
				SecretName: "validator-secret",
			},
			Datacenter:               "DC0",
			PrivilegeValidationRules: []v1alpha1.PrivilegeValidationRule{},
			ComputeResourceRules:     []v1alpha1.ComputeResourceRule{},
			NTPValidationRules:       []v1alpha1.NTPValidationRule{},
			TagValidationRules: []v1alpha1.TagValidationRule{
				{
					RuleName:   "Datacenter k8s-region rule",
					EntityType: "datacenter",
					EntityName: "Datacenter",
					Tag:        "k8s-region",
				},
			},
		},
	}

	vr := &vapi.ValidationResult{}
	vrKey := types.NamespacedName{Name: vres.Name(val), Namespace: validatorNamespace}

	vcSim := vcsim.NewVCSim(username, 8446, logr.Logger{})
	vcSim.Start()
	account := vcSim.Account

	validatorSecret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "validator-secret",
			Namespace: validatorNamespace,
		},
		Immutable: nil,
		Data: map[string][]byte{
			"username":           []byte(account.Username),
			"password":           []byte(account.Password),
			"insecureSkipVerify": []byte("true"),
			"vcenterServer":      []byte(account.Host),
		},
	}

	It("Should create a ValidationResult and update its Status with a failed condition", func() {
		By("By creating a new VsphereValidator")
		ctx := context.Background()

		Expect(k8sClient.Create(ctx, validatorSecret)).Should(Succeed())
		Expect(k8sClient.Create(ctx, val)).Should(Succeed())

		// Wait for the ValidationResult's Status to be updated
		Eventually(func() bool {
			if err := k8sClient.Get(ctx, vrKey, vr); err != nil {
				return false
			}
			stateOk := vr.Status.State == vapi.ValidationFailed
			return stateOk
		}, timeout, interval).Should(BeTrue(), "failed to create a ValidationResult")

		vcSim.Shutdown()
	})
})
