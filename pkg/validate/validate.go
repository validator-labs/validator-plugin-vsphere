// Package validate defines a Validate function that evaluates a VsphereValidatorSpec and returns a ValidationResponse.
package validate

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/object"
	vtags "github.com/vmware/govmomi/vapi/tags"

	"github.com/validator-labs/validator/pkg/types"

	"github.com/validator-labs/validator-plugin-vsphere/api/v1alpha1"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/validators/computeresources"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/validators/ntp"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/validators/privileges"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/validators/tags"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/vsphere"
)

// Validate validates the VsphereValidatorSpec and returns a ValidationResponse.
func Validate(ctx context.Context, spec v1alpha1.VsphereValidatorSpec, driver *vsphere.CloudDriver, log logr.Logger) (types.ValidationResponse, error) {
	resp := types.ValidationResponse{
		ValidationRuleResults: make([]*types.ValidationRuleResult, 0, spec.ResultCount()),
		ValidationRuleErrors:  make([]error, 0, spec.ResultCount()),
	}

	// Get the authorization manager from the Client
	authManager := object.NewAuthorizationManager(driver.Client.Client)
	if authManager == nil {
		return resp, errors.New("invalid vSphere driver; failed to get vim25 authorization manager")
	}

	// Get a finder for the datacenter
	finder, _, err := driver.GetFinderWithDatacenter(ctx, driver.Datacenter)
	if err != nil {
		return resp, fmt.Errorf("failed to get finder with datacenter: %w", err)
	}

	tagsManager := vtags.NewManager(driver.RestClient)

	// Get the current user
	userName, err := driver.GetCurrentVmwareUser(ctx)
	if err != nil {
		return resp, fmt.Errorf("failed to get current user: %w", err)
	}

	// NTP validation rules
	ntpValidationService := ntp.NewValidationService(log, driver, spec.Datacenter)
	for _, rule := range spec.NTPValidationRules {
		vrr, err := ntpValidationService.ReconcileNTPRule(rule, finder)
		if err != nil {
			log.Error(err, "failed to reconcile NTP rule")
		}
		resp.AddResult(vrr, err)
		log.Info("Validated NTP rules")
	}

	// Entity privilege validation rules
	rolePrivilegeValidationService := privileges.NewPrivilegeValidationService(
		log, driver, spec.Datacenter, authManager, userName,
	)
	for _, rule := range spec.EntityPrivilegeValidationRules {
		vrr, err := rolePrivilegeValidationService.ReconcileEntityPrivilegeRule(rule, finder)
		if err != nil {
			log.Error(err, "failed to reconcile entity privilege rule")
		}
		resp.AddResult(vrr, err)
		log.Info("Validated privileges for account", "user", rule.Username)
	}

	// Role privilege validation rules
	for _, rule := range spec.RolePrivilegeValidationRules {
		vrr, err := rolePrivilegeValidationService.ReconcileRolePrivilegesRule(rule, driver, authManager)
		if err != nil {
			log.Error(err, "failed to reconcile role privilege rule")
		}
		resp.AddResult(vrr, err)
		log.Info("Validated privileges for account", "user", rule.Username)
	}

	// Tag validation rules
	tagValidationService := tags.NewValidationService(log)
	for _, rule := range spec.TagValidationRules {
		log.Info("Checking if tags are properly assigned")
		vrr, err := tagValidationService.ReconcileTagRules(tagsManager, finder, driver, rule)
		if err != nil {
			log.Error(err, "failed to reconcile tag validation rule")
		}
		resp.AddResult(vrr, err)
		log.Info("Validated tags", "entity type", rule.EntityType, "entity name", rule.EntityName, "tag", rule.Tag)
	}

	// Compute resource validation rules
	computeResourceValidationService := computeresources.NewValidationService(log, driver)
	for _, rule := range spec.ComputeResourceRules {
		vrr, err := computeResourceValidationService.ReconcileComputeResourceValidationRule(rule, finder, driver)
		if err != nil {
			log.Error(err, "failed to reconcile computeresources validation rule")
		}
		resp.AddResult(vrr, err)
		log.Info("Validated compute resources", "scope", rule.Scope, "entity name", rule.EntityName)
	}

	return resp, nil
}
