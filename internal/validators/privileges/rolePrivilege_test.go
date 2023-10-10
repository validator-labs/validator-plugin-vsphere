package privileges

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/object"
	corev1 "k8s.io/api/core/v1"

	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/api/v1alpha1"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/internal/vcsim"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/internal/vsphere"
	v8or "github.com/spectrocloud-labs/valid8or/api/v1alpha1"
	"github.com/spectrocloud-labs/valid8or/pkg/types"
	"github.com/spectrocloud-labs/valid8or/pkg/util/ptr"
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
		expectedResult types.ValidationResult
	}{
		{
			name: "All privileges available",
			rule: v1alpha1.GenericRolePrivilegeValidationRule{
				Username:   userName,
				Privileges: []string{"Datastore.AllocateSpace"},
			},
			expectedResult: types.ValidationResult{Condition: &v8or.ValidationCondition{
				ValidationType: "vsphere-role-privileges",
				ValidationRule: fmt.Sprintf("validation-%s", userName),
				Message:        "All required vsphere-role-privileges permissions were found",
				Details:        []string{},
				Failures:       nil,
				Status:         corev1.ConditionTrue,
			},
				State: ptr.Ptr(v8or.ValidationSucceeded),
			},
		},
		{
			name: "Cns.Searchable not available",
			rule: v1alpha1.GenericRolePrivilegeValidationRule{
				Username:   userName,
				Privileges: []string{"Cns.Searchable"},
			},
			expectedResult: types.ValidationResult{Condition: &v8or.ValidationCondition{
				ValidationType: "vsphere-role-privileges",
				ValidationRule: fmt.Sprintf("validation-%s", userName),
				Message:        "One or more required privileges was not found, or a condition was not met",
				Details:        []string{},
				Failures:       []string{"Privilege: Cns.Searchable, was not found in the user's privileges"},
				Status:         corev1.ConditionFalse,
			},
				State: ptr.Ptr(v8or.ValidationFailed),
			},
		},
	}

	for _, tc := range testCases {
		vr, err := validationService.ReconcileRolePrivilegesRule(tc.rule, vcSim.Driver, authManager)
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
