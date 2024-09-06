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

// GetK8sClusters returns a sorted list of kubernetes-enabled vCenter clusters
func (v *VCenterDriver) GetK8sClusters(ctx context.Context, datacenter string) ([]string, error) {
	finder, dc, err := v.GetFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return nil, err
	}

	ccrs, err := finder.ClusterComputeResourceList(ctx, "*")
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch vCenter clusters")
	}
	if len(ccrs) == 0 {
		return nil, errors.New("no compute clusters found")
	}

	client := ccrs[0].Client()

	tags, categoryID, err := v.getTagsAndCategory(ctx, client, "ClusterComputeResource", K8sComputeClusterTagCategory)
	if err != nil {
		return nil, err
	}

	clusters := make([]string, 0)
	for _, ccr := range ccrs {
		if v.ifTagHasCategory(tags[ccr.Reference().Value].Tags, categoryID) {
			prefix := fmt.Sprintf("/%s/host/", dc)
			cluster := strings.TrimPrefix(ccr.InventoryPath, prefix)
			clusters = append(clusters, cluster)
		}
	}
	if len(clusters) == 0 {
		return nil, errors.Errorf("no compute clusters with tag category %s found", K8sComputeClusterTagCategory)
	}

	sort.Strings(clusters)
	return clusters, nil
}

func (v *VCenterDriver) getClusterComputeResources(ctx context.Context, finder *find.Finder) ([]*object.ClusterComputeResource, error) {
	ccrs, err := finder.ClusterComputeResourceList(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("failed to get compute cluster resources: %s", err.Error())
	}
	return ccrs, nil
}
