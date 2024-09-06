package vsphere

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
)

// GetNetwork returns a network object if it exists
func (v *VCenterDriver) GetNetwork(ctx context.Context, finder *find.Finder, path string) (*object.Network, error) {
	nr, err := finder.Network(ctx, path)
	if err != nil {
		return nil, err
	}

	network, ok := nr.(*object.Network)
	if !ok {
		return nil, fmt.Errorf("network %s is not of type *object.Network, but is %T", network, nr)
	}

	return network, nil
}

// GetDistributedVirtualPortgroup returns a distributed virtual port group object if it exists
func (v *VCenterDriver) GetDistributedVirtualPortgroup(ctx context.Context, finder *find.Finder, path string) (*object.DistributedVirtualPortgroup, error) {
	nr, err := finder.Network(ctx, path)
	if err != nil {
		return nil, err
	}

	dvp, ok := nr.(*object.DistributedVirtualPortgroup)
	if !ok {
		return nil, fmt.Errorf("network %s is not of type *object.DistributedVirtualPortgroup, but is %T", dvp, nr)
	}

	return dvp, nil
}

// GetDistributedVirtualSwitch returns a distributed virtual switch object if it exists
func (v *VCenterDriver) GetDistributedVirtualSwitch(ctx context.Context, finder *find.Finder, path string) (*object.DistributedVirtualSwitch, error) {
	nr, err := finder.Network(ctx, path)
	if err != nil {
		return nil, err
	}

	dvs, ok := nr.(*object.DistributedVirtualSwitch)
	if !ok {
		return nil, fmt.Errorf("network %s is not of type *object.DistributedVirtualSwitch, but is %T", dvs, nr)
	}

	return dvs, nil
}

// GetOpaqueNetwork returns an opaque network object if it exists
func (v *VCenterDriver) GetOpaqueNetwork(ctx context.Context, finder *find.Finder, path string) (*object.OpaqueNetwork, error) {
	nr, err := finder.Network(ctx, path)
	if err != nil {
		return nil, err
	}

	on, ok := nr.(*object.OpaqueNetwork)
	if !ok {
		return nil, fmt.Errorf("network %s is not of type *object.OpaqueNetwork, but is %T", on, nr)
	}

	return on, nil
}
