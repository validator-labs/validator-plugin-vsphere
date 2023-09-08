package roleprivilege

import (
	"github.com/go-logr/logr"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/api/v1alpha1"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/internal/vcsim"
	"testing"
)

func TestRolePrivilegeValidationService_ReconcileRolePrivilegesRule(t *testing.T) {
	vcSim := vcsim.NewVCSim("admin@vsphere.local")

	vcSim.Start()
	privileges := make(map[string]bool)

	privileges["Cns.Searchable"] = true
	var log logr.Logger
	rule := v1alpha1.GenericRolePrivilegeValidationRule{
		Name:        "Cns.Searchable",
		Description: "",
		IsEnabled:   true,
		Severity:    "",
		RuleType:    "VMwareRolePrivilege",
		Expressions: []string{`"Cns.Searchable" IN (vmware_user_privileges)`},
	}

	validationService := NewRolePrivilegeValidationService(log, vcSim.Driver, nil)
	t.Log(rule, privileges, validationService)
	_, err := validationService.ReconcileRolePrivilegesRule(rule, privileges)
	if err != nil {
		t.Fatal(err)
	}

}
