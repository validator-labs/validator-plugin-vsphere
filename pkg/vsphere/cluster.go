package vsphere

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
)

// GetCluster returns the cluster if it exists
func (v *VCenterDriver) GetCluster(ctx context.Context, finder *find.Finder, datacenter, clusterName string) (*object.ClusterComputeResource, error) {
	path := fmt.Sprintf("/%s/host/%s", datacenter, clusterName)
	cluster, err := finder.ClusterComputeResource(ctx, path)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

// GetClusters returns a sorted list of all vCenter clusters within a datacenter.
func (v *VCenterDriver) GetClusters(ctx context.Context, datacenter string) ([]string, error) {
	prefix, ccrs, err := v.getClusterComputeResources(ctx, datacenter)
	if err != nil {
		return nil, err
	}

	clusters := make([]string, 0)
	for _, ccr := range ccrs {
		cluster := strings.TrimPrefix(ccr.InventoryPath, prefix)
		clusters = append(clusters, cluster)
	}

	sort.Strings(clusters)
	return clusters, nil
}

// GetClustersByTag returns a sorted list of vCenter clusters within a datacenter, filtered by a tag category.
func (v *VCenterDriver) GetClustersByTag(ctx context.Context, datacenter, tagCategory string) ([]string, error) {
	prefix, ccrs, err := v.getClusterComputeResources(ctx, datacenter)
	if err != nil {
		return nil, err
	}
	client := ccrs[0].Client()

	tags, categoryID, err := v.getTagsAndCategory(ctx, client, "ClusterComputeResource", tagCategory)
	if err != nil {
		return nil, err
	}

	clusters := make([]string, 0)
	for _, ccr := range ccrs {
		if !v.ifTagHasCategory(tags[ccr.Reference().Value].Tags, categoryID) {
			continue
		}
		cluster := strings.TrimPrefix(ccr.InventoryPath, prefix)
		clusters = append(clusters, cluster)
	}
	if len(clusters) == 0 {
		return nil, errors.Errorf("no compute clusters with tag category %s found", tagCategory)
	}

	sort.Strings(clusters)
	return clusters, nil
}

func (v *VCenterDriver) getClusterComputeResources(ctx context.Context, datacenter string) (string, []*object.ClusterComputeResource, error) {
	finder, dc, err := v.GetFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return "", nil, err
	}
	prefix := fmt.Sprintf("/%s/host/", dc)

	ccrs, err := finder.ClusterComputeResourceList(ctx, "*")
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to fetch vCenter clusters")
	}
	if len(ccrs) == 0 {
		return "", nil, errors.New("no compute clusters found")
	}

	return prefix, ccrs, nil
}
