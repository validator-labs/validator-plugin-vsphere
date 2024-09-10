package vsphere

import (
	"context"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/validator-labs/validator-plugin-vsphere/api/vcenter"
)

// GetVM returns the vCenter VM if it exists
func (v *VCenterDriver) GetVM(ctx context.Context, finder *find.Finder, vmName string) (*object.VirtualMachine, error) {
	vm, err := finder.VirtualMachine(ctx, vmName)
	if err != nil {
		return nil, err
	}
	return vm, nil
}

// GetVMs returns a list of vCenter VMs
func (v *VCenterDriver) GetVMs(ctx context.Context, datacenter string) ([]vcenter.VM, error) {
	finder, v1, client, err := v.getVMClient(ctx, datacenter)
	if err != nil {
		return nil, err
	}

	vms, e := v.getVMs(ctx, v1, nil)
	if e != nil {
		return nil, e
	}

	return v.getVMInfo(ctx, finder, client, datacenter, v1, vms)
}

func (v *VCenterDriver) getVMClient(ctx context.Context, datacenter string) (*find.Finder, *view.ContainerView, *vim25.Client, error) {
	finder, _, err := v.GetFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return nil, nil, nil, err
	}

	vms, err := finder.VirtualMachineList(ctx, "*")
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to fetch vSphere vms")
	}

	client := vms[0].Client()
	m := view.NewManager(client)
	v1, err := m.CreateContainerView(ctx, client.ServiceContent.RootFolder, []string{"VirtualMachine", "ManagedEntity"}, true)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to get view manager while deleting vms")
	}

	return finder, v1, client, nil
}

func (v *VCenterDriver) getVMs(ctx context.Context, v1 *view.ContainerView, filter *property.Match) ([]mo.VirtualMachine, error) {
	vms := make([]mo.VirtualMachine, 0)
	var err error
	kind := []string{"VirtualMachine"}

	if filter != nil {
		// Retrieve all VM properties by passing ps == nil
		err = v1.RetrieveWithFilter(ctx, kind, nil, &vms, *filter)
	} else {
		// Retrieve name property for VMs
		err = v1.Retrieve(ctx, kind, []string{}, &vms)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get virtual machines")
	}

	return vms, nil
}

func (v *VCenterDriver) getVMInfo(ctx context.Context, finder *find.Finder, client *vim25.Client, datacenter string, v1 *view.ContainerView, vms []mo.VirtualMachine) ([]vcenter.VM, error) {
	metrics, err := v.GetMetrics(ctx, client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get view manager while fetching vSphere vms")
	}

	networks, err := finder.NetworkList(ctx, "*")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the networks while fetching vSphere vms")
	}

	datastores, err := finder.DatastoreList(ctx, "*")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get datastores while fetching vSphere vms")
	}

	folders, err := finder.FolderList(ctx, "*")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get folders while fetching vSphere vms")
	}

	hostSystems, err := v.getHostSystems(ctx, v1)
	if err != nil {
		return nil, err
	}

	_, ccrs, err := v.getClusterComputeResources(ctx, datacenter)
	if err != nil {
		return nil, err
	}

	vmParentRefs, err := v.getVMParentRefs(ctx, v1)
	if err != nil {
		return nil, err
	}

	return ToVSphereVMs(vms, metrics, networks, datastores, folders, hostSystems, ccrs, vmParentRefs), nil
}

func (v *VCenterDriver) getVMParentRefs(ctx context.Context, v1 *view.ContainerView) ([]mo.VirtualMachine, error) {
	var vms []mo.VirtualMachine
	err := v1.Retrieve(ctx, []string{"VirtualMachine"}, []string{"parent", "summary"}, &vms)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get virtual machines parents ref")
	}
	return vms, nil
}

