package vsphere

import (
	"context"

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
