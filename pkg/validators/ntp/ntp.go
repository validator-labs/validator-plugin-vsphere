// Package ntp handles NTP validation rule reconciliation.
package ntp

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/find"
	corev1 "k8s.io/api/core/v1"

	"github.com/validator-labs/validator-plugin-vsphere/api/v1alpha1"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/constants"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/vsphere"
	vapi "github.com/validator-labs/validator/api/v1alpha1"
	vapiconstants "github.com/validator-labs/validator/pkg/constants"
	"github.com/validator-labs/validator/pkg/types"
	"github.com/validator-labs/validator/pkg/util"
)

// ValidationService is a service that validates NTP rules
type ValidationService struct {
	log        logr.Logger
	driver     *vsphere.VCenterDriver
	datacenter string
}

// NewValidationService creates a new ValidationService
func NewValidationService(log logr.Logger, driver *vsphere.VCenterDriver, datacenter string) *ValidationService {
	return &ValidationService{
		log:        log,
		driver:     driver,
		datacenter: datacenter,
	}
}

func buildValidationResult(rule v1alpha1.NTPValidationRule) *types.ValidationRuleResult {
	state := vapi.ValidationSucceeded
	validationType := constants.ValidationTypeNTP

	validationRule := fmt.Sprintf("%s-%s-%s", vapiconstants.ValidationRulePrefix, validationType, rule.Name())

	latestCondition := vapi.DefaultValidationCondition()
	latestCondition.Message = "All required NTP rules were satisfied"
	latestCondition.ValidationRule = util.Sanitize(validationRule)
	latestCondition.ValidationType = validationType

	return &types.ValidationRuleResult{Condition: &latestCondition, State: &state}
}

// ReconcileNTPRule reconciles the NTP rule
func (n *ValidationService) ReconcileNTPRule(rule v1alpha1.NTPValidationRule, finder *find.Finder) (*types.ValidationRuleResult, error) {
	var err error
	vr := buildValidationResult(rule)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	valid, failures, err := n.driver.ValidateHostNTPSettings(ctx, finder, n.datacenter, rule.ClusterName, rule.Hosts)
	if !valid {
		vr.Condition.Failures = failures
	}

	if len(vr.Condition.Failures) > 0 {
		vr.State = util.Ptr(vapi.ValidationFailed)
		vr.Condition.Message = fmt.Sprintf("One or more NTP rules were not satisfied for rule: %s", rule.Name())
		vr.Condition.Status = corev1.ConditionFalse
		err = fmt.Errorf("one or more NTP rules were not satisfied for rule: %s", rule.Name())
	}

	return vr, err
}
