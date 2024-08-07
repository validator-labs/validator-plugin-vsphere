package vsphere

import (
	"context"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
)

// GetDatacenterIfExists returns a datacenter object if it exists
func (v *CloudDriver) GetDatacenterIfExists(ctx context.Context, finder *find.Finder, datacenter string) (bool, *object.Datacenter, error) {
	dc, err := finder.Datacenter(ctx, datacenter)
	if err != nil {
		return false, nil, err
	}
	return true, dc, nil
}

// GetVSphereDatacenters returns a sorted list of datacenters in the vSphere environment
func (v *CloudDriver) GetVSphereDatacenters(ctx context.Context) ([]string, error) {
	finder, err := v.getFinder()
	if err != nil {
		return nil, err
	}

	dcs, err := finder.DatacenterList(ctx, "*")
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch vSphere datacenters")
	}

	if len(dcs) == 0 {
		return nil, errors.New("No datacenters found")
	}

	client := dcs[0].Client()
	tags, categoryID, err := v.getTagsAndCategory(ctx, client, "Datacenter", DatacenterTagCategory)
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
		return nil, errors.Errorf("No datacenter with tag category %s found", DatacenterTagCategory)
	}

	sort.Strings(datacenters)
	return datacenters, nil
}
