package privileges

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/object"
	corev1 "k8s.io/api/core/v1"

	"github.com/spectrocloud-labs/validator-plugin-vsphere/api/v1alpha1"
	"github.com/spectrocloud-labs/validator-plugin-vsphere/internal/vcsim"
	"github.com/spectrocloud-labs/validator-plugin-vsphere/pkg/vsphere"
	vapi "github.com/spectrocloud-labs/validator/api/v1alpha1"
	"github.com/spectrocloud-labs/validator/pkg/types"
	"github.com/spectrocloud-labs/validator/pkg/util"
)

func TestRolePrivilegeValidationService_ReconcileRolePrivilegesRule(t *testing.T) {
	var log logr.Logger
	userPrivilegesMap := make(map[string]bool)
	userName := "admin@vsphere.local"
	vcSim := vcsim.NewVCSim(userName)

	vcSim.Start()
	defer vcSim.Shutdown()

	userPrivilegesMap["Cns.Searchable"] = true

	// monkey-patch get user group and principals
	GetUserAndGroupPrincipals = func(ctx context.Context, username string, driver *vsphere.VSphereCloudDriver) (string, []string, error) {
		return "admin", []string{"Administrators"}, nil
	}

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
		rule           v1alpha1.GenericRolePrivilegeValidationRule
		expectedResult types.ValidationRuleResult
	}{
		{
			name: "All privileges available",
			rule: v1alpha1.GenericRolePrivilegeValidationRule{
				Username:   userName,
				Privileges: []string{"Datastore.AllocateSpace"},
			},
			expectedResult: types.ValidationRuleResult{Condition: &vapi.ValidationCondition{
				ValidationType: "vsphere-role-privileges",
				ValidationRule: fmt.Sprintf("validation-%s", userName),
				Message:        "All required vsphere-role-privileges permissions were found",
				Details:        []string{},
				Failures:       nil,
				Status:         corev1.ConditionTrue,
			},
				State: util.Ptr(vapi.ValidationSucceeded),
			},
		},
		{
			name: "Cns.Searchable not available",
			rule: v1alpha1.GenericRolePrivilegeValidationRule{
				Username:   userName,
				Privileges: []string{"Cns.Searchable"},
			},
			expectedResult: types.ValidationRuleResult{Condition: &vapi.ValidationCondition{
				ValidationType: "vsphere-role-privileges",
				ValidationRule: fmt.Sprintf("validation-%s", userName),
				Message:        "One or more required privileges was not found, or a condition was not met",
				Details:        []string{},
				Failures:       []string{"Privilege: Cns.Searchable, was not found in the user's privileges"},
				Status:         corev1.ConditionFalse,
			},
				State: util.Ptr(vapi.ValidationFailed),
			},
		},
	}

	for _, tc := range testCases {
		vr, err := validationService.ReconcileRolePrivilegesRule(tc.rule, vcSim.Driver, authManager)
		util.CheckTestCase(t, vr, tc.expectedResult, err, tc.expectedErr)
	}
}