// GetMetrics returns the metrics for the given VMs
func (v *VCenterDriver) GetMetrics(ctx context.Context, c *vim25.Client) ([]performance.EntityMetric, error) {
	m := view.NewManager(c)

	v1, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, nil, true)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := v1.Destroy(ctx); err != nil {
			v.log.Error(err, "failed to destroy view")
		}
	}()

	vmsRefs, e := v1.Find(ctx, []string{"VirtualMachine"}, nil)
	if e != nil {
		return nil, e
	}

	// Create a PerfManager
	perfManager := performance.NewManager(c)

	// Create PerfQuerySpec
	spec := types.PerfQuerySpec{
		MaxSample:  1,
		MetricId:   []types.PerfMetricId{{Instance: "*"}},
		IntervalId: 300,
	}

	// Query metrics
	names := []string{"cpu.usage.average", "cpu.corecount.usage.average", "mem.active.average", "mem.usage.average", "disk.usage.average", "disk.provisioned.latest"}
	sample, err := perfManager.SampleByName(ctx, spec, names, vmsRefs)
	if err != nil {
		return nil, err
	}

	result, err := perfManager.ToMetricSeries(ctx, sample)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ToVSphereVMs converts a list of VirtualMachines to a list of VSphereVMs
func ToVSphereVMs(params []mo.VirtualMachine, metrics []performance.EntityMetric, networks []object.NetworkReference, dsNames []*object.Datastore, folders []*object.Folder, hostSystems []mo.HostSystem, ccrs []*object.ClusterComputeResource, parentsRef []mo.VirtualMachine) []vcenter.VM {
	vms := make([]vcenter.VM, 0)
	for _, param := range params {
		vms = append(vms, ToVSphereVM(param, metrics, networks, dsNames, folders, hostSystems, ccrs, parentsRef))
	}
	return vms
}

// ToVSphereVM converts a VirtualMachine to a VSphereVM
func ToVSphereVM(param mo.VirtualMachine, metrics []performance.EntityMetric, networks []object.NetworkReference,
	dsNames []*object.Datastore, folders []*object.Folder, hostSystems []mo.HostSystem,
	ccrs []*object.ClusterComputeResource, parentsRef []mo.VirtualMachine) vcenter.VM {
	vm := vcenter.VM{
		Name:         param.Summary.Config.Name,
		Type:         param.Summary.Vm.Value,
		Status:       string(param.Summary.OverallStatus),
		IPAddress:    param.Guest.IpAddress,
		Host:         getHostName(param, hostSystems),
		CPU:          param.Summary.Config.NumCpu,
		Memory:       param.Summary.Config.MemorySizeMB,
		RootDiskSize: param.Summary.Config.NumVirtualDisks,
		Network:      getNetworks(param),
		VMInfo: vcenter.VMInfo{
			Folder:    getFolderName(param, parentsRef, folders),
			Datastore: getDatastore(param.Datastore, dsNames),
			Network:   getNetwork(networks, param.Network),
			Cluster:   getClusterName(param, hostSystems, ccrs),
		},
		SSHInfo: vcenter.SSHInfo{
			Username: param.Summary.Config.GuestId,
		},
		AdditionalDisk: getVMAdditionalDisks(param),
		Metrics:        ToVMMetrics(param.Summary.Vm.Value, metrics),
		Storage:        getStorage(param.Datastore, dsNames),
	}
	return vm
}

func getHostName(param mo.VirtualMachine, hostSystems []mo.HostSystem) string {
	hostSystem := getHostSystem(param.Runtime.Host, hostSystems)
	if hostSystem == nil {
		return ""
	}
	hostName := hostSystem.ManagedEntity.Name
	return hostName
}

func getNetworks(params mo.VirtualMachine) []vcenter.Network {
	if params.Guest == nil || params.Guest.Net == nil {
		return []vcenter.Network{}
	}
	networks := make([]vcenter.Network, 0)
	ipAddress := []string{}
	for _, param := range params.Guest.Net {
		ipAddress = append(ipAddress, param.IpAddress...)
	}
	for _, ipAddress := range ipAddress {
		networks = append(networks, vcenter.Network{
			IP: ipAddress,
		})
	}
	return networks
}

func getFolderName(param mo.VirtualMachine, parentsRef []mo.VirtualMachine, folders []*object.Folder) string {
	folderName := ""
	for _, ref := range parentsRef {
		if ref.Summary.Config.Name == param.Summary.Config.Name {
			if ref.ManagedEntity.Parent == nil {
				return ""
			}
			folderName = ref.ManagedEntity.Parent.Value
		}
	}

	if folderName == "" {
		return ""
	}

	for _, folder := range folders {
		if folder.Reference().Value == folderName {
			return getNameFromInventory(folder.InventoryPath)
		}
	}
	return ""
}

