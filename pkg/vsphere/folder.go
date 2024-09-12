package vsphere

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"

	"github.com/validator-labs/validator-plugin-vsphere/api/vcenter"
)

// FolderExists checks if a folder exists in the vCenter inventory
func (v *VCenterDriver) FolderExists(ctx context.Context, finder *find.Finder, name string) bool {
	if _, err := finder.Folder(ctx, name); err != nil {
		return false
	}
	return true
}

// GetFolder returns the vCenter VM folder if it exists
func (v *VCenterDriver) GetFolder(ctx context.Context, datacenter, name string) (*object.Folder, error) {
	finder, dc, err := v.GetFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return nil, err
	}
	prefix := folderPrefix(dc)

	folder, err := finder.Folder(ctx, name)
	if err != nil {
		// default to the first folder if multiple are found
		var multipleFound *find.MultipleFoundError
		if errors.As(err, &multipleFound) {
			folders, err := finder.FolderList(ctx, "*")
			if err != nil {
				return nil, err
			}
			for _, f := range folders {
				path := strings.TrimPrefix(f.InventoryPath, prefix)
				if strings.HasPrefix(path, name) {
					v.log.V(1).Info("multiple folders found; returning first match", "name", name, "path", f.InventoryPath)
					return f, nil
				}
			}
		}
		return nil, err
	}
	return folder, nil
}

// GetVMFolders returns a list of vCenter VM folders
func (v *VCenterDriver) GetVMFolders(ctx context.Context, datacenter string) ([]string, error) {
	finder, dc, err := v.GetFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return nil, err
	}
	prefix := folderPrefix(dc)

	fos, err := finder.FolderList(ctx, "*")
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to fetch vSphere folders for Datacenter %s", datacenter))
	}

	folders := make([]string, 0)
	for _, fo := range fos {
		inventoryPath := fo.InventoryPath
		// get vm folders: items with path prefix '/{Datacenter}/vm'
		if strings.HasPrefix(inventoryPath, prefix) {
			folder := strings.TrimPrefix(inventoryPath, prefix)
			folders = append(folders, folder)
		}
	}

	sort.Strings(folders)
	return folders, nil
}

func folderPrefix(datacenter string) string {
	return fmt.Sprintf(vcenter.VMFolderInventoryPrefix, datacenter)
}
