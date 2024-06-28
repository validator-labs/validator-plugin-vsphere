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
)

type VSphereVM struct {
	Name           string
	Type           string
	Status         string
	IpAddress      string
	Host           string
	Cpu            int32
	Memory         int32
	RootDiskSize   int32
	Network        []VSphereNetwork
	LibvirtVmInfo  LibvirtVmInfo
	VSphereVmInfo  VSphereVmInfo
	SshInfo        SshInfo
	AdditionalDisk []AdditionalDisk
	Metrics        Metrics
	Storage        []Datastore
}

type VSphereNetwork struct {
	Type      string
	Ip        string
	Interface string
}

type LibvirtVmInfo struct {
	ImagePool string
	DataPool  string
}

type VSphereVmInfo struct {
	Folder    string
	Cluster   string
	Datastore string
	Network   string
}

type SshInfo struct {
	Username   string
	Password   string
	PublicKey  []string
	PrivateKey []string
}

type AdditionalDisk struct {
	Name      string
	Device    string
	Capacity  string
	Used      string
	Available string
	Usage     string
}

type Metrics struct {
	CpuCores        string
	CpuUsage        string
	MemoryBytes     string
	MemoryUsage     string
	DiskUsage       string
	DiskProvisioned string
}

type Datastore struct {
	Name string
	Id   string
}

func (v *VSphereCloudDriver) GetVMIfExists(ctx context.Context, finder *find.Finder, datacenter, cluster, vmName string) (bool, *object.VirtualMachine, error) {
	vm, err := finder.VirtualMachine(ctx, vmName)
	if err != nil {
		return false, nil, err
	}
	return true, vm, nil
}

func (v *VSphereCloudDriver) GetVSphereVms(ctx context.Context, dcName string) ([]VSphereVM, error) {
	finder, v1, client, err := v.getVmClient(ctx, dcName)
	if err != nil {
		return nil, err
	}

	vms, e := v.getVms(ctx, v1, nil)
	if e != nil {
		return nil, e
	}

	return v.getVmInfo(ctx, finder, client, v1, vms)
}

