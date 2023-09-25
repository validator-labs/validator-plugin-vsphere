package computeresources

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/api/v1alpha1"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/internal/constants"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/internal/vsphere"
	v8or "github.com/spectrocloud-labs/valid8or/api/v1alpha1"
	v8orconstants "github.com/spectrocloud-labs/valid8or/pkg/constants"
	"github.com/spectrocloud-labs/valid8or/pkg/types"
	"github.com/spectrocloud-labs/valid8or/pkg/util/ptr"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/units"
	"github.com/vmware/govmomi/vim25/mo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"strings"
)

type ComputeResourcesValidationService struct {
	log    logr.Logger
	driver *vsphere.VSphereCloudDriver
}

func NewComputeResourcesValidationService(log logr.Logger, driver *vsphere.VSphereCloudDriver) *ComputeResourcesValidationService {
	return &ComputeResourcesValidationService{
		log:    log,
		driver: driver,
	}
}

type resourceRequirement struct {
	CPU       resource.Quantity
	Memory    resource.Quantity
	DiskSpace resource.Quantity
}

func buildValidationResult(rule v1alpha1.ComputeResourceRule, validationType string) *types.ValidationResult {
	state := v8or.ValidationSucceeded
	latestCondition := v8or.DefaultValidationCondition()
	latestCondition.Message = fmt.Sprintf("All required compute resources were satisfied")
	latestCondition.ValidationRule = fmt.Sprintf("%s-%s-%s", v8orconstants.ValidationRulePrefix, rule.Scope, rule.EntityName)
	latestCondition.ValidationType = validationType

	return &types.ValidationResult{Condition: &latestCondition, State: &state}
}

func ghz(val int64) string {
	return fmt.Sprintf("%.1fGHz", float64(val)/1000)
}

func size(val int64) string {
	return units.ByteSize(val).String()
}

func (r *ResourceUsage) summarize(f func(int64) string) {
	r.Usage = 100 * float64(r.Used) / float64(r.Capacity)

	r.Summary.Usage = fmt.Sprintf("%.1f", r.Usage)
	r.Summary.Capacity = f(r.Capacity)
	r.Summary.Used = f(r.Used)
	r.Summary.Free = f(r.Free)
}

type ResourceUsageSummary struct {
	Used     string
	Free     string
	Capacity string
	Usage    string
}

type ResourceUsage struct {
	Used     int64
	Free     int64
	Capacity int64
	Usage    float64
	Summary  ResourceUsageSummary
}

type Usage struct {
	Memory  ResourceUsage
	CPU     ResourceUsage
	Storage ResourceUsage
}

