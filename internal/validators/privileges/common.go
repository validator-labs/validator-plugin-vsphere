package privileges

import (
	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/object"

	"github.com/validator-labs/validator-plugin-vsphere/pkg/vsphere"
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
