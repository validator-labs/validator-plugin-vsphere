package privileges

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/find"
	corev1 "k8s.io/api/core/v1"

	"github.com/spectrocloud-labs/validator-plugin-vsphere/api/v1alpha1"
	"github.com/spectrocloud-labs/validator-plugin-vsphere/internal/constants"
	vapi "github.com/spectrocloud-labs/validator/api/v1alpha1"
	vapiconstants "github.com/spectrocloud-labs/validator/pkg/constants"
	"github.com/spectrocloud-labs/validator/pkg/types"
	"github.com/spectrocloud-labs/validator/pkg/util"
)

func buildEntityPrivilegeValidationResult(rule v1alpha1.EntityPrivilegeValidationRule, validationType string) *types.ValidationRuleResult {
	state := vapi.ValidationSucceeded
	latestCondition := vapi.DefaultValidationCondition()
	latestCondition.Message = fmt.Sprintf("All required %s permissions were found for account: %s", validationType, rule.Username)
	latestCondition.ValidationRule = fmt.Sprintf("%s-%s-%s", vapiconstants.ValidationRulePrefix, rule.EntityType, rule.EntityName)
	latestCondition.ValidationType = validationType

	return &types.ValidationRuleResult{Condition: &latestCondition, State: &state}
}

func (s *PrivilegeValidationService) ReconcileEntityPrivilegeRule(rule v1alpha1.EntityPrivilegeValidationRule, finder *find.Finder) (*types.ValidationRuleResult, error) {
	var err error
	vr := buildEntityPrivilegeValidationResult(rule, constants.ValidationTypeEntityPrivileges)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	valid, failures, err := s.driver.ValidateUserPrivilegeOnEntities(ctx, s.authManager, s.datacenter, finder, rule.EntityName, rule.EntityType, rule.Privileges, rule.Username, rule.ClusterName)
	if !valid {
		vr.Condition.Failures = failures
	}

	if len(vr.Condition.Failures) > 0 {
		vr.State = util.Ptr(vapi.ValidationFailed)
		vr.Condition.Message = "One or more required privileges was not found, or a condition was not met"
		vr.Condition.Status = corev1.ConditionFalse
		err = fmt.Errorf("one or more required entity privileges was not found for account: %s", rule.Username)
	}

	return vr, err
}
