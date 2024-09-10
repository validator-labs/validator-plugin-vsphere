package vsphere

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
)

// FolderExists checks if a folder exists in the vCenter inventory
func (v *VCenterDriver) FolderExists(ctx context.Context, finder *find.Finder, name string) bool {
	if _, err := finder.Folder(ctx, name); err != nil {
		return false
	}
	return true
}

// GetFolder returns the vCenter VM folder if it exists
func (v *VCenterDriver) GetFolder(ctx context.Context, finder *find.Finder, name string) (*object.Folder, error) {
	folder, err := finder.Folder(ctx, name)
	if err != nil {
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

	fos, err := finder.FolderList(ctx, "*")
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to fetch vSphere folders for Datacenter %s", datacenter))
	}

	prefix := fmt.Sprintf("/%s/vm/", dc)
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

// GetFolderNameByID returns the folder name by ID
func (v *VCenterDriver) GetFolderNameByID(ctx context.Context, datacenter, id string) (string, error) {
	finder, dc, err := v.GetFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return "", err
	}

	fos, govErr := finder.FolderList(ctx, "*")
	if govErr != nil {
		return "", fmt.Errorf("failed to fetch vSphere folders. Datacenter: %s, Error: %s", datacenter, govErr.Error())
	}

	prefix := fmt.Sprintf("/%s/vm/", dc)
	for _, fo := range fos {
		inventoryPath := fo.InventoryPath
		// get vm folders: items with path prefix '/{Datacenter}/vm'
		if strings.HasPrefix(inventoryPath, prefix) {
			folderName := strings.TrimPrefix(inventoryPath, prefix)
			if fo.Reference().Value == id {
				return folderName, nil
			}
		}
	}

	return "", fmt.Errorf("unable to find folder with id: %s", id)
}

// CreateVMFolders creates one or more vCenter VM folder(s)
func (v *VCenterDriver) CreateVMFolders(ctx context.Context, datacenter string, folders []string) error {
	finder, _, err := v.GetFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return err
	}

	for _, folder := range folders {
		f, err := v.GetFolder(ctx, finder, folder)
		if err != nil {
			if strings.HasSuffix(err.Error(), "not found") {
				v.log.V(1).Info("folder does not exist; will create it", "path", folder)
			} else {
				return errors.Wrap(err, fmt.Sprintf("failed to check if folder %s exists", folder))
			}
		}
		if f != nil {
			v.log.V(1).Info("folder already exists; skipping", "path", folder)
			continue
		}

		dir := path.Dir(folder)
		name := path.Base(folder)

		if dir == "" {
			dir = "/"
		}

		folder, err := finder.Folder(ctx, dir)
		if err != nil {
			return fmt.Errorf("failed to fetch folder from directory %s: %w", dir, err)
		}

		if _, err := folder.CreateFolder(ctx, name); err != nil {
			return fmt.Errorf("failed to create folder %s: %w", name, err)
		}
	}

	return nil
}
