package validate

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/go-logr/logr"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/validator-labs/validator-plugin-vsphere/api/v1alpha1"
	"github.com/validator-labs/validator-plugin-vsphere/api/vcenter/entity"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/vcsim"
)

func TestValidate(t *testing.T) {
	vcSim := vcsim.NewVCSim("admin@vsphere.local", 8449, logr.Logger{})
	vcSim.Start()
	defer vcSim.Shutdown()

	opts := vcSim.Options

	tests := []struct {
		name     string
		spec     v1alpha1.VsphereValidatorSpec
		expected string
	}{
		{
			name: "Cluster_Pass",
			spec: v1alpha1.VsphereValidatorSpec{
				Auth: v1alpha1.VsphereAuth{
					Account: &vcSim.Account,
				},
				Datacenter: opts.Datacenter,
				PrivilegeValidationRules: testRules([]privilegeRuleInput{
					{
						EntityType: entity.Cluster,
						EntityName: opts.Cluster,
						Privileges: []string{"Alarm.Acknowledge"},
					},
				}),
			},
			expected: `{"ValidationRuleResults":[{"Condition":{"validationType":"vsphere-privileges","validationRule":"validation-cluster-DC0_C0","message":"All required vsphere-privileges permissions were found for account: admin@vsphere.local","status":"True","lastValidationTime":null},"State":"Succeeded"}],"ValidationRuleErrors":[null]}`,
		},
		{
			name: "Cluster_Fail",
			spec: v1alpha1.VsphereValidatorSpec{
				Auth: v1alpha1.VsphereAuth{
					Account: &vcSim.Account,
				},
				Datacenter: opts.Datacenter,
				PrivilegeValidationRules: testRules([]privilegeRuleInput{
					{
						EntityType: entity.Cluster,
						EntityName: opts.Cluster,
						Privileges: []string{"Nonexistent"},
					},
				}),
			},
			expected: `{"ValidationRuleResults":[{"Condition":{"validationType":"vsphere-privileges","validationRule":"validation-cluster-DC0_C0","message":"One or more required privileges was not found, or a condition was not met for account: admin@vsphere.local","failures":["user: admin@vsphere.local does not have privilege: Nonexistent on entity type: cluster with name: DC0_C0"],"status":"False","lastValidationTime":null},"State":"Failed"}],"ValidationRuleErrors":[null]}`,
		},
		{
			name: "Root_Pass",
			spec: v1alpha1.VsphereValidatorSpec{
				Auth: v1alpha1.VsphereAuth{
					Account: &vcSim.Account,
				},
				Datacenter: opts.Datacenter,
				PrivilegeValidationRules: testRules([]privilegeRuleInput{
					{
						EntityType: entity.VCenterRoot,
						EntityName: "",
						Privileges: []string{"Alarm.Acknowledge"},
					},
				}),
			},
			expected: `{"ValidationRuleResults":[{"Condition":{"validationType":"vsphere-privileges","validationRule":"validation-root","message":"All required vsphere-privileges permissions were found for account: admin@vsphere.local","status":"True","lastValidationTime":null},"State":"Succeeded"}],"ValidationRuleErrors":[null]}`,
		},
		{
			name: "Datastore_Pass",
			spec: v1alpha1.VsphereValidatorSpec{
				Auth: v1alpha1.VsphereAuth{
					Account: &vcSim.Account,
				},
				Datacenter: opts.Datacenter,
				PrivilegeValidationRules: testRules([]privilegeRuleInput{
					{
						EntityType: entity.Datastore,
						EntityName: opts.Datastore,
						Privileges: []string{"Alarm.Acknowledge"},
					},
				}),
			},
			expected: `{"ValidationRuleResults":[{"Condition":{"validationType":"vsphere-privileges","validationRule":"validation-datastore-LocalDS_0","message":"All required vsphere-privileges permissions were found for account: admin@vsphere.local","status":"True","lastValidationTime":null},"State":"Succeeded"}],"ValidationRuleErrors":[null]}`,
		},
		{
			name: "Network_Pass",
			spec: v1alpha1.VsphereValidatorSpec{
				Auth: v1alpha1.VsphereAuth{
					Account: &vcSim.Account,
				},
				Datacenter: opts.Datacenter,
				PrivilegeValidationRules: testRules([]privilegeRuleInput{
					{
						EntityType: entity.Network,
						EntityName: opts.Network.Name,
						Privileges: []string{"Alarm.Acknowledge"},
					},
				}),
			},
			expected: `{"ValidationRuleResults":[{"Condition":{"validationType":"vsphere-privileges","validationRule":"validation-network-VM Network","message":"All required vsphere-privileges permissions were found for account: admin@vsphere.local","status":"True","lastValidationTime":null},"State":"Succeeded"}],"ValidationRuleErrors":[null]}`,
		},
		// DistributedVirtualSwitch not yet supported in govmomi
		// {
		// 	name: "DistributedVirtualSwitch_Pass",
		// 	spec: v1alpha1.VsphereValidatorSpec{
		// 		Auth: v1alpha1.VsphereAuth{
		// 			Account: &vcSim.Account,
		// 		},
		// 		Datacenter: opts.Datacenter,
		// 		PrivilegeValidationRules: testRules([]privilegeRuleInput{
		// 			{
		// 				EntityType: entity.DistributedVirtualSwitch,
		// 				EntityName: "dvs",
		// 				Privileges: []string{"Alarm.Acknowledge"},
		// 			},
		// 		}),
		// 	},
		// 	expected: `{"ValidationRuleResults":[{"Condition":{"validationType":"vsphere-privileges","validationRule":"validation-datastore-LocalDS_0","message":"All required vsphere-privileges permissions were found for account: admin@vsphere.local","status":"True","lastValidationTime":null},"State":"Succeeded"}],"ValidationRuleErrors":[null]}`,
		// },
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := Validate(context.TODO(), tc.spec, logr.Logger{})
			for _, r := range result.ValidationRuleResults {
				r.Condition.LastValidationTime = v1.Time{}
			}
			resultJson, _ := json.Marshal(result)
			resultStr := string(resultJson)

			if !reflect.DeepEqual(resultStr, tc.expected) {
				t.Errorf("Validate() got %s != expected %s", resultStr, tc.expected)
			}
		})
	}
}

type privilegeRuleInput struct {
	EntityType entity.Entity
	EntityName string
	Privileges []string
}

func testRules(inputs []privilegeRuleInput) []v1alpha1.PrivilegeValidationRule {
	rules := make([]v1alpha1.PrivilegeValidationRule, 0)
	for i, input := range inputs {
		r := v1alpha1.PrivilegeValidationRule{
			RuleName:   fmt.Sprintf("rule %d", i),
			EntityType: input.EntityType,
			EntityName: input.EntityName,
			Privileges: input.Privileges,
		}
		rules = append(rules, r)
	}
	return rules
}
