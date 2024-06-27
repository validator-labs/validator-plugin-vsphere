package privileges

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/object"
	corev1 "k8s.io/api/core/v1"

	"github.com/validator-labs/validator-plugin-vsphere/api/v1alpha1"
	"github.com/validator-labs/validator-plugin-vsphere/internal/vcsim"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/vsphere"
	vapi "github.com/validator-labs/validator/api/v1alpha1"
	"github.com/validator-labs/validator/pkg/types"
	"github.com/validator-labs/validator/pkg/util"
)

func TestRolePrivilegeValidationService_ReconcileRolePrivilegesRule(t *testing.T) {
	var log logr.Logger

	userPrivilegesMap := make(map[string]bool)
	userName := "admin@vsphere.local"
	vcSim := vcsim.NewVCSim(userName, log)

	vcSim.Start()
	defer vcSim.Shutdown()

	userPrivilegesMap["Cns.Searchable"] = true

	// monkey-patch GetUserGroupAndPrincipals and IsAdminAccount
	GetUserAndGroupPrincipals = func(ctx context.Context, username string, driver *vsphere.VSphereCloudDriver, log logr.Logger) (string, []string, error) {
		return "admin", []string{"Administrators"}, nil
	}

	vsphere.IsAdminAccount = func(ctx context.Context, driver *vsphere.VSphereCloudDriver) (bool, error) {
		return true, nil
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
				Message:        fmt.Sprintf("One or more required privileges was not found, or a condition was not met for account: %s", userName),
				Details:        []string{},
				Failures:       []string{"Privilege: Cns.Searchable, was not found in the user's privileges"},
				Status:         corev1.ConditionFalse,
			},
				State: util.Ptr(vapi.ValidationFailed),
			},
			expectedErr: ErrRequiredRolePrivilegesNotFound,
		},
	}

	for _, tc := range testCases {
		vr, err := validationService.ReconcileRolePrivilegesRule(tc.rule, vcSim.Driver, authManager)
		util.CheckTestCase(t, vr, tc.expectedResult, err, tc.expectedErr)
	}
}

func TestIsSameUser(t *testing.T) {
	testCases := []struct {
		name          string
		userPrincipal string
		username      string
		expected      bool
	}{
		{
			name:          `Valid match`,
			userPrincipal: `VSPHERE.LOCAL\username`,
			username:      `username@vsphere.local`,
			expected:      true,
		},
		{
			name:          `Different usernames`,
			userPrincipal: `VSPHERE.LOCAL\username`,
			username:      `differentUsername@vsphere.local`,
			expected:      false,
		},
		{
			name:          `Different domain`,
			userPrincipal: `VSPHERE.LOCAL\username`,
			username:      `username@vsphere.notlocal`,
			expected:      false,
		},
		{
			name:          `Invalid input - missing domain`,
			userPrincipal: `VSPHERE.LOCAL\username`,
			username:      `username`,
			expected:      false,
		},
		{
			name:          `Invalid input - missing username`,
			userPrincipal: `VSPHERE.LOCAL\`,
			username:      `username@vsphere.local`,
			expected:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isSameUser(tc.userPrincipal, tc.username)
			if result != tc.expected {
				t.Errorf("Expected %v but got %v", tc.expected, result)
			}
		})
	}
}
