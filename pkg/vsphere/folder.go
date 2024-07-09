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

// FolderExists checks if a folder exists in the vSphere inventory
func (v *CloudDriver) FolderExists(ctx context.Context, finder *find.Finder, folderName string) (bool, error) {
	if _, err := finder.Folder(ctx, folderName); err != nil {
		return false, nil
	}
	return true, nil
}

// GetFolderIfExists returns the folder if it exists
func (v *CloudDriver) GetFolderIfExists(ctx context.Context, finder *find.Finder, folderName string) (bool, *object.Folder, error) {
	folder, err := finder.Folder(ctx, folderName)
	if err != nil {
		return false, nil, err
	}
	return true, folder, nil
}

// GetVSphereVMFolders returns a list of vSphere VM folders
func (v *CloudDriver) GetVSphereVMFolders(ctx context.Context, datacenter string) ([]string, error) {
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
		//get vm folders, items with path prefix '/{Datacenter}/vm'
		if strings.HasPrefix(inventoryPath, prefix) {
			folder := strings.TrimPrefix(inventoryPath, prefix)
			//skip spectro folders & sub-folders
			if !strings.HasPrefix(folder, "spc-") &&
				!strings.Contains(folder, "/spc-") {
				folders = append(folders, folder)
			}
		}
	}

	sort.Strings(folders)
	return folders, nil
}

// GetFolderNameByID returns the folder name by ID
func (v *CloudDriver) GetFolderNameByID(ctx context.Context, datacenter, id string) (string, error) {
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
		//get vm folders, items with path prefix '/{Datacenter}/vm'
		if strings.HasPrefix(inventoryPath, prefix) {
			folderName := strings.TrimPrefix(inventoryPath, prefix)
			//skip spectro folders & sub-folders
			if !strings.HasPrefix(folderName, "spc-") && !strings.Contains(folderName, "/spc-") {
				if fo.Reference().Value == id {
					return folderName, nil
				}
			}
		}
	}

	return "", fmt.Errorf("unable to find folder with id: %s", id)
}

// CreateVSphereVMFolder creates a vSphere VM folder
func (v *CloudDriver) CreateVSphereVMFolder(ctx context.Context, datacenter string, folders []string) error {
	finder, _, err := v.GetFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return err
	}

	for _, folder := range folders {
		folderExists, _, err := v.GetFolderIfExists(ctx, finder, folder)
		if err != nil {
			if strings.HasSuffix(err.Error(), "not found") {
				v.log.V(1).Info("folder does not exist; will create it", "path", folder)
			} else {
				return errors.Wrap(err, fmt.Sprintf("failed to check if folder %s exists", folder))
			}
		}
		if folderExists {
			continue
		}

		dir := path.Dir(folder)
		name := path.Base(folder)

		if dir == "" {
			dir = "/"
		}

		folder, err := finder.Folder(ctx, dir)
		if err != nil {
			return fmt.Errorf("error fetching folder from directory %s: %w", dir, err)
		}

		if _, err := folder.CreateFolder(ctx, name); err != nil {
			return fmt.Errorf("error creating folder %s: %w", name, err)
		}
	}

	return nil
}
