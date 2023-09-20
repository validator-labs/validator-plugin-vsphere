package tags

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/api/v1alpha1"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/internal/constants"
	v8or "github.com/spectrocloud-labs/valid8or/api/v1alpha1"
	v8orconstants "github.com/spectrocloud-labs/valid8or/pkg/constants"
	"github.com/spectrocloud-labs/valid8or/pkg/types"
	v8ortypes "github.com/spectrocloud-labs/valid8or/pkg/types"
	"github.com/spectrocloud-labs/valid8or/pkg/util/ptr"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/mo"
	corev1 "k8s.io/api/core/v1"
)

// to enable monkey patching in integration tests
var GetCategories = getCategories
var GetAttachedTagsOnObjects = getAttachedTagsOnObjects

type RegionZoneCategoryExistsInput struct {
	Datacenter         string
	Cluster            []string
	RegionCategoryName string
	ZoneCategoryName   string
}

type TagsValidationService struct {
	Log logr.Logger
}

func NewTagsValidationService(log logr.Logger) *TagsValidationService {
	return &TagsValidationService{
		Log: log,
	}
}

func (s *TagsValidationService) ReconcileRegionZoneTagRules(tagsManager *tags.Manager, finder *find.Finder, regionZoneValidationRule v1alpha1.RegionZoneValidationRule) (*types.ValidationResult, error) {
	vr := buildValidationResult(constants.ValidationTypeTag)

	input := RegionZoneCategoryExistsInput{
		RegionCategoryName: regionZoneValidationRule.RegionCategoryName,
		ZoneCategoryName:   regionZoneValidationRule.ZoneCategoryName,
		Datacenter:         regionZoneValidationRule.Datacenter,
		Cluster:            regionZoneValidationRule.Clusters,
	}

	valid, err := RegionZoneCategoryExists(tagsManager, finder, input)
	if !valid {
		vr.State = ptr.Ptr(v8or.ValidationFailed)
		vr.Condition.Failures = append(vr.Condition.Failures, "One or more required tags was not found")
		vr.Condition.Message = "One or more required tags was not found"
		vr.Condition.Status = corev1.ConditionFalse
		return vr, err
	}

	s.Log.V(0).Info("Region and Zone tags exist")
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

func RegionZoneCategoryExists(tagsManager *tags.Manager, finder *find.Finder, input RegionZoneCategoryExistsInput) (bool, error) {
	regionCategoryID, zoneCategoryID := "", ""

	cats, err := GetCategories(tagsManager)
	if err != nil {
		return false, err
	}
	var regionZoneTags []tags.Category
	for _, category := range cats {
		switch category.Name {
		case input.RegionCategoryName:
			regionCategoryID = category.ID
			regionZoneTags = append(regionZoneTags, category)
		case input.ZoneCategoryName:
			zoneCategoryID = category.ID
			regionZoneTags = append(regionZoneTags, category)
		}
	}

	if len(regionZoneTags) < 2 {
		return false, errors.New("one or more region/zone tags don't exist")
	}

	// check if datacenter has region tag
	list, err := finder.ManagedObjectList(context.TODO(), fmt.Sprintf("/%s", input.Datacenter))
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
	isDatacenterTaggedWithRegion := false

	for _, attachedTag := range attachedTags {
		for _, tagName := range attachedTag.Tags {
			if tagName.CategoryID == regionCategoryID {
				isDatacenterTaggedWithRegion = true
				break
			}
		}
	}

	// check if all compute clusters has zone tag
	areComputeClustersTaggedWithZone := true
	for _, cluster := range input.Cluster {
		list, err = finder.ManagedObjectList(context.TODO(), fmt.Sprintf("/%s/host/%s", input.Datacenter, cluster))
		if err != nil {
			return false, err
		}
		// return early if no can't find the managedobject list
		if len(list) == 0 {
			return false, nil
		}
		refs = nil
		refs = append(refs, list[0].Object.Reference())
		clusterattachedTags, err := GetAttachedTagsOnObjects(tagsManager, refs)
		if err != nil {
			return false, err
		}
		found := false
		for _, tag := range clusterattachedTags {
			if found {
				break
			}
			for _, tagName := range tag.Tags {
				if tagName.CategoryID == zoneCategoryID {
					found = true
					break
				}
			}
		}
		areComputeClustersTaggedWithZone = areComputeClustersTaggedWithZone && found
	}

	if areComputeClustersTaggedWithZone && isDatacenterTaggedWithRegion && len(regionZoneTags) >= 2 {
		return true, nil
	}

	return false, errors.New("region zone categories don't exist")
}

func getAttachedTagsOnObjects(tagsManager *tags.Manager, refs []mo.Reference) ([]tags.AttachedTags, error) {
	return tagsManager.GetAttachedTagsOnObjects(context.TODO(), refs)
}

func getCategories(tm *tags.Manager) ([]tags.Category, error) {
	return tm.GetCategories(context.TODO())
}
