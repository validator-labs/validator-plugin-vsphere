package vsphere

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/exp/slices"

	"github.com/validator-labs/validator-plugin-vsphere/api/v1alpha1"
	"github.com/validator-labs/validator-plugin-vsphere/api/vcenter"
	"github.com/validator-labs/validator-plugin-vsphere/api/vcenter/entity"
)

// CurrentUser returns the username of the user the vCenter driver is currently authenticated as.
func (v *VCenterDriver) CurrentUser(ctx context.Context) (string, error) {
	session, err := v.Client.SessionManager.UserSession(ctx)
	if err != nil {
		return "", err
	}
	return session.UserName, nil
}

// CurrentDomains returns the current domains for the user the vCenter driver is currently authenticated as.
func (v *VCenterDriver) CurrentDomains(ctx context.Context) ([]string, error) {
	pc := v.Client.PropertyCollector()
	var ud mo.UserDirectory
	if err := pc.RetrieveOne(ctx, v.Client.ServiceContent.UserDirectory.Reference(), nil, &ud); err != nil {
		return nil, err
	}
	if !slices.Contains(ud.DomainList, vcenter.DefaultDomain) {
		ud.DomainList = append(ud.DomainList, vcenter.DefaultDomain)
	}
	if os.Getenv("IS_TEST") == "true" {
		// required for vcsim because its permission principals aren't formatted as DOMAIN\principal
		ud.DomainList = append(ud.DomainList, "")
	}
	v.log.V(1).Info("Retrieved current domains", "domains", ud.DomainList)
	return ud.DomainList, nil
}

// ValidateUserPrivilegeOnEntities validates the user's privileges and permissions on a specific entity.
func (v *VCenterDriver) ValidateUserPrivilegeOnEntities(ctx context.Context, authManager *object.AuthorizationManager, datacenter, username string, finder *find.Finder, rule v1alpha1.PrivilegeValidationRule) ([]string, error) {
	failures := make([]string, 0)

	// Fetch the managed object reference associated with the rule's entity
	objRefPtr, err := v.getObjRef(ctx, datacenter, finder, rule)
	if err != nil {
		return nil, err
	}
	objRef := *objRefPtr

	// List active user's privileges on the entity
	privilegeResults, err := authManager.FetchUserPrivilegeOnEntities(ctx,
		[]types.ManagedObjectReference{objRef},
		username,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to fetch privileges on %s %s for user %s: %w",
			rule.EntityType, rule.EntityName, username, err,
		)
	}

	// Ensure that the user has all required privileges on the entity
	privilegesMap := make(map[string]bool)
	for _, result := range privilegeResults {
		for _, privilege := range result.Privileges {
			privilegesMap[privilege] = true
		}
	}
	for _, privilege := range rule.Privileges {
		if _, ok := privilegesMap[privilege]; !ok {
			failure := fmt.Sprintf(
				"user: %s does not have privilege: %s on entity type: %s",
				username, privilege, rule.EntityType,
			)
			if rule.EntityName != "" {
				failure = fmt.Sprintf("%s with name: %s", failure, rule.EntityName)
			}
			failures = append(failures, failure)
		}
	}

	if rule.Propagation.Enabled {
		// Determine whether the privileges were granted to the user via a permission with propagation enabled
		permissionPropagated, err := v.getPermissionPropagation(ctx, authManager, username, rule, objRef)
		if err != nil {
			return nil, err
		}
		if rule.Propagation.Propagated && !permissionPropagated {
			failure := fmt.Sprintf(
				"propagation is not enabled on the permission that grants privileges to %s on %s",
				username, rule.EntityType,
			)
			if rule.EntityName != "" {
				failure = fmt.Sprintf("%s with name: %s", failure, rule.EntityName)
			}
			failures = append(failures, failure)
		}
	}

	return failures, nil
}