func (c *ComputeResourcesValidationService) ReconcileComputeResourceValidationRule(rule v1alpha1.ComputeResourceRule, finder *find.Finder, driver *vsphere.VSphereCloudDriver) (*types.ValidationResult, error) {
	var res Usage

	vr := buildValidationResult(rule, constants.ValidationTypeComputeResources)

	resourceReq, err := getResourceRequirements(rule.NodepoolResourceRequirements)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	switch rule.Scope {
	case "cluster":
		var cluster mo.ClusterComputeResource
		var hosts []mo.HostSystem
		var datastores []mo.Datastore

		obj, err := finder.ClusterComputeResource(ctx, rule.EntityName)
		if err != nil {
			return nil, err
		}

		pc := property.DefaultCollector(obj.Client())
		err = pc.RetrieveOne(ctx, obj.Reference(), []string{"datastore", "host"}, &cluster)
		if err != nil {
			return nil, err
		}

		err = pc.Retrieve(ctx, cluster.Host, []string{"summary"}, &hosts)
		if err != nil {
			return nil, err
		}

		for _, host := range hosts {
			res.CPU.Capacity += int64(int32(host.Summary.Hardware.NumCpuCores) * host.Summary.Hardware.CpuMhz)
			res.CPU.Used += int64(host.Summary.QuickStats.OverallCpuUsage)

			res.Memory.Capacity += host.Summary.Hardware.MemorySize
			res.Memory.Used += int64(host.Summary.QuickStats.OverallMemoryUsage) << 20
		}

		err = pc.Retrieve(ctx, cluster.Datastore, []string{"summary"}, &datastores)
		if err != nil {
			return nil, err
		}

		res.Storage.Capacity, res.Storage.Free = getDatastoreInfo(datastores)

	case "resourcepool":
		var resourcePool mo.ResourcePool
		var virtualMachines []mo.VirtualMachine
		var cluster mo.ClusterComputeResource
		var datastores []mo.Datastore

		obj, err := finder.ResourcePool(ctx, fmt.Sprintf(constants.ResourcePoolInventoryPath, driver.Datacenter, rule.ClusterName, rule.EntityName))
		if err != nil {
			return nil, err
		}

		pc := property.DefaultCollector(obj.Client())
		err = pc.RetrieveOne(ctx, obj.Reference(), nil, &resourcePool)
		if err != nil {
			return nil, err
		}

		err = pc.Retrieve(ctx, resourcePool.Vm, nil, &virtualMachines)
		if err != nil {
			return nil, err
		}

		res.CPU.Capacity += *resourcePool.Config.CpuAllocation.Limit
		res.Memory.Capacity += *resourcePool.Config.MemoryAllocation.Limit << 20

		for _, vm := range virtualMachines {
			res.CPU.Used += int64(vm.Summary.QuickStats.OverallCpuUsage)
			res.Memory.Used += int64(vm.Summary.QuickStats.HostMemoryUsage) << 20
		}

		clusterObj, err := finder.ClusterComputeResource(ctx, rule.ClusterName)
		if err != nil {
			return nil, err
		}

		clusterPc := property.DefaultCollector(clusterObj.Client())
		err = clusterPc.RetrieveOne(ctx, clusterObj.Reference(), []string{"datastore", "host"}, &cluster)
		if err != nil {
			return nil, err
		}

		err = clusterPc.Retrieve(ctx, cluster.Datastore, []string{"summary"}, &datastores)
		if err != nil {
			return nil, err
		}

		res.Storage.Capacity, res.Storage.Free = getDatastoreInfo(datastores)

	case "host":
		var hostSystem mo.HostSystem
		var virtualMachines []mo.VirtualMachine

		var datastores []mo.Datastore

		obj, err := finder.HostSystem(ctx, fmt.Sprintf("%s", rule.EntityName))
		if err != nil {
			return nil, err
		}

		pc := property.DefaultCollector(obj.Client())
		err = pc.RetrieveOne(ctx, obj.Reference(), nil, &hostSystem)
		if err != nil {
			return nil, err
		}

		err = pc.Retrieve(ctx, hostSystem.Vm, nil, &virtualMachines)
		if err != nil {
			return nil, err
		}

		res.CPU.Capacity += int64(hostSystem.Summary.Hardware.CpuMhz * int32(hostSystem.Hardware.CpuInfo.NumCpuCores))
		res.Memory.Capacity += hostSystem.Summary.Hardware.MemorySize

		res.CPU.Used = int64(hostSystem.Summary.QuickStats.OverallCpuUsage)
		res.Memory.Used = int64(hostSystem.Summary.QuickStats.OverallMemoryUsage) << 20

		err = pc.Retrieve(ctx, hostSystem.Datastore, nil, &datastores)
		if err != nil {
			return nil, err
		}

		res.Storage.Capacity, res.Storage.Free = getDatastoreInfo(datastores)
	}

	res.CPU.Free = res.CPU.Capacity - res.CPU.Used
	res.CPU.summarize(ghz)

	res.Memory.Free = res.Memory.Capacity - res.Memory.Used
	res.Memory.summarize(size)

	res.Storage.Used = res.Storage.Capacity - res.Storage.Free
	res.Storage.summarize(size)
	freeCPU := convertStringToQuantity(sanitizeStrUnits(res.CPU.Summary.Free, "cpu"))
	freeMemory := convertStringToQuantity(sanitizeStrUnits(res.Memory.Summary.Free, "memory"))
	freeStorage := convertStringToQuantity(sanitizeStrUnits(res.Storage.Summary.Free, "storage"))

	cpuCapacityAvailable := requestedQuantityAvailable(freeCPU, resourceReq.CPU)
	memoryCapacityAvailable := requestedQuantityAvailable(freeMemory, resourceReq.Memory)
	diskCapacityAvailable := requestedQuantityAvailable(freeStorage, resourceReq.DiskSpace)

	if !cpuCapacityAvailable || !memoryCapacityAvailable || !diskCapacityAvailable {
		vr.State = ptr.Ptr(v8or.ValidationFailed)
		vr.Condition.Failures = append(vr.Condition.Failures, fmt.Sprintf("Not enough resources available. CPU available: %t, Memory available: %t, Storage available: %t", cpuCapacityAvailable, memoryCapacityAvailable, diskCapacityAvailable))
		vr.Condition.Message = "One or more resource requirements were not satisfied"
		vr.Condition.Status = corev1.ConditionFalse

		return vr, errors.New("Rule not satisfied")
	}

	return vr, nil
}

func getDatastoreInfo(datastores []mo.Datastore) (capacity int64, freeSpace int64) {
	for _, datastore := range datastores {
		// skip host local storage
		shared := datastore.Summary.MultipleHostAccess
		if shared != nil && *shared == false {
			continue
		}

		capacity += datastore.Summary.Capacity
		freeSpace += datastore.Summary.FreeSpace
	}

	return capacity, freeSpace
}

func requestedQuantityAvailable(freeResource resource.Quantity, requestedResource resource.Quantity) bool {
	available := freeResource.Cmp(requestedResource)

	switch available {
	case 0, -1:
		return false
	}

	return true
}

func convertStringToQuantity(resourceStr string) resource.Quantity {
	return resource.MustParse(resourceStr)
}

func getResourceRequirements(requirements []v1alpha1.NodepoolResourceRequirement) (*resourceRequirement, error) {
	var finalCPU, finalMemory, finalDisk resource.Quantity
	for _, requirement := range requirements {
		requiredCPU := sanitizeStrUnits(requirement.CPU, "cpu")
		totalCPU := getTotalQuantity(requiredCPU, requirement.NumberOfNodes)
		finalCPU.Add(totalCPU)

		totalMemory := getTotalQuantity(requirement.Memory, requirement.NumberOfNodes)
		finalMemory.Add(totalMemory)

		totalDisk := getTotalQuantity(requirement.DiskSpace, requirement.NumberOfNodes)
		finalDisk.Add(totalDisk)
	}

	return &resourceRequirement{
		CPU:       finalCPU,
		Memory:    finalMemory,
		DiskSpace: finalDisk,
	}, nil
}

func sanitizeStrUnits(resource string, resourceType string) string {
	switch resourceType {
	case "memory", "storage":
		return strings.TrimSuffix(resource, "B")
	}
	return strings.ReplaceAll(resource, "Hz", "")
}

func getTotalQuantity(quantity string, numberOfNodes int) resource.Quantity {
	var totalCPU resource.Quantity

	for i := 0; i < numberOfNodes; i++ {
		totalCPU.Add(resource.MustParse(quantity))
	}
	return totalCPU
}
