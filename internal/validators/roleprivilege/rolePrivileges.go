package roleprivilege

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/api/v1alpha1"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/internal/constants"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/internal/vsphere"
	v8or "github.com/spectrocloud-labs/valid8or/api/v1alpha1"
	v8orconstants "github.com/spectrocloud-labs/valid8or/pkg/constants"
	"github.com/spectrocloud-labs/valid8or/pkg/types"
	"github.com/spectrocloud-labs/valid8or/pkg/util/ptr"
	"github.com/vmware/govmomi/object"
	corev1 "k8s.io/api/core/v1"
)

func buildValidationResult(rule v1alpha1.RolePrivilegeValidationRule, validationType string) *types.ValidationResult {
	state := v8or.ValidationSucceeded
	latestCondition := v8or.DefaultValidationCondition()
	latestCondition.Message = fmt.Sprintf("All required %s permissions were found", validationType)
	latestCondition.ValidationRule = fmt.Sprintf("%s-%s", v8orconstants.ValidationRulePrefix, rule.Name)
	latestCondition.ValidationType = validationType
	return &types.ValidationResult{Condition: &latestCondition, State: &state}
}

func GetUserRolePrivilegesMapping(driver *vsphere.VSphereCloudDriver) (map[string]bool, error) {
	privileges, err := validateRolePrivileges(driver)
	if err != nil {
		fmt.Println(err, "Error validating Role privileges")
		return nil, err
	}
	return privileges, nil
}

func ReconcileRolePrivilegesRule(rule v1alpha1.RolePrivilegeValidationRule, privileges map[string]bool) (*types.ValidationResult, error) {

	vr := buildValidationResult(rule, constants.ValidationTypeRolePrivileges)

	valid := vsphere.IsValidRule(rule, privileges)
	if !valid {
		vr.State = ptr.Ptr(v8or.ValidationFailed)
		vr.Condition.Failures = append(vr.Condition.Failures, "Rule: %s, was not found in the user's privileges")
		vr.Condition.Message = "One or more required privileges was not found, or a condition was not met"
		vr.Condition.Status = corev1.ConditionFalse
		return vr, errors.New("Rule not valid")
	}

	return vr, nil
}

func validateRolePrivileges(vsphereCloudDriver *vsphere.VSphereCloudDriver) (map[string]bool, error) {

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	// Get the authorization manager from the Client
	authManager := object.NewAuthorizationManager(vsphereCloudDriver.Client.Client)
	if authManager == nil {
		return nil, fmt.Errorf("Error getting authorization manager")
	}

	// Get the current user
	userName, err := vsphereCloudDriver.GetCurrentVmwareUser(ctx)
	if err != nil {
		return nil, err
	}

	// Get list of roles for current user
	userPrivileges, err := vsphereCloudDriver.GetVmwareUserPrivileges(userName, authManager)
	if err != nil {
		return nil, err
	}

	return userPrivileges, nil
}
