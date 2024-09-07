package privileges

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	corev1 "k8s.io/api/core/v1"

	vapi "github.com/validator-labs/validator/api/v1alpha1"
	"github.com/validator-labs/validator/pkg/test"
	"github.com/validator-labs/validator/pkg/types"
	"github.com/validator-labs/validator/pkg/util"

	"github.com/validator-labs/validator-plugin-vsphere/api/v1alpha1"
	"github.com/validator-labs/validator-plugin-vsphere/api/vcenter/entity"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/vcsim"
)

func TestPrivilegeValidationService_ReconcilePrivilegeRule(t *testing.T) {
	var log logr.Logger

	vcSim := vcsim.NewVCSim("admin2@vsphere.local", 8448, log)
	vcSim.Start()
	defer vcSim.Shutdown()

	finder := find.NewFinder(vcSim.Driver.Client.Client)
	authManager := object.NewAuthorizationManager(vcSim.Driver.Client.Client)
	if authManager == nil {
		t.Fatal("Error in creating auth manager")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	username, err := vcSim.Driver.CurrentUser(ctx)
	if err != nil {
		t.Fatal("Error in getting current VMware user from username")
	}

	validationService := NewPrivilegeValidationService(log, vcSim.Driver, "DC0", username, authManager)

	testCases := []struct {
		name           string
		expectedErr    error
		rule           v1alpha1.PrivilegeValidationRule
		expectedResult types.ValidationRuleResult
	}{
		{
			name: "All privileges available",
			rule: v1alpha1.PrivilegeValidationRule{
				RuleName:    "VirtualMachine.Config.AddExistingDisk",
				ClusterName: "DC0_C0",
				EntityType:  entity.Cluster,
				EntityName:  "DC0_C0",
				Privileges: []string{
					"VirtualMachine.Config.AddExistingDisk",
				},
			},
			expectedResult: types.ValidationRuleResult{Condition: &vapi.ValidationCondition{
				ValidationType: "vsphere-privileges",
				ValidationRule: "validation-cluster-DC0_C0",
				Message:        fmt.Sprintf("All required vsphere-privileges permissions were found for account: %s", username),
				Details:        []string{},
				Failures:       []string{},
				Status:         corev1.ConditionTrue,
			},
				State: util.Ptr(vapi.ValidationSucceeded),
			},
		},
		{
			name: "Certain privilege not available",
			rule: v1alpha1.PrivilegeValidationRule{
				RuleName:    "VirtualMachine.Config.AddExistingDisk",
				ClusterName: "DC0_C0",
				EntityType:  entity.Cluster,
				EntityName:  "DC0_C0",
				Privileges: []string{
					"VirtualMachine.Config.DestroyExistingDisk",
				},
			},
			expectedResult: types.ValidationRuleResult{Condition: &vapi.ValidationCondition{
				ValidationType: "vsphere-privileges",
				ValidationRule: "validation-cluster-DC0_C0",
				Message:        fmt.Sprintf("One or more required privileges was not found, or a condition was not met for account: %s", username),
				Details:        []string{},
				Failures:       []string{"user: admin2@vsphere.local does not have privilege: VirtualMachine.Config.DestroyExistingDisk on entity type: cluster with name: DC0_C0"},
				Status:         corev1.ConditionFalse,
			},
				State: util.Ptr(vapi.ValidationFailed),
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		vr, err := validationService.ReconcilePrivilegeRule(tc.rule, finder)
		test.CheckTestCase(t, vr, tc.expectedResult, err, tc.expectedErr)
	}
}
