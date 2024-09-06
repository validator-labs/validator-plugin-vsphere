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

// CurrentUser returns the username of the user the vCenter driver is currently authenticated as
func (v *VCenterDriver) CurrentUser(ctx context.Context) (string, error) {
	session, err := v.Client.SessionManager.UserSession(ctx)
	if err != nil {
		return "", err
	}
	return session.UserName, nil
}

// ValidateUserPrivilegeOnEntities validates the user privileges on the entities
func (v *VCenterDriver) ValidateUserPrivilegeOnEntities(ctx context.Context, authManager *object.AuthorizationManager, datacenter string, finder *find.Finder, rule v1alpha1.PrivilegeValidationRule) ([]string, error) {

	var obj object.Common

	// TODO: add network, datacenter, datastore, vCenter root, VDS

	switch rule.EntityType {
	case vcenter.Cluster:
		_, cluster, err := v.GetClusterIfExists(ctx, finder, datacenter, rule.EntityName)
		if err != nil {
			return nil, err
		}
		obj = cluster.Common
	case vcenter.Folder:
		_, folder, err := v.GetFolderIfExists(ctx, finder, rule.EntityName)
		if err != nil {
			return nil, err
		}
		obj = folder.Common
	case vcenter.Host:
		_, host, err := v.GetHostIfExists(ctx, finder, datacenter, rule.ClusterName, rule.EntityName)
		if err != nil {
			return nil, err
		}
		obj = host.Common
	case vcenter.ResourcePool:
		_, resourcePool, err := v.GetResourcePoolIfExists(ctx, finder, datacenter, rule.ClusterName, rule.EntityName)
		if err != nil {
			return nil, err
		}
		obj = resourcePool.Common
	case vcenter.VApp:
		_, vapp, err := v.GetVAppIfExists(ctx, finder, rule.EntityName)
		if err != nil {
			return nil, err
		}
		obj = vapp.Common
	case vcenter.VM:
		_, vm, err := v.GetVMIfExists(ctx, finder, rule.EntityName)
		if err != nil {
			return nil, err
		}
		obj = vm.Common
	default:
		return nil, fmt.Errorf("unsupported entity type: %s", rule.EntityType)
	}

	privilegeResult, err := authManager.FetchUserPrivilegeOnEntities(ctx,
		[]types.ManagedObjectReference{obj.Reference()},
		getUserPrincipalFromUsername(rule.Username),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to fetch privileges on %s %s for user %s: %w",
			rule.EntityType, rule.EntityName, rule.Username, err,
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
			failures = append(failures, fmt.Sprintf(
				"user: %s does not have privilege: %s on entity type: %s with name: %s",
				rule.Username, privilege, rule.EntityType, rule.EntityName,
			))
		}
	}

	return failures, nil
}

func getUserPrincipalFromUsername(username string) string {
	splitStr := strings.Split(username, "@")
	return fmt.Sprintf("%s\\%s", strings.ToUpper(splitStr[1]), splitStr[0])
}
