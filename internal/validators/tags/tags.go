package tags

import (
	"fmt"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/api/v1alpha1"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/internal/vsphere"
	"github.com/spectrocloud-labs/valid8or/pkg/types"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vapi/tags"
)

func ReconcileRegionZoneTagRules(regionZoneValidationRule v1alpha1.RegionZoneValidationRule, vsphereCloudDriver *vsphere.VSphereCloudDriver) (*types.ValidationResult, error) {
	tagsManager := tags.NewManager(vsphereCloudDriver.RestClient)
	finder := find.NewFinder(vsphereCloudDriver.Client.Client, true)

	input := vsphere.RegionZoneCategoryExistsInput{
		RegionCategoryName: regionZoneValidationRule.RegionCategoryName,
		ZoneCategoryName:   regionZoneValidationRule.ZoneCategoryName,
		Datacenter:         regionZoneValidationRule.Datacenter,
		Cluster:            regionZoneValidationRule.Clusters,
	}

	regionZoneCategoryExist, err := vsphere.RegionZoneCategoryExists(tagsManager, finder, input)
	if err != nil {
		return nil, err
	}
	if regionZoneCategoryExist != nil && *regionZoneCategoryExist {
		fmt.Println("Region and Zone tags exist")
	}

	return nil, nil
}
