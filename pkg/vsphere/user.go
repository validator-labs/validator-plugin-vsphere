package vsphere

import (
	"context"
	"fmt"
	"strings"

	"github.com/validator-labs/validator-plugin-vsphere/api/v1alpha1"
	"github.com/validator-labs/validator-plugin-vsphere/api/vcenter"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

// GetCurrentVmwareUser returns the user name the CloudDriver is currently authenticated with
func (v *CloudDriver) GetCurrentVmwareUser(ctx context.Context) (string, error) {
	userSession, err := v.Client.SessionManager.UserSession(ctx)
	if err != nil {
		return "", err
	}

	return userSession.UserName, nil
}

// ValidateUserPrivilegeOnEntities validates the user privileges on the entities
func (v *CloudDriver) ValidateUserPrivilegeOnEntities(ctx context.Context, authManager *object.AuthorizationManager, datacenter string, finder *find.Finder, rule v1alpha1.PrivilegeValidationRule) (isValid bool, failures []string, err error) {

	var moID types.ManagedObjectReference

	// TODO: add network, datacenter, datastore, vCenter root, VDS

	switch rule.EntityType {
	case vcenter.Cluster:
		var cluster *object.ClusterComputeResource
		_, cluster, err = v.GetClusterIfExists(ctx, finder, datacenter, rule.EntityName)
		if err != nil {
			return false, failures, err
		}
		moID = cluster.Reference()
	case vcenter.Folder:
		var folder *object.Folder
		_, folder, err = v.GetFolderIfExists(ctx, finder, rule.EntityName)
		if err != nil {
			return false, failures, err
		}
		moID = folder.Reference()
	case vcenter.Host:
		var host *object.HostSystem
		_, host, err = v.GetHostIfExists(ctx, finder, datacenter, rule.ClusterName, rule.EntityName)
		if err != nil {
			return false, failures, err
		}
		moID = host.Reference()
	case vcenter.ResourcePool:
		var resourcePool *object.ResourcePool
		_, resourcePool, err = v.GetResourcePoolIfExists(ctx, finder, datacenter, rule.ClusterName, rule.EntityName)
		if err != nil {
			return false, failures, err
		}
		moID = resourcePool.Reference()
	case vcenter.VApp:
		var vapp *object.VirtualApp
		_, vapp, err = v.GetVAppIfExists(ctx, finder, rule.EntityName)
		if err != nil {
			return false, failures, err
		}
		moID = vapp.Reference()
	case vcenter.VM:
		var vm *object.VirtualMachine
		_, vm, err = v.GetVMIfExists(ctx, finder, rule.EntityName)
		if err != nil {
			return false, failures, err
		}
		moID = vm.Reference()
	default:
		return false, failures, fmt.Errorf("unsupported entity type: %s", rule.EntityType)
	}

	userPrincipal := getUserPrincipalFromUsername(rule.Username)
	privilegeResult, err := authManager.FetchUserPrivilegeOnEntities(ctx, []types.ManagedObjectReference{moID}, userPrincipal)
	if err != nil {
		return false, failures, err
	}

	privilegesMap := make(map[string]bool)
	for _, result := range privilegeResult {
		for _, privilege := range result.Privileges {
			privilegesMap[privilege] = true
		}
	}

	for _, privilege := range rule.Privileges {
		if _, ok := privilegesMap[privilege]; !ok {
			err = fmt.Errorf("some entity privileges were not found for user: %s", rule.Username)
			failures = append(failures, fmt.Sprintf(
				"user: %s does not have privilege: %s on entity type: %s with name: %s",
				rule.Username, privilege, rule.EntityType, rule.EntityName,
			))
		}
	}

	if len(failures) == 0 {
		isValid = true
	}

	return isValid, failures, err
}

func getUserPrincipalFromUsername(username string) string {
	splitStr := strings.Split(username, "@")
	return fmt.Sprintf("%s\\%s", strings.ToUpper(splitStr[1]), splitStr[0])
}
