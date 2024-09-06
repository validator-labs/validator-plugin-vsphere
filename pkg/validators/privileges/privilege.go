// Package privileges handles reconciliation of PrivilegeValidationRules.
package privileges

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	corev1 "k8s.io/api/core/v1"

	"github.com/validator-labs/validator-plugin-vsphere/api/v1alpha1"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/constants"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/vsphere"
	vapi "github.com/validator-labs/validator/api/v1alpha1"
	vapiconstants "github.com/validator-labs/validator/pkg/constants"
	"github.com/validator-labs/validator/pkg/types"
	"github.com/validator-labs/validator/pkg/util"
)

var errRequiredPrivilegesNotFound = errors.New("one or more required privileges was not found")

// PrivilegeValidationService is a service that validates user privileges
type PrivilegeValidationService struct {
	log         logr.Logger
	driver      *vsphere.VCenterDriver
	datacenter  string
	authManager *object.AuthorizationManager
	userName    string
}

// NewPrivilegeValidationService creates a new PrivilegeValidationService
func NewPrivilegeValidationService(log logr.Logger, driver *vsphere.VCenterDriver, datacenter string, authManager *object.AuthorizationManager, userName string) *PrivilegeValidationService {
	return &PrivilegeValidationService{
		log:         log,
		driver:      driver,
		datacenter:  datacenter,
		authManager: authManager,
		userName:    userName,
	}
}

// ReconcilePrivilegeRule reconciles a privilege rule
func (s *PrivilegeValidationService) ReconcilePrivilegeRule(rule v1alpha1.PrivilegeValidationRule, finder *find.Finder) (*types.ValidationRuleResult, error) {
	var err error
	vr := buildPrivilegeValidationResult(rule, constants.ValidationTypePrivileges)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vr.Condition.Failures, err = s.driver.ValidateUserPrivilegeOnEntities(ctx, s.authManager, s.datacenter, finder, rule)

	if len(vr.Condition.Failures) > 0 {
		vr.State = util.Ptr(vapi.ValidationFailed)
		vr.Condition.Message = fmt.Sprintf("One or more required privileges was not found, or a condition was not met for account: %s", rule.Username)
		vr.Condition.Status = corev1.ConditionFalse
		err = errRequiredPrivilegesNotFound
	}

	return vr, err
}

func buildPrivilegeValidationResult(rule v1alpha1.PrivilegeValidationRule, validationType string) *types.ValidationRuleResult {
	state := vapi.ValidationSucceeded
	latestCondition := vapi.DefaultValidationCondition()
	latestCondition.Message = fmt.Sprintf("All required %s permissions were found for account: %s", validationType, rule.Username)
	latestCondition.ValidationRule = fmt.Sprintf("%s-%s-%s", vapiconstants.ValidationRulePrefix, rule.EntityType, rule.EntityName)
	latestCondition.ValidationType = validationType

	return &types.ValidationRuleResult{Condition: &latestCondition, State: &state}
}
