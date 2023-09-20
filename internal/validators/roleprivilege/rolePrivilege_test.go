package roleprivilege

import (
	"github.com/go-logr/logr"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/api/v1alpha1"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/internal/vcsim"
	"github.com/vmware/govmomi/object"
	"testing"
)

func TestRolePrivilegeValidationService_ReconcileRolePrivilegesRule(t *testing.T) {
	userName := "admin@vsphere.local"
	vcSim := vcsim.NewVCSim(userName)

	vcSim.Start()
	privileges := make(map[string]bool)

	privileges["Cns.Searchable"] = true
	var log logr.Logger
	rule := v1alpha1.GenericRolePrivilegeValidationRule{
		Name:        "Cns.Searchable",
		IsEnabled:   true,
		RuleType:    "VMwareRolePrivilege",
		Expressions: []string{`"Cns.Searchable" IN (vmware_user_privileges)`},
	}

	authManager := object.NewAuthorizationManager(vcSim.Driver.Client.Client)
	if authManager == nil {
		t.Fatal("Error in creating auth manager")
	}

	validationService := NewRolePrivilegeValidationService(log, vcSim.Driver, "DC0", authManager, userName)
	t.Log(rule, privileges, validationService)
	_, err := validationService.ReconcileRolePrivilegesRule(rule, privileges)
	if err != nil {
		t.Fatal(err)
	}

}
