// Package validate defines a Validate function that evaluates a VsphereValidatorSpec and returns a ValidationResponse.
package validate

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/object"
	vtags "github.com/vmware/govmomi/vapi/tags"

	vapi "github.com/validator-labs/validator/api/v1alpha1"
	vconstants "github.com/validator-labs/validator/pkg/constants"
	"github.com/validator-labs/validator/pkg/types"

	"github.com/validator-labs/validator-plugin-vsphere/api/v1alpha1"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/constants"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/validators/computeresources"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/validators/ntp"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/validators/privileges"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/validators/tags"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/vsphere"
)

// Validate validates the VsphereValidatorSpec and returns a ValidationResponse.
func Validate(ctx context.Context, spec v1alpha1.VsphereValidatorSpec, log logr.Logger) types.ValidationResponse {
	resp := types.ValidationResponse{
		ValidationRuleResults: make([]*types.ValidationRuleResult, 0, spec.ResultCount()),
		ValidationRuleErrors:  make([]error, 0, spec.ResultCount()),
	}

	vrr := buildValidationResult()

	if spec.Auth.Account == nil {
		resp.AddResult(vrr, errors.New("invalid spec; account must not be nil"))
		return resp
	}

	// Create a new vCenter driver
	driver, err := vsphere.NewVCenterDriver(*spec.Auth.Account, spec.Datacenter, log)
	if err != nil {
		resp.AddResult(vrr, fmt.Errorf("failed to create vCenter driver: %w", err))
		return resp
	}

	// Get the authorization manager from the driver
	authManager := object.NewAuthorizationManager(driver.Client.Client)
	if authManager == nil {
		resp.AddResult(vrr, errors.New("invalid vCenter driver; failed to get vim25 authorization manager"))
		return resp
	}

	// Get a finder for the datacenter
	finder, _, err := driver.GetFinderWithDatacenter(ctx, driver.Datacenter)
	if err != nil {
		resp.AddResult(vrr, fmt.Errorf("failed to get finder with datacenter: %w", err))
		return resp
	}

	tagsManager := vtags.NewManager(driver.RestClient)

	// Get the current user
	username, err := driver.CurrentUser(ctx)
	if err != nil {
		resp.AddResult(vrr, fmt.Errorf("failed to get current user: %w", err))
		return resp
	}

	// NTP validation rules
	ntpValidationService := ntp.NewValidationService(log, driver, spec.Datacenter)
	for _, rule := range spec.NTPValidationRules {
		vrr, err := ntpValidationService.ReconcileNTPRule(rule, finder)
		if err != nil {
			log.Error(err, "failed to reconcile NTP rule")
		}
		vrr.Finalize(err)
		resp.AddResult(vrr, err)
		log.Info("Validated NTP rules")
	}

	// Privilege validation rules
	privilegeValidationService := privileges.NewPrivilegeValidationService(
		log, driver, spec.Datacenter, username, authManager,
	)
	for _, rule := range spec.PrivilegeValidationRules {
		vrr, err := privilegeValidationService.ReconcilePrivilegeRule(rule, finder)
		if err != nil {
			log.Error(err, "failed to reconcile privilege rule")
		}
		vrr.Finalize(err)
		resp.AddResult(vrr, err)
		log.Info("Validated privileges for account", "user", username)
	}

	// Tag validation rules
	tagValidationService := tags.NewValidationService(log)
	for _, rule := range spec.TagValidationRules {
		log.Info("Checking if tags are properly assigned")
		vrr, err := tagValidationService.ReconcileTagRules(tagsManager, finder, driver, rule)
		if err != nil {
			log.Error(err, "failed to reconcile tag validation rule")
		}
		vrr.Finalize(err)
		resp.AddResult(vrr, err)
		log.Info("Validated tags", "entity type", rule.EntityType, "entity name", rule.EntityName, "tag", rule.Tag)
	}

	// Compute resource validation rules
	computeResourceValidationService := computeresources.NewValidationService(log, driver)
	seenScope := make(map[string]bool, 0)
	for _, rule := range spec.ComputeResourceRules {
		vrr, err := computeResourceValidationService.ReconcileComputeResourceValidationRule(rule, finder, driver, seenScope)
		if err != nil {
			log.Error(err, "failed to reconcile computeresources validation rule")
		}
		vrr.Finalize(err)
		resp.AddResult(vrr, err)
		log.Info("Validated compute resources", "scope", rule.Scope, "entity name", rule.EntityName)

		key, err := computeresources.GetScopeKey(rule)
		if err != nil {
			log.Error(err, "failed to get scope key for rule")
		} else {
			seenScope[key] = true
		}
	}

	return resp
}

func buildValidationResult() *types.ValidationRuleResult {
	state := vapi.ValidationSucceeded
	latestCondition := vapi.DefaultValidationCondition()
	latestCondition.Message = "Initialization succeeded"
	latestCondition.ValidationRule = fmt.Sprintf("%s-%s", vconstants.ValidationRulePrefix, constants.PluginCode)
	latestCondition.ValidationType = constants.PluginCode

	return &types.ValidationRuleResult{Condition: &latestCondition, State: &state}
}
