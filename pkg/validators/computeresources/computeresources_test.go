package computeresources

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25/mo"
	vtypes "github.com/vmware/govmomi/vim25/types"
	corev1 "k8s.io/api/core/v1"

	"github.com/validator-labs/validator-plugin-vsphere/api/v1alpha1"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/vcsim"
	vapi "github.com/validator-labs/validator/api/v1alpha1"
	"github.com/validator-labs/validator/pkg/types"
	"github.com/validator-labs/validator/pkg/util"
)

func TestComputeResourcesValidationService_ReconcileComputeResourceValidationRule(t *testing.T) {
	var log logr.Logger

	vcSim := vcsim.NewVCSim("admin@vsphere.local", 8447, log)
	vcSim.Start()
	defer vcSim.Shutdown()

	finder := find.NewFinder(vcSim.Driver.Client.Client)

	validationService := NewValidationService(log, vcSim.Driver)
	testCases := []struct {
		name           string
		expectedErr    error
		rule           v1alpha1.ComputeResourceRule
		expectedResult types.ValidationRuleResult
	}{
		{
			name: "All Resources available",
			rule: v1alpha1.ComputeResourceRule{
				Name:        "Test Resource Validation rule",
				ClusterName: "DC0_C0",
				Scope:       "cluster",
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
				ValidationRule: "validation-cluster-DC0_C0",
				Message:        "All required compute resources were satisfied",
				Details:        []string{},
				Failures:       nil,
				Status:         corev1.ConditionTrue,
			},
				State: util.Ptr(vapi.ValidationSucceeded),
			},
		},
		{
			name:        "cluster CPU not available",
			expectedErr: errInsufficientComputeResources,
			rule: v1alpha1.ComputeResourceRule{
				Name:        "Test Resource Validation rule",
				ClusterName: "DC0_C0",
				Scope:       "cluster",
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
				ValidationRule: "validation-cluster-DC0_C0",
				Message:        "One or more resource requirements were not satisfied",
				Details:        []string{},
				Failures:       []string{"Not enough resources available. CPU available: false, Memory available: true, Storage available: true"},
				Status:         corev1.ConditionFalse,
			},
				State: util.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name:        "cluster Memory not available",
			expectedErr: errInsufficientComputeResources,
			rule: v1alpha1.ComputeResourceRule{
				Name:        "Test Resource Validation rule",
				ClusterName: "DC0_C0",
				Scope:       "cluster",
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
				ValidationRule: "validation-cluster-DC0_C0",
				Message:        "One or more resource requirements were not satisfied",
				Details:        []string{},
				Failures:       []string{"Not enough resources available. CPU available: true, Memory available: false, Storage available: true"},
				Status:         corev1.ConditionFalse,
			},
				State: util.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name:        "cluster Disk not available",
			expectedErr: errInsufficientComputeResources,
			rule: v1alpha1.ComputeResourceRule{
				Name:        "Test Resource Validation rule",
				ClusterName: "DC0_C0",
				Scope:       "cluster",
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
				ValidationRule: "validation-cluster-DC0_C0",
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
				Name:       "Test Host Resource Validation rule",
				Scope:      "host",
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
				ValidationRule: "validation-host-DC0_C0_H0",
				Message:        "All required compute resources were satisfied",
				Details:        []string{},
				Failures:       nil,
				Status:         corev1.ConditionTrue,
			},
				State: util.Ptr(vapi.ValidationSucceeded),
			},
		},
		{
			name:        "Host CPU not available",
			expectedErr: errInsufficientComputeResources,
			rule: v1alpha1.ComputeResourceRule{
				Name:       "Test Host Resource Validation rule",
				Scope:      "host",
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
				ValidationRule: "validation-host-DC0_C0_H0",
				Message:        "One or more resource requirements were not satisfied",
				Details:        []string{},
				Failures:       []string{"Not enough resources available. CPU available: false, Memory available: true, Storage available: true"},
				Status:         corev1.ConditionFalse,
			},
				State: util.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name:        "Host Memory not available",
			expectedErr: errInsufficientComputeResources,
			rule: v1alpha1.ComputeResourceRule{
				Name:       "Test Host Resource Validation rule",
				Scope:      "host",
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
				ValidationRule: "validation-host-DC0_C0_H0",
				Message:        "One or more resource requirements were not satisfied",
				Details:        []string{},
				Failures:       []string{"Not enough resources available. CPU available: true, Memory available: false, Storage available: true"},
				Status:         corev1.ConditionFalse,
			},
				State: util.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name:        "Host Disk not available",
			expectedErr: errInsufficientComputeResources,
			rule: v1alpha1.ComputeResourceRule{
				Name:       "Test Host Resource Validation rule",
				Scope:      "host",
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
				ValidationRule: "validation-host-DC0_C0_H0",
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
				Name:        "Test Host Resource Validation rule",
				Scope:       "resourcepool",
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
				ValidationRule: "validation-resourcepool-DC0_C0_RP0",
				Message:        "All required compute resources were satisfied",
				Details:        []string{},
				Failures:       nil,
				Status:         corev1.ConditionTrue,
			},
				State: util.Ptr(vapi.ValidationSucceeded),
			},
		},
		{
			name:        "Resourcepool CPU not available",
			expectedErr: errInsufficientComputeResources,
			rule: v1alpha1.ComputeResourceRule{
				Name:        "Test Host Resource Validation rule",
				Scope:       "resourcepool",
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
				ValidationRule: "validation-resourcepool-DC0_C0_RP0",
				Message:        "One or more resource requirements were not satisfied",
				Details:        []string{},
				Failures:       []string{"Not enough resources available. CPU available: false, Memory available: true, Storage available: true"},
				Status:         corev1.ConditionFalse,
			},
				State: util.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name:        "Resourcepool Memory not available",
			expectedErr: errInsufficientComputeResources,
			rule: v1alpha1.ComputeResourceRule{
				Name:        "Test Resourcepool Resource Validation rule",
				Scope:       "resourcepool",
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
				ValidationRule: "validation-resourcepool-DC0_C0_RP0",
				Message:        "One or more resource requirements were not satisfied",
				Details:        []string{},
				Failures:       []string{"Not enough resources available. CPU available: true, Memory available: false, Storage available: true"},
				Status:         corev1.ConditionFalse,
			},
				State: util.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name:        "Resourcepool Disk not available",
			expectedErr: errInsufficientComputeResources,
			rule: v1alpha1.ComputeResourceRule{
				Name:        "Test Resourcepool Resource Validation rule",
				Scope:       "resourcepool",
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
				ValidationRule: "validation-resourcepool-DC0_C0_RP0",
				Message:        "One or more resource requirements were not satisfied",
				Details:        []string{},
				Failures:       []string{"Not enough resources available. CPU available: true, Memory available: true, Storage available: false"},
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

	for _, tc := range testCases {
		vr, err := validationService.ReconcileComputeResourceValidationRule(tc.rule, finder, vcSim.Driver)
		util.CheckTestCase(t, vr, tc.expectedResult, err, tc.expectedErr)
	}
}
