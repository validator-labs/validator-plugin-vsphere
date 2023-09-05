package roleprivilege

import (
	"context"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/go-logr/logr"
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
	"sort"
)

type VMwareRolePrivilege struct {
	rule       v1alpha1.RolePrivilegeValidationRule
	Privileges map[string]bool
}

type RolePrivilegeValidationService struct {
	log    logr.Logger
	driver *vsphere.VSphereCloudDriver
}

func NewRolePrivilegeValidationService(log logr.Logger, driver *vsphere.VSphereCloudDriver) *RolePrivilegeValidationService {
	return &RolePrivilegeValidationService{
		log:    log,
		driver: driver,
	}
}

func buildValidationResult(rule v1alpha1.RolePrivilegeValidationRule, validationType string) *types.ValidationResult {
	state := v8or.ValidationSucceeded
	latestCondition := v8or.DefaultValidationCondition()
	latestCondition.Message = fmt.Sprintf("All required %s permissions were found", validationType)
	latestCondition.ValidationRule = fmt.Sprintf("%s-%s", v8orconstants.ValidationRulePrefix, rule.Name)
	latestCondition.ValidationType = validationType

	return &types.ValidationResult{Condition: &latestCondition, State: &state}
}

func (s *RolePrivilegeValidationService) GetUserRolePrivilegesMapping() (map[string]bool, error) {
	privileges, err := getUserPrivileges(s.driver)
	if err != nil {
		return nil, err
	}
	return privileges, nil
}

func (s *RolePrivilegeValidationService) ReconcileRolePrivilegesRule(rule v1alpha1.RolePrivilegeValidationRule, privileges map[string]bool) (*types.ValidationResult, error) {

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

func isValidRule(rule v1alpha1.RolePrivilegeValidationRule, privileges map[string]bool) bool {
	// convert the keys of the map to a slice of strings
	keys := make([]string, 0, len(privileges))
	for k := range privileges {
		keys = append(keys, k)
	}

	// sort the slice of keys
	sort.Strings(keys)

	if rule.IsEnabled {
		switch rule.RuleType {
		case "VMwareRolePrivilege":
			rolePrivilegeRule := VMwareRolePrivilege{}
			rolePrivilegeRule.rule = rule
			rolePrivilegeRule.Privileges = privileges
			return rolePrivilegeRule.validateVMwareRolePrivilege()
		}
	}

	return false
}

func (v *VMwareRolePrivilege) validateVMwareRolePrivilege() bool {
	data := map[string]interface{}{
		"vmware_user_privileges": toSlice(v.Privileges),
	}
	for _, expr := range v.rule.Expressions {
		expression, err := govaluate.NewEvaluableExpression(expr)
		if err != nil {
			return false
		} else {
			result, err := expression.Evaluate(data)
			if err != nil {
				return false
			} else {
				if result == false {
					return false
				}
			}
		}
	}
	return true
}

func toSlice(m map[string]bool) []interface{} {
	values := make([]interface{}, 0, len(m))
	for k, v := range m {
		if v {
			values = append(values, k)
		}
	}
	return values
}

func getUserPrivileges(vsphereCloudDriver *vsphere.VSphereCloudDriver) (map[string]bool, error) {

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
