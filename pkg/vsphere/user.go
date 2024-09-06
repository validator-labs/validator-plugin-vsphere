package vsphere

import (
	"context"
	"fmt"
	"strings"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/validator-labs/validator-plugin-vsphere/api/v1alpha1"
	"github.com/validator-labs/validator-plugin-vsphere/api/vcenter/entity"
)

// CurrentUser returns the username of the user the vCenter driver is currently authenticated as
func (v *VCenterDriver) CurrentUser(ctx context.Context) (string, error) {
	session, err := v.Client.SessionManager.UserSession(ctx)
	if err != nil {
		return "", err
	}
	return session.UserName, nil
}

// ValidateUserPrivilegeOnEntities validates the user privileges on the entities
func (v *VCenterDriver) ValidateUserPrivilegeOnEntities(ctx context.Context, authManager *object.AuthorizationManager, datacenter, username string, finder *find.Finder, rule v1alpha1.PrivilegeValidationRule) ([]string, error) {

	var objRef types.ManagedObjectReference

	switch rule.EntityType {
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
	case entity.DistributedVirtualSwitch:
		dvs, err := v.GetDistributedVirtualSwitch(ctx, finder, rule.EntityName)
		if err != nil {
			return nil, err
		}
		objRef = dvs.Common.Reference()
	case entity.Folder:
		folder, err := v.GetFolder(ctx, finder, rule.EntityName)
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

	privilegeResult, err := authManager.FetchUserPrivilegeOnEntities(ctx,
		[]types.ManagedObjectReference{objRef},
		getUserPrincipalFromUsername(username),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to fetch privileges on %s %s for user %s: %w",
			rule.EntityType, rule.EntityName, username, err,
		)
	}

	failures := make([]string, 0)
	privilegesMap := make(map[string]bool)
	for _, result := range privilegeResult {
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

	return failures, nil
}

func getUserPrincipalFromUsername(username string) string {
	splitStr := strings.Split(username, "@")
	return fmt.Sprintf("%s\\%s", strings.ToUpper(splitStr[1]), splitStr[0])
}
