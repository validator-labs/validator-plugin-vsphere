package privileges

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	corev1 "k8s.io/api/core/v1"

	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/api/v1alpha1"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/internal/vcsim"
	v8or "github.com/spectrocloud-labs/valid8or/api/v1alpha1"
	"github.com/spectrocloud-labs/valid8or/pkg/types"
	"github.com/spectrocloud-labs/valid8or/pkg/util/ptr"
)

func TestRolePrivilegeValidationService_ReconcileEntityPrivilegeRule(t *testing.T) {
	var log logr.Logger
	userPrivilegesMap := make(map[string]bool)

	userName := "admin2@vsphere.local"
	vcSim := vcsim.NewVCSim(userName)

	vcSim.Start()
	defer vcSim.Shutdown()

	userPrivilegesMap["Cns.Searchable"] = true

	finder := find.NewFinder(vcSim.Driver.Client.Client)
	authManager := object.NewAuthorizationManager(vcSim.Driver.Client.Client)
	if authManager == nil {
		t.Fatal("Error in creating auth manager")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	userName, err := vcSim.Driver.GetCurrentVmwareUser(ctx)
	if err != nil {
		t.Fatal("Error in getting current VMware user from username")
	}

	validationService := NewPrivilegeValidationService(log, vcSim.Driver, "DC0", authManager, userName)
	testCases := []struct {
		name           string
		expectedErr    error
		rule           v1alpha1.EntityPrivilegeValidationRule
		expectedResult types.ValidationResult
	}{
		{
			name: "All privileges available",
			rule: v1alpha1.EntityPrivilegeValidationRule{
				Name:        "VirtualMachine.Config.AddExistingDisk",
				Username:    userName,
				ClusterName: "DC0_C0",
				EntityType:  "cluster",
				EntityName:  "DC0_C0",
				Privileges: []string{
					"VirtualMachine.Config.AddExistingDisk",
				},
			},
			expectedResult: types.ValidationResult{Condition: &v8or.ValidationCondition{
				ValidationType: "vsphere-entity-privileges",
				ValidationRule: "validation-cluster-DC0_C0",
				Message:        fmt.Sprintf("All required vsphere-entity-privileges permissions were found for account: %s", userName),
				Details:        []string{},
				Failures:       nil,
				Status:         corev1.ConditionTrue,
			},
				State: ptr.Ptr(v8or.ValidationSucceeded),
			},
		},
		{
			name: "Certain privilege not available",
			rule: v1alpha1.EntityPrivilegeValidationRule{
				Name:        "VirtualMachine.Config.AddExistingDisk",
				Username:    userName,
				ClusterName: "DC0_C0",
				EntityType:  "cluster",
				EntityName:  "DC0_C0",
				Privileges: []string{
					"VirtualMachine.Config.DestroyExistingDisk",
				},
			},
			expectedResult: types.ValidationResult{Condition: &v8or.ValidationCondition{
				ValidationType: "vsphere-entity-privileges",
				ValidationRule: "validation-cluster-DC0_C0",
				Message:        "One or more required privileges was not found, or a condition was not met",
				Details:        []string{},
				Failures:       []string{"user: admin2@vsphere.local does not have privilege: VirtualMachine.Config.DestroyExistingDisk on entity type: cluster with name: DC0_C0"},
				Status:         corev1.ConditionFalse,
			},
				State: ptr.Ptr(v8or.ValidationFailed),
			},
		},
	}

	for _, tc := range testCases {
		vr, err := validationService.ReconcileEntityPrivilegeRule(tc.rule, finder)
		CheckTestCase(t, vr, tc.expectedResult, err, tc.expectedErr)
	}
}
