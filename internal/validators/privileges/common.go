package privileges

import (
	"github.com/go-logr/logr"
	"github.com/spectrocloud-labs/validator-plugin-vsphere/pkg/vsphere"
	"github.com/vmware/govmomi/object"
)

type PrivilegeValidationService struct {
	log         logr.Logger
	driver      *vsphere.VSphereCloudDriver
	datacenter  string
	authManager *object.AuthorizationManager
	userName    string
}

func NewPrivilegeValidationService(log logr.Logger, driver *vsphere.VSphereCloudDriver, datacenter string, authManager *object.AuthorizationManager, userName string) *PrivilegeValidationService {
	return &PrivilegeValidationService{
		log:         log,
		driver:      driver,
		datacenter:  datacenter,
		authManager: authManager,
		userName:    userName,
	}
}
