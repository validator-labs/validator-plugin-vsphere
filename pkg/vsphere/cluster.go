package vsphere

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
)

func (v *VSphereCloudDriver) GetClusterIfExists(ctx context.Context, finder *find.Finder, datacenter, clusterName string) (bool, *object.ClusterComputeResource, error) {
	path := fmt.Sprintf("/%s/host/%s", datacenter, clusterName)
	cluster, err := finder.ClusterComputeResource(ctx, path)
	if err != nil {
		return false, nil, err
	}
	return true, cluster, nil
}

func (v *VSphereCloudDriver) GetVSphereClusters(ctx context.Context, datacenter string) ([]string, error) {
	finder, dc, err := v.GetFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return nil, err
	}

	ccrs, err := finder.ClusterComputeResourceList(ctx, "*")
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch vSphere clusters")
	}

	if len(ccrs) == 0 {
		return nil, errors.New("No compute clusters found")
	}

	client := ccrs[0].Client()

	tags, categoryId, err := v.getTagsAndCategory(ctx, client, "ClusterComputeResource", ComputeClusterTagCategory)
	if err != nil {
		return nil, err
	}

	clusters := make([]string, 0)
	for _, ccr := range ccrs {
		if v.ifTagHasCategory(tags[ccr.Reference().Value].Tags, categoryId) {
			prefix := fmt.Sprintf("/%s/host/", dc)
			cluster := strings.TrimPrefix(ccr.InventoryPath, prefix)
			clusters = append(clusters, cluster)
		}
	}

	if len(clusters) == 0 {
		return nil, errors.Errorf("No compute clusters with tag category %s found", ComputeClusterTagCategory)
	}

	sort.Strings(clusters)
	return clusters, nil
}

func (v *VSphereCloudDriver) getClusterDatastores(ctx context.Context, finder *find.Finder, datacenter string, cluster mo.ClusterComputeResource) (datastores []string, err error) {
	dsMobjRefs := cluster.Datastore

	for i := range dsMobjRefs {
		inventoryPath := ""
		dsObjRef, err := finder.ObjectReference(ctx, dsMobjRefs[i])
		if err != nil {
			return nil, fmt.Errorf("error: %s, code: %d", err.Error(), http.StatusBadRequest)
		}
		if dsObjRef != nil {
			ref := dsObjRef
			switch ref.(type) {
			case *object.Datastore:
				n := dsObjRef.(*object.Datastore)
				inventoryPath = n.InventoryPath
			default:
				continue
			}

			if inventoryPath != "" {
				prefix := fmt.Sprintf("/%s/datastore/", datacenter)
				datastore := strings.TrimPrefix(inventoryPath, prefix)
				datastores = append(datastores, datastore)
			}
		}
	}

	sort.Strings(datastores)
	return datastores, nil
}

func (v *VSphereCloudDriver) getClusterComputeResources(ctx context.Context, finder *find.Finder) ([]*object.ClusterComputeResource, error) {
	ccrs, err := finder.ClusterComputeResourceList(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("failed to get compute cluster resources: %s", err.Error())
	}
	return ccrs, nil
}