func (v *VCenterDriver) getObjRef(ctx context.Context, datacenter string, finder *find.Finder, rule v1alpha1.PrivilegeValidationRule) (*types.ManagedObjectReference, error) {
	var objRef types.ManagedObjectReference

	switch e := entity.Map[rule.EntityType]; e {
	case entity.Cluster:
		cluster, err := v.GetCluster(ctx, finder, datacenter, rule.EntityName)
		if err != nil {
			return nil, err
		}
		objRef = cluster.Common.Reference()
	case entity.Datacenter:
		datacenter, err := v.GetDatacenter(ctx, finder, rule.EntityName)
		if err != nil {
			return nil, err
		}
		objRef = datacenter.Common.Reference()
	case entity.Datastore:
		datastore, err := v.GetDatastore(ctx, finder, rule.EntityName)
		if err != nil {
			return nil, err
		}
		objRef = datastore.Common.Reference()
	case entity.DistributedVirtualPortgroup:
		dvp, err := v.GetDistributedVirtualPortgroup(ctx, finder, rule.EntityName)
		if err != nil {
			return nil, err
		}
		objRef = dvp.Common.Reference()
	case entity.DistributedVirtualSwitch:
		dvs, err := v.GetDistributedVirtualSwitch(ctx, finder, rule.EntityName)
		if err != nil {
			return nil, err
		}
		objRef = dvs.Common.Reference()
	case entity.Folder:
		folder, err := v.GetFolder(ctx, datacenter, rule.EntityName)
		if err != nil {
			return nil, err
		}
		objRef = folder.Common.Reference()
	case entity.Host:
		host, err := v.GetHost(ctx, finder, datacenter, rule.ClusterName, rule.EntityName)
		if err != nil {
			return nil, err
		}
		objRef = host.Common.Reference()
	case entity.Network:
		network, err := v.GetNetwork(ctx, finder, rule.EntityName)
		if err != nil {
			return nil, err
		}
		objRef = network.Common.Reference()
	case entity.ResourcePool:
		resourcePool, err := v.GetResourcePool(ctx, finder, datacenter, rule.ClusterName, rule.EntityName)
		if err != nil {
			return nil, err
		}
		objRef = resourcePool.Common.Reference()
	case entity.VirtualApp:
		vApp, err := v.GetVApp(ctx, finder, rule.EntityName)
		if err != nil {
			return nil, err
		}
		objRef = vApp.Common.Reference()
	case entity.VCenterRoot:
		objRef = v.Client.Client.ServiceContent.RootFolder
	case entity.VirtualMachine:
		vm, err := v.GetVM(ctx, finder, rule.EntityName)
		if err != nil {
			return nil, err
		}
		objRef = vm.Common.Reference()
	default:
		return nil, fmt.Errorf("unsupported entity type: %s", rule.EntityType)
	}

	return &objRef, nil
}

func (v *VCenterDriver) getPermissionPropagation(ctx context.Context, authManager *object.AuthorizationManager, username string, rule v1alpha1.PrivilegeValidationRule, objRef types.ManagedObjectReference) (bool, error) {
	// Retrieve a list of all permissions on the entity
	permissions, err := authManager.RetrieveEntityPermissions(ctx, objRef, true)
	if err != nil {
		return false, fmt.Errorf(
			"failed to fetch privileges on %s %s: %w",
			rule.EntityType, rule.EntityName, err,
		)
	}
	v.log.V(1).Info("Retrieved entity permissions", "entityType", rule.EntityType, "entityName", rule.EntityName, "permissions", permissions)

	// Retrieve a list of all the domains associated with the user
	currentDomains, err := v.CurrentDomains(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to fetch current vCenter domains: %w", err)
	}

	// Build a map of group principals associated with the user.
	// Ensure that each group principal is formatted as DOMAIN\group-name,
	// and that DOMAIN is listed in the active user's domains.
	groupPrincipals := make([]string, 0)
	groupPrincipalsMap := make(map[string]bool, 0)
	for _, g := range rule.Propagation.GroupPrincipals {
		var gpOk bool
		for _, d := range currentDomains {
			if strings.HasPrefix(g, d) {
				gpOk = true
				break
			}
		}
		if !gpOk {
			return false, fmt.Errorf(
				"group principal %s does not match any vCenter domains %v: %w",
				g, currentDomains, err,
			)
		}
		groupPrincipals = append(groupPrincipals, g)
		groupPrincipalsMap[g] = true
	}

	// Determine if a permission exists on the entity that's scoped to the user's principal.
	// If so, it takes priority. Otherwise we consider all group permissions and if propagation
	// is enabled on any of them, we consider propagation enabled.
	var userScopedPermission *types.Permission
	userPrincipal := userPrincipalFromUsername(username)
	groupPermissions := make([]types.Permission, 0)

	for _, p := range permissions {
		p := p
		if p.Principal == userPrincipal {
			userScopedPermission = &p
			break
		}
		if _, ok := groupPrincipalsMap[p.Principal]; ok {
			groupPermissions = append(groupPermissions, p)
		}
	}
	if userScopedPermission == nil && len(groupPermissions) == 0 {
		return false, fmt.Errorf(
			"no permissions found on %s %s associated with the user principal %s or group principals %v",
			rule.EntityType, rule.EntityName, username, groupPrincipals,
		)
	}

	// Determine whether the permission on the entity relevant to the active user has propagation enabled
	var permissionPropagated bool
	if userScopedPermission != nil {
		permissionPropagated = userScopedPermission.Propagate
	} else {
		for _, gp := range groupPermissions {
			if gp.Propagate {
				permissionPropagated = true
				break
			}
		}
	}

	return permissionPropagated, nil
}

// given admin@vsphere.local, returns admin
func userPrincipalFromUsername(username string) string {
	return strings.Split(username, "@")[0]
}
