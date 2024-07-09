package vsphere

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func TestToVSphereVMs(t *testing.T) {
	tests := []struct {
		name        string
		params      []mo.VirtualMachine
		metrics     []performance.EntityMetric
		networks    []object.NetworkReference
		dsNames     []*object.Datastore
		folders     []*object.Folder
		hostSystems []mo.HostSystem
		ccrs        []*object.ClusterComputeResource
		parentsRef  []mo.VirtualMachine
		expectedVMs []VSphereVM
	}{
		{
			name: "VM Conversion",
			params: []mo.VirtualMachine{
				{
					Summary: types.VirtualMachineSummary{
						Config: types.VirtualMachineConfigSummary{
							Name:            "TestVM",
							NumCpu:          2,
							MemorySizeMB:    4096,
							NumVirtualDisks: 1,
						},
						Guest: &types.VirtualMachineGuestSummary{
							IpAddress: "192.168.1.100",
						},
						OverallStatus: "green",
						Vm:            &types.ManagedObjectReference{Value: "vm-123"},
					},
					Runtime: types.VirtualMachineRuntimeInfo{
						Host: &types.ManagedObjectReference{Value: "host-456"},
					},
					Datastore: []types.ManagedObjectReference{{Value: "ds-789"}},
					Guest: &types.GuestInfo{
						Net: []types.GuestNicInfo{
							{
								IpAddress: []string{"192.168.1.100"},
							},
						},
						IpAddress: "192.168.1.100",
					},
					Network: []types.ManagedObjectReference{{Value: "network-123"}},
					Config: &types.VirtualMachineConfigInfo{
						Hardware: types.VirtualHardware{
							Device: []types.BaseVirtualDevice{},
						},
					},
				},
			},
			metrics: []performance.EntityMetric{
				{
					Entity: types.ManagedObjectReference{Value: "vm-123"},
					Value: []performance.MetricSeries{
						{
							Name:  "cpu.corecount.usage.average",
							Value: []int64{3},
						},
						{
							Name:  "cpu.usage.average",
							Value: []int64{123},
						},
						{
							Name:  "mem.active.average",
							Value: []int64{3123},
						},
						{
							Name:  "mem.usage.average",
							Value: []int64{23},
						},
						{
							Name:  "disk.usage.average",
							Value: []int64{9883},
						},
						{
							Name:  "disk.provisioned.latest",
							Value: []int64{844},
						},
					},
				},
			},
			networks: []object.NetworkReference{},
			dsNames:  []*object.Datastore{},
			folders: []*object.Folder{
				{
					Common: object.Common{
						InventoryPath: "TestVM",
					},
				},
			},
			hostSystems: []mo.HostSystem{
				{
					Summary: types.HostListSummary{
						Host: &types.ManagedObjectReference{Value: "host-456"},
					},
					ManagedEntity: mo.ManagedEntity{
						Parent: &types.ManagedObjectReference{Value: "folder-123"},
					},
				},
			},
			ccrs: []*object.ClusterComputeResource{},
			parentsRef: []mo.VirtualMachine{
				{
					Summary: types.VirtualMachineSummary{
						Config: types.VirtualMachineConfigSummary{
							Name: "TestVM",
						},
					},
					ManagedEntity: mo.ManagedEntity{
						Parent: &types.ManagedObjectReference{Value: "folder-123"},
					},
				},
			},
			expectedVMs: []VSphereVM{
				{
					Name:         "TestVM",
					Type:         "vm-123",
					Status:       "green",
					IPAddress:    "192.168.1.100",
					Host:         "",
					CPU:          2,
					Memory:       4096,
					RootDiskSize: 1,
					Network: []VSphereNetwork{
						{
							Ip:        "192.168.1.100",
							Interface: "",
						},
					},
					VSphereVMInfo: VSphereVMInfo{
						Folder:    "",
						Datastore: "",
						Network:   "",
						Cluster:   "",
					},
					SSHInfo: SSHInfo{
						Username: "",
					},
					AdditionalDisk: []AdditionalDisk{},
					Metrics: Metrics{
						CPUCores:        "3",
						CPUUsage:        "1",
						MemoryBytes:     "3123",
						MemoryUsage:     "0",
						DiskUsage:       "9883",
						DiskProvisioned: "844",
					},
					Storage: []Datastore{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := ToVSphereVMs(tt.params, tt.metrics, tt.networks, tt.dsNames, tt.folders, tt.hostSystems, tt.ccrs, tt.parentsRef)
			assert.Equal(t, tt.expectedVMs, results)
		})
	}
}
