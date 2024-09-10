// Package tags handles tag validation rule reconciliation.
package tags

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/mo"
	corev1 "k8s.io/api/core/v1"

	vapi "github.com/validator-labs/validator/api/v1alpha1"
	vapiconstants "github.com/validator-labs/validator/pkg/constants"
	vapitypes "github.com/validator-labs/validator/pkg/types"
	"github.com/validator-labs/validator/pkg/util"

	"github.com/validator-labs/validator-plugin-vsphere/api/v1alpha1"
	"github.com/validator-labs/validator-plugin-vsphere/api/vcenter"
	"github.com/validator-labs/validator-plugin-vsphere/api/vcenter/entity"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/constants"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/vsphere"
)

var (
	// GetCategories is defined to enable monkey patching the getCategories function in integration tests
	GetCategories = getCategories

	// GetAttachedTagsOnObjects is defined to enable monkey patching the getAttachedTagsOnObjects function in integration tests
	GetAttachedTagsOnObjects = getAttachedTagsOnObjects

	// SupportedEntities indicates which vCenter entities may be referenced in a tag validation rule.
	SupportedEntities = []string{
		entity.Cluster.String(),
		entity.Datacenter.String(),
		entity.Folder.String(),
		entity.Host.String(),
		entity.ResourcePool.String(),
		entity.VirtualMachine.String(),
	}
)

// ValidationService is a service that validates tag rules
type ValidationService struct {
	Log logr.Logger
}

// NewValidationService creates a new ValidationService
func NewValidationService(log logr.Logger) *ValidationService {
	return &ValidationService{
		Log: log,
	}
}

// ReconcileTagRules reconciles the tag rules
func (s *ValidationService) ReconcileTagRules(tagsManager *tags.Manager, finder *find.Finder, driver *vsphere.VCenterDriver, rule v1alpha1.TagValidationRule) (*vapitypes.ValidationRuleResult, error) {
	vr := buildValidationResult(rule, constants.ValidationTypeTag)

	valid, err := tagIsValid(tagsManager, finder, driver.Datacenter, rule)
	if !valid {
		vr.State = util.Ptr(vapi.ValidationFailed)
		vr.Condition.Failures = append(vr.Condition.Failures, "One or more required tags was not found")
		vr.Condition.Message = "One or more required tags was not found"
		vr.Condition.Status = corev1.ConditionFalse
		return vr, err
	}

	s.Log.V(0).Info("Entity tags exist")
	return vr, nil
}

func buildValidationResult(rule v1alpha1.TagValidationRule, validationType string) *vapitypes.ValidationRuleResult {
	state := vapi.ValidationSucceeded

	validationRule := fmt.Sprintf("%s-%s-%s-%s", vapiconstants.ValidationRulePrefix, "tag", rule.EntityType, rule.Tag)

	latestCondition := vapi.DefaultValidationCondition()
	latestCondition.Message = "Required entity tags were found"
	latestCondition.ValidationRule = util.Sanitize(validationRule)
	latestCondition.ValidationType = validationType

	return &vapitypes.ValidationRuleResult{Condition: &latestCondition, State: &state}
}

func tagIsValid(tagsManager *tags.Manager, finder *find.Finder, datacenter string, rule v1alpha1.TagValidationRule) (bool, error) {
	categoryID := ""
	var inventoryPath string

	cats, err := GetCategories(tagsManager)
	if err != nil {
		return false, err
	}
	for _, category := range cats {
		switch category.Name {
		case rule.Tag:
			categoryID = category.ID
		}
	}

	switch rule.EntityType {
	case entity.Cluster:
		inventoryPath = fmt.Sprintf(vcenter.ClusterInventoryPath, datacenter, rule.EntityName)
	case entity.Datacenter:
		inventoryPath = rule.EntityName
	case entity.Folder:
		inventoryPath = rule.EntityName
	case entity.Host:
		inventoryPath = fmt.Sprintf(vcenter.HostSystemInventoryPath, datacenter, rule.ClusterName, rule.EntityName)
	case entity.ResourcePool:
		inventoryPath = fmt.Sprintf(vcenter.ResourcePoolInventoryPath, datacenter, rule.ClusterName, rule.EntityName)
		if rule.EntityName == vcenter.ClusterDefaultResourcePoolName {
			inventoryPath = fmt.Sprintf("/%s/host/%s/%s", datacenter, rule.ClusterName, rule.EntityName)
		}
	case entity.VirtualMachine:
		inventoryPath = rule.EntityName
	default:
		return false, fmt.Errorf("unsupported entity type: %s", rule.EntityType)
	}

	// check if object has tag
	list, err := finder.ManagedObjectList(context.TODO(), inventoryPath)
	if err != nil {
		return false, err
	}

	// return early if no can't find the managedobject list
	if len(list) == 0 {
		return false, nil
	}
	var refs []mo.Reference
	refs = append(refs, list[0].Object.Reference())
	attachedTags, err := GetAttachedTagsOnObjects(tagsManager, refs)
	if err != nil {
		return false, err
	}
	isEntityTagged := false

	for _, attachedTag := range attachedTags {
		for _, tagName := range attachedTag.Tags {
			if tagName.CategoryID == categoryID {
				isEntityTagged = true
				break
			}
		}
	}

	if isEntityTagged {
		return true, nil
	}

	return false, errors.New("entity tags don't exist")
}

func getAttachedTagsOnObjects(tagsManager *tags.Manager, refs []mo.Reference) ([]tags.AttachedTags, error) {
	return tagsManager.GetAttachedTagsOnObjects(context.TODO(), refs)
}

func getCategories(tm *tags.Manager) ([]tags.Category, error) {
	return tm.GetCategories(context.TODO())
}
