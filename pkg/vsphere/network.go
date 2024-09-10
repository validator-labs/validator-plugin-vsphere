package vsphere

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
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

// GetNetworkTypeByName determines the type of a network given its datacenter and name.
func (v *VCenterDriver) GetNetworkTypeByName(ctx context.Context, datacenter, name string) (string, error) {
	finder, dc, err := v.GetFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return "", err
	}

	inventoryPath := fmt.Sprintf("/%s/network/%s", dc, name)

	nr, err := finder.Network(ctx, inventoryPath)
	if err != nil {
		return "", fmt.Errorf("failed to lookup network %s: %w", inventoryPath, err)
	}

	switch nr.(type) {
	case *object.Network:
		return "Network", nil
	case *object.DistributedVirtualPortgroup:
		return "Distributed Port Group", nil
	case *object.DistributedVirtualSwitch:
		return "Distributed Switch", nil
	case *object.OpaqueNetwork:
		return "Opaque Network", nil
	default:
		return "", fmt.Errorf("unsupported network type %T", nr)
	}
}

// GetNetworks returns a sorted list of all vCenter networks.
func (v *VCenterDriver) GetNetworks(ctx context.Context, datacenter string) ([]string, error) {
	prefix, networkRefs, err := v.getNetworkReferences(ctx, datacenter)
	if err != nil {
		return nil, err
	}

	networks := make([]string, 0)
	for _, n := range networkRefs {
		_, ok := n.(*object.Network)
		if !ok {
			continue
		}
		network := strings.TrimPrefix(n.GetInventoryPath(), prefix)
		networks = append(networks, network)
	}

	return networks, nil
}

func (v *VCenterDriver) getNetworkReferences(ctx context.Context, datacenter string) (string, []object.NetworkReference, error) {
	finder, dc, err := v.GetFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return "", nil, err
	}
	prefix := fmt.Sprintf("/%s/network/", dc)

	ns, err := finder.NetworkList(ctx, "*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch vCenter networks: %w", err)
	}
	if len(ns) == 0 {
		return "", nil, errors.New("No networks found")
	}

	return prefix, ns, nil
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

// GetDistributedVirtualPortgroups returns a sorted list of all vCenter distributed port groups.
func (v *VCenterDriver) GetDistributedVirtualPortgroups(ctx context.Context, datacenter string) ([]string, error) {
	prefix, networkRefs, err := v.getNetworkReferences(ctx, datacenter)
	if err != nil {
		return nil, err
	}

	networks := make([]string, 0)
	for _, n := range networkRefs {
		_, ok := n.(*object.DistributedVirtualPortgroup)
		if !ok {
			continue
		}
		network := strings.TrimPrefix(n.GetInventoryPath(), prefix)
		networks = append(networks, network)
	}

	return networks, nil
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

// GetDistributedVirtualSwitches returns a sorted list of all vCenter distributed switches.
func (v *VCenterDriver) GetDistributedVirtualSwitches(ctx context.Context, datacenter string) ([]string, error) {
	prefix, networkRefs, err := v.getNetworkReferences(ctx, datacenter)
	if err != nil {
		return nil, err
	}

	networks := make([]string, 0)
	for _, n := range networkRefs {
		_, ok := n.(*object.DistributedVirtualSwitch)
		if !ok {
			continue
		}
		network := strings.TrimPrefix(n.GetInventoryPath(), prefix)
		networks = append(networks, network)
	}

	return networks, nil
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

// GetOpaqueNetworks returns a sorted list of all vCenter opaque networks.
func (v *VCenterDriver) GetOpaqueNetworks(ctx context.Context, datacenter string) ([]string, error) {
	prefix, networkRefs, err := v.getNetworkReferences(ctx, datacenter)
	if err != nil {
		return nil, err
	}

	networks := make([]string, 0)
	for _, n := range networkRefs {
		_, ok := n.(*object.OpaqueNetwork)
		if !ok {
			continue
		}
		network := strings.TrimPrefix(n.GetInventoryPath(), prefix)
		networks = append(networks, network)
	}

	return networks, nil
}
