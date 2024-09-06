package vsphere

import (
	"context"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
)

// GetVApp returns the virtual app if it exists
func (v *VCenterDriver) GetVApp(ctx context.Context, finder *find.Finder, vAppName string) (*object.VirtualApp, error) {
	vApp, err := finder.VirtualApp(ctx, vAppName)
	if err != nil {
		return nil, err
	}
	return vApp, nil
}

// GetVApps returns a list of virtual apps
func (v *VCenterDriver) GetVApps(ctx context.Context) ([]mo.VirtualApp, error) {
	m := view.NewManager(v.Client.Client)

	containerView, err := m.CreateContainerView(ctx, v.Client.Client.ServiceContent.RootFolder, []string{"VirtualApp"}, true)
	if err != nil {
		return nil, err
	}

	var vApps []mo.VirtualApp
	if err := containerView.Retrieve(ctx, []string{"VirtualApp"}, nil, &vApps); err != nil {
		return nil, err
	}
	return vApps, nil
}
