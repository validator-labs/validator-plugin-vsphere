package tags

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/api/v1alpha1"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/internal/constants"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/internal/vsphere"
	v8or "github.com/spectrocloud-labs/valid8or/api/v1alpha1"
	v8orconstants "github.com/spectrocloud-labs/valid8or/pkg/constants"
	"github.com/spectrocloud-labs/valid8or/pkg/types"
	v8ortypes "github.com/spectrocloud-labs/valid8or/pkg/types"
	"github.com/spectrocloud-labs/valid8or/pkg/util/ptr"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vapi/tags"
	corev1 "k8s.io/api/core/v1"
)

type TagsValidationService struct {
	log    logr.Logger
	driver *vsphere.VSphereCloudDriver
}

func NewTagsValidationService(log logr.Logger, driver *vsphere.VSphereCloudDriver) *TagsValidationService {
	return &TagsValidationService{
		log:    log,
		driver: driver,
	}
}

func (s *TagsValidationService) ReconcileRegionZoneTagRules(regionZoneValidationRule v1alpha1.RegionZoneValidationRule) (*types.ValidationResult, error) {
	tagsManager := tags.NewManager(s.driver.RestClient)
	finder := find.NewFinder(s.driver.Client.Client, true)

	vr := buildValidationResult(constants.ValidationTypeTag)

	input := vsphere.RegionZoneCategoryExistsInput{
		RegionCategoryName: regionZoneValidationRule.RegionCategoryName,
		ZoneCategoryName:   regionZoneValidationRule.ZoneCategoryName,
		Datacenter:         regionZoneValidationRule.Datacenter,
		Cluster:            regionZoneValidationRule.Clusters,
	}

	regionZoneCategoryExist, err := vsphere.RegionZoneCategoryExists(tagsManager, finder, input)
	if err != nil {
		vr.State = ptr.Ptr(v8or.ValidationFailed)
		vr.Condition.Failures = append(vr.Condition.Failures, "One or more required tags was not found")
		vr.Condition.Message = "One or more required tags was not found"
		vr.Condition.Status = corev1.ConditionFalse
		return nil, err
	}
	if regionZoneCategoryExist != nil && *regionZoneCategoryExist {
		s.log.V(0).Info("Region and Zone tags exist")
	}

	return vr, nil
}

func buildValidationResult(validationType string) *types.ValidationResult {
	state := v8or.ValidationSucceeded
	latestCondition := v8or.DefaultValidationCondition()
	latestCondition.Message = "All required region/zone tags were found"
	latestCondition.ValidationRule = fmt.Sprintf("%s-%s-%s", v8orconstants.ValidationRulePrefix, "region", "zone")
	latestCondition.ValidationType = validationType
	validationResult := &v8ortypes.ValidationResult{Condition: &latestCondition, State: &state}

	return validationResult
}
