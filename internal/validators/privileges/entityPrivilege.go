package privileges

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/api/v1alpha1"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/internal/constants"
	v8or "github.com/spectrocloud-labs/valid8or/api/v1alpha1"
	v8orconstants "github.com/spectrocloud-labs/valid8or/pkg/constants"
	"github.com/spectrocloud-labs/valid8or/pkg/types"
	"github.com/spectrocloud-labs/valid8or/pkg/util/ptr"
	"github.com/vmware/govmomi/find"
	corev1 "k8s.io/api/core/v1"
)

func buildEntityPrivilegeValidationResult(rule v1alpha1.EntityPrivilegeValidationRule, validationType string) *types.ValidationResult {
	state := v8or.ValidationSucceeded
	latestCondition := v8or.DefaultValidationCondition()
	latestCondition.Message = fmt.Sprintf("All required %s permissions were found", validationType)
	latestCondition.ValidationRule = fmt.Sprintf("%s-%s-%s", v8orconstants.ValidationRulePrefix, rule.EntityType, rule.EntityName)
	latestCondition.ValidationType = validationType

	return &types.ValidationResult{Condition: &latestCondition, State: &state}
}

func (s *PrivilegeValidationService) ReconcileEntityPrivilegeRule(rule v1alpha1.EntityPrivilegeValidationRule, finder *find.Finder) (*types.ValidationResult, error) {
	vr := buildEntityPrivilegeValidationResult(rule, constants.ValidationTypeEntityPrivileges)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	valid, err := s.driver.GetUserPrivilegeOnEntities(ctx, s.authManager, s.datacenter, finder, rule.EntityName, rule.EntityType, rule.Privileges, s.userName, rule.ClusterName)
	if !valid {
		vr.State = ptr.Ptr(v8or.ValidationFailed)
		vr.Condition.Failures = append(vr.Condition.Failures, fmt.Sprintf("Rule: %s, failed as required privileges were not found on enity: %s of type: %s", rule.Name, rule.EntityName, rule.EntityType))
		vr.Condition.Message = "One or more required privileges was not found, or a condition was not met"
		vr.Condition.Status = corev1.ConditionFalse

		return vr, errors.Errorf("Rule not valid. err: %s", err.Error())
	}
	return vr, nil
}
