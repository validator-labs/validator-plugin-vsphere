package vsphere

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
)

// GetDatastore returns a datastore object if it exists
func (v *VCenterDriver) GetDatastore(ctx context.Context, finder *find.Finder, datastore string) (*object.Datastore, error) {
	ds, err := finder.Datastore(ctx, datastore)
	if err != nil {
		return nil, err
	}
	return ds, nil
}

// GetDatastores returns a sorted list of all vCenter datastores within a datacenter.
func (v *VCenterDriver) GetDatastores(ctx context.Context, datacenter string) ([]string, error) {
	prefix, ds, err := v.getDatastores(ctx, datacenter)
	if err != nil {
		return nil, err
	}

	datastores := make([]string, len(ds))
	for i, d := range ds {
		datastore := strings.TrimPrefix(d.InventoryPath, prefix)
		datastores[i] = datastore
	}

	return datastores, nil
}

func (v *VCenterDriver) getDatastores(ctx context.Context, datacenter string) (string, []*object.Datastore, error) {
	finder, dc, err := v.GetFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return "", nil, err
	}
	prefix := fmt.Sprintf("/%s/datastore/", dc)

	ds, err := finder.DatastoreList(ctx, "*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch vCenter datastores: %w", err)
	}
	if len(ds) == 0 {
		return "", nil, errors.New("No datastores found")
	}

	return prefix, ds, nil
}
