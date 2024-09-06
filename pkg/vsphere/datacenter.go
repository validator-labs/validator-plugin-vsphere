package vsphere

import (
	"context"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
)

// GetDatacenter returns a datacenter object if it exists
func (v *VCenterDriver) GetDatacenter(ctx context.Context, finder *find.Finder, datacenter string) (*object.Datacenter, error) {
	dc, err := finder.Datacenter(ctx, datacenter)
	if err != nil {
		return nil, err
	}
	return dc, nil
}

// GetK8sDatacenters returns a sorted list of kubernetes-enabled datacenters in the vCenter environment
func (v *VCenterDriver) GetK8sDatacenters(ctx context.Context) ([]string, error) {
	finder, err := v.getFinder()
	if err != nil {
		return nil, err
	}

	dcs, err := finder.DatacenterList(ctx, "*")
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch vCenter datacenters")
	}
	if len(dcs) == 0 {
		return nil, errors.New("No datacenters found")
	}

	client := dcs[0].Client()
	tags, categoryID, err := v.getTagsAndCategory(ctx, client, "Datacenter", K8sDatacenterTagCategory)
	if err != nil {
		return nil, err
	}

	datacenters := make([]string, 0)
	for _, dc := range dcs {
		if v.ifTagHasCategory(tags[dc.Reference().Value].Tags, categoryID) {
			dcName := strings.TrimPrefix(dc.InventoryPath, "/")
			datacenters = append(datacenters, dcName)
		}
	}
	if len(datacenters) == 0 {
		return nil, errors.Errorf("no datacenter with tag category %s found", K8sDatacenterTagCategory)
	}

	sort.Strings(datacenters)
	return datacenters, nil
}
