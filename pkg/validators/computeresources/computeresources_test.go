package computeresources

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25/mo"
	vtypes "github.com/vmware/govmomi/vim25/types"
	corev1 "k8s.io/api/core/v1"

	vapi "github.com/validator-labs/validator/api/v1alpha1"
	"github.com/validator-labs/validator/pkg/test"
	"github.com/validator-labs/validator/pkg/types"
	"github.com/validator-labs/validator/pkg/util"

	"github.com/validator-labs/validator-plugin-vsphere/api/v1alpha1"
	"github.com/validator-labs/validator-plugin-vsphere/api/vcenter/entity"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/vcsim"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/vsphere"
)

func TestReconcileComputeResourceValidationRule(t *testing.T) {
	var log logr.Logger

	vcSim := vcsim.NewVCSim("admin@vsphere.local", 8447, log)
	vcSim.Start()
	defer vcSim.Shutdown()

	driver, err := vsphere.NewVCenterDriver(vcSim.Account, vcSim.Options.Datacenter, logr.Logger{})
	if err != nil {
		t.Fatal(err)
	}

	finder := find.NewFinder(driver.Client.Client)

	validationService := NewValidationService(log, driver)
	testCases := []struct {
		name           string
		expectedErr    error
		rule           v1alpha1.ComputeResourceRule
		expectedResult types.ValidationRuleResult
	}{
		{
			name: "All Resources available",
			rule: v1alpha1.ComputeResourceRule{
				RuleName:    "Test Resource Validation rule",
				ClusterName: "DC0_C0",
				Scope:       entity.Cluster.String(),
				EntityName:  "DC0_C0",
				NodepoolResourceRequirements: []v1alpha1.NodepoolResourceRequirement{
					{
						Name:          "masterpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "500Mi",
						DiskSpace:     "50Gi",
					},
					{
						Name:          "workerpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "1Gi",
						DiskSpace:     "100Gi",
					},
				},
			},
			expectedResult: types.ValidationRuleResult{Condition: &vapi.ValidationCondition{
				ValidationType: "vsphere-compute-resources",
				ValidationRule: "validation-cluster-dc0-c0",
				Message:        "All required compute resources were satisfied",
				Details:        []string{},
				Failures:       nil,
				Status:         corev1.ConditionTrue,
			},
				State: util.Ptr(vapi.ValidationSucceeded),
			},
		},
		{
			name: "cluster CPU not available",
			rule: v1alpha1.ComputeResourceRule{
				RuleName:    "Test Resource Validation rule",
				ClusterName: "DC0_C0",
				Scope:       entity.Cluster.String(),
				EntityName:  "DC0_C0",
				NodepoolResourceRequirements: []v1alpha1.NodepoolResourceRequirement{
					{
						Name:          "masterpool",
						NumberOfNodes: 1,
						CPU:           "10GHz",
						Memory:        "500Mi",
						DiskSpace:     "50Gi",
					},
					{
						Name:          "workerpool",
						NumberOfNodes: 1,
						CPU:           "10GHz",
						Memory:        "1Gi",
						DiskSpace:     "100Gi",
					},
				},
			},
			expectedResult: types.ValidationRuleResult{Condition: &vapi.ValidationCondition{
				ValidationType: "vsphere-compute-resources",
				ValidationRule: "validation-cluster-dc0-c0",
				Message:        "One or more resource requirements were not satisfied",
				Details:        []string{},
				Failures:       []string{"Not enough resources available. CPU available: false, Memory available: true, Storage available: true"},
				Status:         corev1.ConditionFalse,
			},
				State: util.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name: "cluster Memory not available",
			rule: v1alpha1.ComputeResourceRule{
				RuleName:    "Test Resource Validation rule",
				ClusterName: "DC0_C0",
				Scope:       entity.Cluster.String(),
				EntityName:  "DC0_C0",
				NodepoolResourceRequirements: []v1alpha1.NodepoolResourceRequirement{
					{
						Name:          "masterpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "500Gi",
						DiskSpace:     "50Gi",
					},
					{
						Name:          "workerpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "100Gi",
						DiskSpace:     "100Gi",
					},
				},
			},
			expectedResult: types.ValidationRuleResult{Condition: &vapi.ValidationCondition{
				ValidationType: "vsphere-compute-resources",
				ValidationRule: "validation-cluster-dc0-c0",
				Message:        "One or more resource requirements were not satisfied",
				Details:        []string{},
				Failures:       []string{"Not enough resources available. CPU available: true, Memory available: false, Storage available: true"},
				Status:         corev1.ConditionFalse,
			},
				State: util.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name: "cluster Disk not available",
			rule: v1alpha1.ComputeResourceRule{
				RuleName:    "Test Resource Validation rule",
				ClusterName: "DC0_C0",
				Scope:       entity.Cluster.String(),
				EntityName:  "DC0_C0",
				NodepoolResourceRequirements: []v1alpha1.NodepoolResourceRequirement{
					{
						Name:          "masterpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "500Mi",
						DiskSpace:     "500Ti",
					},
					{
						Name:          "workerpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "1Gi",
						DiskSpace:     "100Ti",
					},
				},
			},
			expectedResult: types.ValidationRuleResult{Condition: &vapi.ValidationCondition{
				ValidationType: "vsphere-compute-resources",
				ValidationRule: "validation-cluster-dc0-c0",
				Message:        "One or more resource requirements were not satisfied",
				Details:        []string{},
				Failures:       []string{"Not enough resources available. CPU available: true, Memory available: true, Storage available: false"},
				Status:         corev1.ConditionFalse,
			},
				State: util.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name: "Host - All Resources available",
			rule: v1alpha1.ComputeResourceRule{
				RuleName:   "Test Host Resource Validation rule",
				Scope:      entity.Host.String(),
				EntityName: "DC0_C0_H0",
				NodepoolResourceRequirements: []v1alpha1.NodepoolResourceRequirement{
					{
						Name:          "masterpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "500Mi",
						DiskSpace:     "50Gi",
					},
					{
						Name:          "workerpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "1Gi",
						DiskSpace:     "100Gi",
					},
				},
			},
			expectedResult: types.ValidationRuleResult{Condition: &vapi.ValidationCondition{
				ValidationType: "vsphere-compute-resources",
				ValidationRule: "validation-esxi-host-dc0-c0-h0",
				Message:        "All required compute resources were satisfied",
				Details:        []string{},
				Failures:       nil,
				Status:         corev1.ConditionTrue,
			},
				State: util.Ptr(vapi.ValidationSucceeded),
			},
		},
		{
			name: "Host CPU not available",
			rule: v1alpha1.ComputeResourceRule{
				RuleName:   "Test Host Resource Validation rule",
				Scope:      entity.Host.String(),
				EntityName: "DC0_C0_H0",
				NodepoolResourceRequirements: []v1alpha1.NodepoolResourceRequirement{
					{
						Name:          "masterpool",
						NumberOfNodes: 1,
						CPU:           "10GHz",
						Memory:        "500Mi",
						DiskSpace:     "50Gi",
					},
					{
						Name:          "workerpool",
						NumberOfNodes: 1,
						CPU:           "10GHz",
						Memory:        "1Gi",
						DiskSpace:     "100Gi",
					},
				},
			},
			expectedResult: types.ValidationRuleResult{Condition: &vapi.ValidationCondition{
				ValidationType: "vsphere-compute-resources",
				ValidationRule: "validation-esxi-host-dc0-c0-h0",
				Message:        "One or more resource requirements were not satisfied",
				Details:        []string{},
				Failures:       []string{"Not enough resources available. CPU available: false, Memory available: true, Storage available: true"},
				Status:         corev1.ConditionFalse,
			},
				State: util.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name: "Host Memory not available",
			rule: v1alpha1.ComputeResourceRule{
				RuleName:   "Test Host Resource Validation rule",
				Scope:      entity.Host.String(),
				EntityName: "DC0_C0_H0",
				NodepoolResourceRequirements: []v1alpha1.NodepoolResourceRequirement{
					{
						Name:          "masterpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "500Gi",
						DiskSpace:     "50Gi",
					},
					{
						Name:          "workerpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "100Gi",
						DiskSpace:     "100Gi",
					},
				},
			},
			expectedResult: types.ValidationRuleResult{Condition: &vapi.ValidationCondition{
				ValidationType: "vsphere-compute-resources",
				ValidationRule: "validation-esxi-host-dc0-c0-h0",
				Message:        "One or more resource requirements were not satisfied",
				Details:        []string{},
				Failures:       []string{"Not enough resources available. CPU available: true, Memory available: false, Storage available: true"},
				Status:         corev1.ConditionFalse,
			},
				State: util.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name: "Host Disk not available",
			rule: v1alpha1.ComputeResourceRule{
				RuleName:   "Test Host Resource Validation rule",
				Scope:      entity.Host.String(),
				EntityName: "DC0_C0_H0",
				NodepoolResourceRequirements: []v1alpha1.NodepoolResourceRequirement{
					{
						Name:          "masterpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "500Mi",
						DiskSpace:     "500Ti",
					},
					{
						Name:          "workerpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "1Gi",
						DiskSpace:     "100Ti",
					},
				},
			},
			expectedResult: types.ValidationRuleResult{Condition: &vapi.ValidationCondition{
				ValidationType: "vsphere-compute-resources",
				ValidationRule: "validation-esxi-host-dc0-c0-h0",
				Message:        "One or more resource requirements were not satisfied",
				Details:        []string{},
				Failures:       []string{"Not enough resources available. CPU available: true, Memory available: true, Storage available: false"},
				Status:         corev1.ConditionFalse,
			},
				State: util.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name: "Resourcepool - All Resources available",
			rule: v1alpha1.ComputeResourceRule{
				RuleName:    "Test Host Resource Validation rule",
				Scope:       entity.ResourcePool.String(),
				ClusterName: "DC0_C0",
				EntityName:  "DC0_C0_RP0",
				NodepoolResourceRequirements: []v1alpha1.NodepoolResourceRequirement{
					{
						Name:          "masterpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "500Mi",
						DiskSpace:     "50Gi",
					},
					{
						Name:          "workerpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "1Gi",
						DiskSpace:     "100Gi",
					},
				},
			},
			expectedResult: types.ValidationRuleResult{Condition: &vapi.ValidationCondition{
				ValidationType: "vsphere-compute-resources",
				ValidationRule: "validation-resource-pool-dc0-c0-rp0",
				Message:        "All required compute resources were satisfied",
				Details:        []string{},
				Failures:       nil,
				Status:         corev1.ConditionTrue,
			},
				State: util.Ptr(vapi.ValidationSucceeded),
			},
		},
		{
			name: "Resourcepool CPU not available",
			rule: v1alpha1.ComputeResourceRule{
				RuleName:    "Test Host Resource Validation rule",
				Scope:       entity.ResourcePool.String(),
				ClusterName: "DC0_C0",
				EntityName:  "DC0_C0_RP0",
				NodepoolResourceRequirements: []v1alpha1.NodepoolResourceRequirement{
					{
						Name:          "masterpool",
						NumberOfNodes: 1,
						CPU:           "10000GHz",
						Memory:        "500Mi",
						DiskSpace:     "50Gi",
					},
					{
						Name:          "workerpool",
						NumberOfNodes: 1,
						CPU:           "10GHz",
						Memory:        "1Gi",
						DiskSpace:     "100Gi",
					},
				},
			},
			expectedResult: types.ValidationRuleResult{Condition: &vapi.ValidationCondition{
				ValidationType: "vsphere-compute-resources",
				ValidationRule: "validation-resource-pool-dc0-c0-rp0",
				Message:        "One or more resource requirements were not satisfied",
				Details:        []string{},
				Failures:       []string{"Not enough resources available. CPU available: false, Memory available: true, Storage available: true"},
				Status:         corev1.ConditionFalse,
			},
				State: util.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name: "Resourcepool Memory not available",
			rule: v1alpha1.ComputeResourceRule{
				RuleName:    "Test Resourcepool Resource Validation rule",
				Scope:       entity.ResourcePool.String(),
				ClusterName: "DC0_C0",
				EntityName:  "DC0_C0_RP0",
				NodepoolResourceRequirements: []v1alpha1.NodepoolResourceRequirement{
					{
						Name:          "masterpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "500Ti",
						DiskSpace:     "50Gi",
					},
					{
						Name:          "workerpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "100Gi",
						DiskSpace:     "100Gi",
					},
				},
			},
			expectedResult: types.ValidationRuleResult{Condition: &vapi.ValidationCondition{
				ValidationType: "vsphere-compute-resources",
				ValidationRule: "validation-resource-pool-dc0-c0-rp0",
				Message:        "One or more resource requirements were not satisfied",
				Details:        []string{},
				Failures:       []string{"Not enough resources available. CPU available: true, Memory available: false, Storage available: true"},
				Status:         corev1.ConditionFalse,
			},
				State: util.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name: "Resourcepool Disk not available",
			rule: v1alpha1.ComputeResourceRule{
				RuleName:    "Test Resourcepool Resource Validation rule",
				Scope:       entity.ResourcePool.String(),
				ClusterName: "DC0_C0",
				EntityName:  "DC0_C0_RP0",
				NodepoolResourceRequirements: []v1alpha1.NodepoolResourceRequirement{
					{
						Name:          "masterpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "500Mi",
						DiskSpace:     "500Ti",
					},
					{
						Name:          "workerpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "1Gi",
						DiskSpace:     "100Ti",
					},
				},
			},
			expectedResult: types.ValidationRuleResult{Condition: &vapi.ValidationCondition{
				ValidationType: "vsphere-compute-resources",
				ValidationRule: "validation-resource-pool-dc0-c0-rp0",
				Message:        "One or more resource requirements were not satisfied",
				Details:        []string{},
				Failures:       []string{"Not enough resources available. CPU available: true, Memory available: true, Storage available: false"},
				Status:         corev1.ConditionFalse,
			},
				State: util.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name: "Duplicate scope resourcepool",
			rule: v1alpha1.ComputeResourceRule{
				RuleName:    "Test Resourcepool Resource Validation rule",
				Scope:       entity.ResourcePool.String(),
				ClusterName: "DC0_C1",
				EntityName:  "DC0_C1_RP0",
				NodepoolResourceRequirements: []v1alpha1.NodepoolResourceRequirement{
					{
						Name:          "masterpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "500Mi",
						DiskSpace:     "500Ti",
					},
					{
						Name:          "workerpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "1Gi",
						DiskSpace:     "100Ti",
					},
				},
			},
			expectedResult: types.ValidationRuleResult{Condition: &vapi.ValidationCondition{
				ValidationType: "vsphere-compute-resources",
				ValidationRule: "validation-resource-pool-dc0-c1-rp0",
				Message:        "Rule for scope already processed",
				Details:        []string{},
				Failures:       []string{"Rule for scope resource-pool-dc0-c1 already processed"},
				Status:         corev1.ConditionFalse,
			},
				State: util.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name: "Duplicate scope cluster",
			rule: v1alpha1.ComputeResourceRule{
				RuleName:   "Test Resourcepool Resource Validation rule",
				Scope:      entity.Cluster.String(),
				EntityName: "DC0_C1",
				NodepoolResourceRequirements: []v1alpha1.NodepoolResourceRequirement{
					{
						Name:          "masterpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "500Mi",
						DiskSpace:     "500Ti",
					},
					{
						Name:          "workerpool",
						NumberOfNodes: 1,
						CPU:           "1GHz",
						Memory:        "1Gi",
						DiskSpace:     "100Ti",
					},
				},
			},
			expectedResult: types.ValidationRuleResult{Condition: &vapi.ValidationCondition{
				ValidationType: "vsphere-compute-resources",
				ValidationRule: "validation-cluster-dc0-c1",
				Message:        "Rule for scope already processed",
				Details:        []string{},
				Failures:       []string{"Rule for scope cluster-dc0-c1 already processed"},
				Status:         corev1.ConditionFalse,
			},
				State: util.Ptr(vapi.ValidationFailed),
			},
		},
	}

	GetResourcePoolAndVMs = func(ctx context.Context, inventoryPath string, finder *find.Finder) (*mo.ResourcePool, *[]mo.VirtualMachine, error) {
		rpCPULimit := int64(80000)
		rpMemLimit := int64(500000)
		resourcePool := mo.ResourcePool{
			Config: vtypes.ResourceConfigSpec{
				CpuAllocation: vtypes.ResourceAllocationInfo{
					Limit: &rpCPULimit,
				},
				MemoryAllocation: vtypes.ResourceAllocationInfo{
					Limit: &rpMemLimit,
				},
			},
		}
		virtualmachines := []mo.VirtualMachine{
			{
				Summary: vtypes.VirtualMachineSummary{
					QuickStats: vtypes.VirtualMachineQuickStats{
						OverallCpuUsage: 1000,
						HostMemoryUsage: 50000,
					},
				},
			},
		}

		return &resourcePool, &virtualmachines, nil
	}

	seenScopes := map[string]bool{
		"resource-pool-dc0-c1": true,
		"cluster-dc0-c1":       true,
	}

	for _, tc := range testCases {
		vr, err := validationService.ReconcileComputeResourceValidationRule(tc.rule, finder, driver, seenScopes)
		test.CheckTestCase(t, vr, tc.expectedResult, err, tc.expectedErr)
	}
}
