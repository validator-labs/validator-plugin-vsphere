package vsphere

import (
	"context"
	"fmt"
	"strings"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

func (v *VSphereCloudDriver) GetCurrentVmwareUser(ctx context.Context) (string, error) {
	userSession, err := v.Client.SessionManager.UserSession(ctx)
	if err != nil {
		return "", err
	}

	return userSession.UserName, nil
}

func (v *VSphereCloudDriver) ValidateUserPrivilegeOnEntities(ctx context.Context, authManager *object.AuthorizationManager, datacenter string, finder *find.Finder, entityName, entityType string, privileges []string, userName, clusterName string) (isValid bool, failures []string, err error) {
	var folder *object.Folder
	var cluster *object.ClusterComputeResource
	var host *object.HostSystem
	var vapp *object.VirtualApp
	var resourcePool *object.ResourcePool
	var vm *object.VirtualMachine

	var moID types.ManagedObjectReference

	switch entityType {
	case "folder":
		_, folder, err = v.GetFolderIfExists(ctx, finder, datacenter, entityName)
		if err != nil {
			return false, failures, err
		}
		moID = folder.Reference()
	case "resourcepool":
		_, resourcePool, err = v.GetResourcePoolIfExists(ctx, finder, datacenter, clusterName, entityName)
		if err != nil {
			return false, failures, err
		}
		moID = resourcePool.Reference()
	case "vapp":
		_, vapp, err = v.GetVAppIfExists(ctx, finder, datacenter, entityName)
		if err != nil {
			return false, failures, err
		}
		moID = vapp.Reference()
	case "vm":
		_, vm, err = v.GetVMIfExists(ctx, finder, datacenter, clusterName, entityName)
		if err != nil {
			return false, failures, err
		}
		moID = vm.Reference()
	case "host":
		_, host, err = v.GetHostIfExists(ctx, finder, datacenter, clusterName, entityName)
		if err != nil {
			return false, failures, err
		}
		moID = host.Reference()
	case "cluster":
		_, cluster, err = v.GetClusterIfExists(ctx, finder, datacenter, entityName)
		if err != nil {
			return false, failures, err
		}
		moID = cluster.Reference()
	}

	userPrincipal := getUserPrincipalFromUsername(userName)
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

	for _, privilege := range privileges {
		if _, ok := privilegesMap[privilege]; !ok {
			err = fmt.Errorf("some entity privileges were not found for user: %s", userName)
			failures = append(failures, fmt.Sprintf("user: %s does not have privilege: %s on entity type: %s with name: %s", userName, privilege, entityType, entityName))
		}
	}

	if len(failures) == 0 {
		isValid = true
	}

	return isValid, failures, nil
}

func GetVmwareUserPrivileges(ctx context.Context, userPrincipal string, groupPrincipals []string, authManager *object.AuthorizationManager) (map[string]bool, error) {
	groupPrincipalMap := make(map[string]bool)
	for _, principal := range groupPrincipals {
		groupPrincipalMap[principal] = true
	}

	// Get the current user's roles
	authRoles, err := authManager.RoleList(ctx)
	if err != nil {
		return nil, err
	}

	// create a map to store privileges for current user
	privileges := make(map[string]bool)

	// Print the roles
	for _, authRole := range authRoles {
		// print permissions for every role
		permissions, err := authManager.RetrieveRolePermissions(ctx, authRole.RoleId)
		if err != nil {
			return nil, err
		}
		for _, perm := range permissions {
			if perm.Principal == userPrincipal || groupPrincipalMap[perm.Principal] {
				for _, priv := range authRole.Privilege {
					privileges[priv] = true
				}
			}
		}
	}
	return privileges, nil
}

func getUserPrincipalFromUsername(username string) string {
	splitStr := strings.Split(username, "@")
	return fmt.Sprintf("%s\\%s", strings.ToUpper(splitStr[1]), splitStr[0])
}
