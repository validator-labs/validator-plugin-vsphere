package privileges

import (
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

func buildValidationResult(rule v1alpha1.GenericRolePrivilegeValidationRule, validationType string) *types.ValidationResult {
	state := v8or.ValidationSucceeded
	latestCondition := v8or.DefaultValidationCondition()
	latestCondition.Message = fmt.Sprintf("All required %s permissions were found", validationType)
	latestCondition.ValidationRule = fmt.Sprintf("%s-%s", v8orconstants.ValidationRulePrefix, rule.Name)
	latestCondition.ValidationType = validationType

	return &types.ValidationResult{Condition: &latestCondition, State: &state}
}

func (s *PrivilegeValidationService) GetUserRolePrivilegesMapping() (map[string]bool, error) {
	privileges, err := getUserPrivileges(s.driver, s.authManager, s.datacenter, s.userName)
	if err != nil {
		return nil, err
	}
	return privileges, nil
}

func (s *PrivilegeValidationService) ReconcileRolePrivilegesRule(rule v1alpha1.GenericRolePrivilegeValidationRule, privileges map[string]bool) (*types.ValidationResult, error) {

	vr := buildValidationResult(rule, constants.ValidationTypeRolePrivileges)
	valid := isValidRule(rule, privileges)
	if !valid {
		vr.State = ptr.Ptr(v8or.ValidationFailed)
		vr.Condition.Failures = append(vr.Condition.Failures, fmt.Sprintf("Rule: %s, was not found in the user's privileges", rule.Name))
		vr.Condition.Message = "One or more required privileges was not found, or a condition was not met"
		vr.Condition.Status = corev1.ConditionFalse

		return vr, errors.New("Rule not valid")
	}

	return vr, nil
}

func isValidRule(rule v1alpha1.GenericRolePrivilegeValidationRule, privileges map[string]bool) bool {
	return privileges[rule.Name]
}

func getUserPrivileges(vsphereCloudDriver *vsphere.VSphereCloudDriver, authManager *object.AuthorizationManager, datacenter, userName string) (map[string]bool, error) {
	// Get list of roles for current user
	userPrivileges, err := vsphereCloudDriver.GetVmwareUserPrivileges(userName, datacenter, vsphereCloudDriver, authManager)
	if err != nil {
		return nil, err
	}

	return userPrivileges, nil
}
