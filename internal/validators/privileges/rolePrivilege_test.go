package privileges

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/api/v1alpha1"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/internal/vcsim"
	v8or "github.com/spectrocloud-labs/valid8or/api/v1alpha1"
	"github.com/spectrocloud-labs/valid8or/pkg/types"
	"github.com/spectrocloud-labs/valid8or/pkg/util/ptr"
	"github.com/vmware/govmomi/object"
	corev1 "k8s.io/api/core/v1"
	"reflect"
	"testing"
)

func TestRolePrivilegeValidationService_ReconcileRolePrivilegesRule(t *testing.T) {
	var log logr.Logger
	userPrivilegesMap := make(map[string]bool)

	userName := "admin@vsphere.local"
	vcSim := vcsim.NewVCSim(userName)

	vcSim.Start()
	defer vcSim.Shutdown()

	userPrivilegesMap["Cns.Searchable"] = true

	//finder := find.NewFinder(vcSim.Driver.Client.Client)
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
		expectedResult types.ValidationResult
	}{
		{
			name: "All privileges available",
			rule: v1alpha1.GenericRolePrivilegeValidationRule{
				Name: "Cns.Searchable",
			},
			expectedResult: types.ValidationResult{Condition: &v8or.ValidationCondition{
				ValidationType: "vsphere-role-privileges",
				ValidationRule: "validation-Cns.Searchable",
				Message:        "All required vsphere-role-privileges permissions were found",
				Details:        []string{},
				Failures:       nil,
				Status:         corev1.ConditionTrue,
			},
				State: ptr.Ptr(v8or.ValidationSucceeded),
			},
		},
		{
			name: "InventoryService.Tagging.CreateTag not available",
			rule: v1alpha1.GenericRolePrivilegeValidationRule{
				Name: "InventoryService.Tagging.CreateTag",
			},
			expectedResult: types.ValidationResult{Condition: &v8or.ValidationCondition{
				ValidationType: "vsphere-role-privileges",
				ValidationRule: "validation-InventoryService.Tagging.CreateTag",
				Message:        "One or more required privileges was not found, or a condition was not met",
				Details:        []string{},
				Failures:       []string{"Rule: InventoryService.Tagging.CreateTag, was not found in the user's privileges"},
				Status:         corev1.ConditionFalse,
			},
				State: ptr.Ptr(v8or.ValidationFailed),
			},
		},
	}

	for _, tc := range testCases {
		vr, err := validationService.ReconcileRolePrivilegesRule(tc.rule, userPrivilegesMap, nil, nil, nil)
		CheckTestCase(t, vr, tc.expectedResult, err, tc.expectedErr)
	}

}

func CheckTestCase(t *testing.T, res *types.ValidationResult, expectedResult types.ValidationResult, err, expectedError error) {
	if !reflect.DeepEqual(res.State, expectedResult.State) {
		t.Errorf("expected state (%+v), got (%+v)", expectedResult.State, res.State)
	}
	if !reflect.DeepEqual(res.Condition.ValidationType, expectedResult.Condition.ValidationType) {
		t.Errorf("expected validation type (%s), got (%s)", expectedResult.Condition.ValidationType, res.Condition.ValidationType)
	}
	if !reflect.DeepEqual(res.Condition.ValidationRule, expectedResult.Condition.ValidationRule) {
		t.Errorf("expected validation rule (%s), got (%s)", expectedResult.Condition.ValidationRule, res.Condition.ValidationRule)
	}
	if !reflect.DeepEqual(res.Condition.Message, expectedResult.Condition.Message) {
		t.Errorf("expected message (%s), got (%s)", expectedResult.Condition.Message, res.Condition.Message)
	}
	if !reflect.DeepEqual(res.Condition.Details, expectedResult.Condition.Details) {
		t.Errorf("expected details (%s), got (%s)", expectedResult.Condition.Details, res.Condition.Details)
	}
	if !reflect.DeepEqual(res.Condition.Failures, expectedResult.Condition.Failures) {
		t.Errorf("expected failures (%s), got (%s)", expectedResult.Condition.Failures, res.Condition.Failures)
	}
	if !reflect.DeepEqual(res.Condition.Status, expectedResult.Condition.Status) {
		t.Errorf("expected status (%s), got (%s)", expectedResult.Condition.Status, res.Condition.Status)
	}
}
