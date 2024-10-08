// Package privileges handles reconciliation of PrivilegeValidationRules.
package privileges

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	corev1 "k8s.io/api/core/v1"

	vapi "github.com/validator-labs/validator/api/v1alpha1"
	vapiconstants "github.com/validator-labs/validator/pkg/constants"
	"github.com/validator-labs/validator/pkg/types"
	"github.com/validator-labs/validator/pkg/util"

	"github.com/validator-labs/validator-plugin-vsphere/api/v1alpha1"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/constants"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/vsphere"
)

// PrivilegeValidationService is a service that validates user privileges
type PrivilegeValidationService struct {
	log         logr.Logger
	driver      *vsphere.VCenterDriver
	datacenter  string
	authManager *object.AuthorizationManager
	username    string
}

// NewPrivilegeValidationService creates a new PrivilegeValidationService
func NewPrivilegeValidationService(log logr.Logger, driver *vsphere.VCenterDriver, datacenter, username string, authManager *object.AuthorizationManager) *PrivilegeValidationService {
	return &PrivilegeValidationService{
		log:         log,
		driver:      driver,
		datacenter:  datacenter,
		authManager: authManager,
		username:    username,
	}
}

// ReconcilePrivilegeRule reconciles a privilege rule
func (s *PrivilegeValidationService) ReconcilePrivilegeRule(rule v1alpha1.PrivilegeValidationRule, finder *find.Finder) (*types.ValidationRuleResult, error) {
	var err error
	vr := buildValidationResult(rule, s.username)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vr.Condition.Failures, err = s.driver.ValidateUserPrivilegeOnEntities(ctx, s.authManager, s.datacenter, s.username, finder, rule)

	if len(vr.Condition.Failures) > 0 {
		vr.State = util.Ptr(vapi.ValidationFailed)
		vr.Condition.Message = fmt.Sprintf("One or more required privileges was not found, or a condition was not met for account: %s", s.username)
		vr.Condition.Status = corev1.ConditionFalse
	}

	return vr, err
}

func buildValidationResult(rule v1alpha1.PrivilegeValidationRule, username string) *types.ValidationRuleResult {
	state := vapi.ValidationSucceeded
	validationType := constants.ValidationTypePrivileges

	validationRule := fmt.Sprintf("%s-%s-%s", vapiconstants.ValidationRulePrefix, validationType, rule.EntityType)
	if rule.EntityName != "" {
		validationRule = fmt.Sprintf("%s-%s", validationRule, rule.EntityName)
	}

	latestCondition := vapi.DefaultValidationCondition()
	latestCondition.Message = fmt.Sprintf("All required %s permissions were found for account: %s", constants.ValidationTypePrivileges, username)
	latestCondition.ValidationRule = util.Sanitize(validationRule)
	latestCondition.ValidationType = validationType

	return &types.ValidationRuleResult{Condition: &latestCondition, State: &state}
}
