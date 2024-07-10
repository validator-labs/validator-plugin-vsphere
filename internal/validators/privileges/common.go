// Package privileges handles privilege validation rule reconciliation.
package privileges

import (
	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/object"

	"github.com/validator-labs/validator-plugin-vsphere/pkg/vsphere"
)

// PrivilegeValidationService is a service that validates user privileges
type PrivilegeValidationService struct {
	log         logr.Logger
	driver      *vsphere.CloudDriver
	datacenter  string
	authManager *object.AuthorizationManager
	userName    string
}

// NewPrivilegeValidationService creates a new PrivilegeValidationService
func NewPrivilegeValidationService(log logr.Logger, driver *vsphere.CloudDriver, datacenter string, authManager *object.AuthorizationManager, userName string) *PrivilegeValidationService {
	return &PrivilegeValidationService{
		log:         log,
		driver:      driver,
		datacenter:  datacenter,
		authManager: authManager,
		userName:    userName,
	}
}