func getDatastore(ds []types.ManagedObjectReference, dsNames []*object.Datastore) string {
	if len(ds) == 0 {
		return ""
	}
	dataStore := ds[0].Value
	for _, ds := range dsNames {
		if ds.Reference().Value == dataStore {
			return getNameFromInventory(ds.InventoryPath)
		}
	}
	return ""
}

func getNetwork(networks []object.NetworkReference, n []types.ManagedObjectReference) string {
	if len(n) == 0 {
		return ""
	}

	networkName := n[0].Value
	for _, network := range networks {
		if network.Reference().Value == networkName {
			return getNameFromInventory(network.GetInventoryPath())
		}
	}
	return ""
}

func getClusterName(param mo.VirtualMachine, hostSystems []mo.HostSystem, ccrs []*object.ClusterComputeResource) string {
	hostSystem := getHostSystem(param.Runtime.Host, hostSystems)
	if hostSystem == nil {
		return ""
	}
	cluster := getVMCluster(hostSystem.ManagedEntity.Parent.Value, ccrs)
	if cluster == nil {
		return ""
	}
	clusterName := getNameFromInventory(cluster.InventoryPath)
	return clusterName
}

func getVMAdditionalDisks(param mo.VirtualMachine) []vcenter.AdditionalDisk {
	disks := []vcenter.AdditionalDisk{}
	if param.Config == nil {
		return disks
	}
	for _, device := range param.Config.Hardware.Device {
		switch disk := device.(type) {
		case *types.VirtualDisk:
			deviceInfo := disk.GetVirtualDevice()
			disks = append(disks, vcenter.AdditionalDisk{
				Name:      deviceInfo.DeviceInfo.(*types.Description).Label,
				Capacity:  deviceInfo.DeviceInfo.(*types.Description).Summary,
				Used:      "",
				Available: "",
				Usage:     "",
			})
		}
	}
	return disks
}

func getStorage(ds []types.ManagedObjectReference, dsNames []*object.Datastore) []vcenter.Datastore {
	if len(ds) == 0 {
		return nil
	}
	datastores := make([]vcenter.Datastore, 0)
	dsMap := make(map[string]string, 0)
	for _, n := range dsNames {
		dsMap[n.Reference().Value] = n.InventoryPath
	}
	for _, d := range ds {
		if path, ok := dsMap[d.Value]; ok {
			datastores = append(datastores, vcenter.Datastore{
				ID:   d.Value,
				Name: getNameFromInventory(path),
			})
		}
	}
	return datastores
}

// ToVMMetrics finds the EntityMetric with the provided name and converts it to Metrics
func ToVMMetrics(name string, metrics []performance.EntityMetric) vcenter.Metrics {
	for _, metric := range metrics {
		if metric.Entity.Value == name {
			return ToVsphereMetrics(metric)
		}
	}
	return vcenter.Metrics{}
}

// ToVsphereMetrics converts the EntityMetric to Metrics
func ToVsphereMetrics(metric performance.EntityMetric) vcenter.Metrics {
	return vcenter.Metrics{
		CPUCores:        getMetric("cpu.corecount.usage.average", metric.Value),
		CPUUsage:        getPercentage(getMetric("cpu.usage.average", metric.Value)),
		MemoryBytes:     getMetric("mem.active.average", metric.Value),
		MemoryUsage:     getPercentage(getMetric("mem.usage.average", metric.Value)),
		DiskUsage:       getMetric("disk.usage.average", metric.Value),
		DiskProvisioned: getMetric("disk.provisioned.latest", metric.Value),
	}
}

func getVMCluster(clusterName string, clusters []*object.ClusterComputeResource) *object.ClusterComputeResource {
	for _, cluster := range clusters {
		if cluster.ComputeResource.Reference().Value == clusterName {
			return cluster
		}
	}
	return nil
}

func getNameFromInventory(inventoryPath string) string {
	arr := strings.Split(inventoryPath, "/")
	return arr[len(arr)-1]
}

func getMetric(name string, series []performance.MetricSeries) string {
	for _, val := range series {
		if val.Name == name {
			if len(val.Value) > 0 {
				return strconv.FormatInt(val.Value[0], 10)
			}
			return ""
		}
	}
	return "0.0"
}

func getPercentage(param string) string {
	if param == "" {
		return ""
	}
	if i, err := strconv.ParseInt(param, 10, 64); err == nil {
		return strconv.FormatInt(i/100, 10)
	}
	return ""
}