func (v *VSphereCloudDriver) getVmClient(ctx context.Context, dcName string) (*find.Finder, *view.ContainerView, *vim25.Client, error) {
	finder, _, err := v.GetFinderWithDatacenter(ctx, dcName)
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

func (v *VSphereCloudDriver) getVms(ctx context.Context, v1 *view.ContainerView, filter *property.Match) ([]mo.VirtualMachine, error) {
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

func (v *VSphereCloudDriver) getVmInfo(ctx context.Context, finder *find.Finder, client *vim25.Client, v1 *view.ContainerView, vms []mo.VirtualMachine) ([]VSphereVM, error) {
	metrics, err := v.GetMetrics(ctx, client, vms)
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

	ccrs, err := v.getClusterComputeResources(ctx, finder)
	if err != nil {
		return nil, err
	}

	vmParentRefs, err := v.getVmParentRefs(ctx, v1)
	if err != nil {
		return nil, err
	}

	return ToVSphereVMs(vms, metrics, networks, datastores, folders, hostSystems, ccrs, vmParentRefs), nil
}

func (v *VSphereCloudDriver) getVmParentRefs(ctx context.Context, v1 *view.ContainerView) ([]mo.VirtualMachine, error) {
	var vms []mo.VirtualMachine
	err := v1.Retrieve(ctx, []string{"VirtualMachine"}, []string{"parent", "summary"}, &vms)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get virtual machines parents ref")
	}
	return vms, nil
}

func (v *VSphereCloudDriver) GetMetrics(ctx context.Context, c *vim25.Client, vms []mo.VirtualMachine) ([]performance.EntityMetric, error) {
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

func ToVSphereVMs(params []mo.VirtualMachine, metrics []performance.EntityMetric, networks []object.NetworkReference, dsNames []*object.Datastore, folders []*object.Folder, hostSystems []mo.HostSystem, ccrs []*object.ClusterComputeResource, parentsRef []mo.VirtualMachine) []VSphereVM {
	vms := make([]VSphereVM, 0)
	for _, param := range params {
		vms = append(vms, ToVSphereVM(param, metrics, networks, dsNames, folders, hostSystems, ccrs, parentsRef))
	}
	return vms
}

func ToVSphereVM(param mo.VirtualMachine, metrics []performance.EntityMetric, networks []object.NetworkReference,
	dsNames []*object.Datastore, folders []*object.Folder, hostSystems []mo.HostSystem,
	ccrs []*object.ClusterComputeResource, parentsRef []mo.VirtualMachine) VSphereVM {
	vm := VSphereVM{
		Name:          param.Summary.Config.Name,
		Type:          param.Summary.Vm.Value,
		Status:        string(param.Summary.OverallStatus),
		IpAddress:     param.Guest.IpAddress,
		Host:          getHostName(param, hostSystems),
		Cpu:           param.Summary.Config.NumCpu,
		Memory:        param.Summary.Config.MemorySizeMB,
		RootDiskSize:  param.Summary.Config.NumVirtualDisks,
		LibvirtVmInfo: LibvirtVmInfo{},
		Network:       getNetworks(param),
		VSphereVmInfo: VSphereVmInfo{
			Folder:    getFolderName(param, parentsRef, folders),
			Datastore: getDatastore(param.Datastore, dsNames),
			Network:   getNetwork(networks, param.Network),
			Cluster:   getClusterName(param, hostSystems, ccrs),
		},
		SshInfo: SshInfo{
			Username: param.Summary.Config.GuestId,
		},
		AdditionalDisk: getVmAdditionalDisks(param),
		Metrics:        ToVmMetrics(param.Summary.Vm.Value, metrics),
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

func getNetworks(params mo.VirtualMachine) []VSphereNetwork {
	if params.Guest == nil || params.Guest.Net == nil {
		return []VSphereNetwork{}
	}
	networks := make([]VSphereNetwork, 0)
	ipAddress := []string{}
	for _, param := range params.Guest.Net {
		ipAddress = append(ipAddress, param.IpAddress...)
	}
	for _, ipAddress := range ipAddress {
		networks = append(networks, VSphereNetwork{
			Ip: ipAddress,
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
	cluster := getVmCluster(hostSystem.ManagedEntity.Parent.Value, ccrs)
	if cluster == nil {
		return ""
	}
	clusterName := getNameFromInventory(cluster.InventoryPath)
	return clusterName
}

func getVmAdditionalDisks(param mo.VirtualMachine) []AdditionalDisk {
	disks := []AdditionalDisk{}
	if param.Config == nil {
		return disks
	}
	for _, device := range param.Config.Hardware.Device {
		switch disk := device.(type) {
		case *types.VirtualDisk:
			deviceInfo := disk.GetVirtualDevice()
			disks = append(disks, AdditionalDisk{
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

func getStorage(ds []types.ManagedObjectReference, dsNames []*object.Datastore) []Datastore {
	if len(ds) == 0 {
		return nil
	}
	datastores := make([]Datastore, 0)
	dsMap := make(map[string]string, 0)
	for _, n := range dsNames {
		dsMap[n.Reference().Value] = n.InventoryPath
	}
	for _, d := range ds {
		if path, ok := dsMap[d.Value]; ok {
			datastores = append(datastores, Datastore{
				Id:   d.Value,
				Name: getNameFromInventory(path),
			})
		}
	}
	return datastores
}

func ToVmMetrics(name string, metrics []performance.EntityMetric) Metrics {
	for _, metric := range metrics {
		if metric.Entity.Value == name {
			return ToVsphereMetrics(metric)
		}
	}
	return Metrics{}
}

func ToVsphereMetrics(metric performance.EntityMetric) Metrics {
	return Metrics{
		CpuCores:        getMetric("cpu.corecount.usage.average", metric.Value),
		CpuUsage:        getPercentage(getMetric("cpu.usage.average", metric.Value)),
		MemoryBytes:     getMetric("mem.active.average", metric.Value),
		MemoryUsage:     getPercentage(getMetric("mem.usage.average", metric.Value)),
		DiskUsage:       getMetric("disk.usage.average", metric.Value),
		DiskProvisioned: getMetric("disk.provisioned.latest", metric.Value),
	}
}

func getVmCluster(clusterName string, clusters []*object.ClusterComputeResource) *object.ClusterComputeResource {
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
