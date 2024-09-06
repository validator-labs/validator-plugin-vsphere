// Package computeresources handles compute resource rule reconciliation.
package computeresources

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/units"
	"github.com/vmware/govmomi/vim25/mo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/validator-labs/validator-plugin-vsphere/api/v1alpha1"
	"github.com/validator-labs/validator-plugin-vsphere/api/vcenter"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/constants"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/vsphere"
	vapi "github.com/validator-labs/validator/api/v1alpha1"
	vapiconstants "github.com/validator-labs/validator/pkg/constants"
	"github.com/validator-labs/validator/pkg/types"
	"github.com/validator-labs/validator/pkg/util"
)

var (
	// GetResourcePoolAndVMs is defined to enable monkey patching the getResourcePoolAndVMs function in integration tests
	GetResourcePoolAndVMs           = getResourcePoolAndVMs
	errInsufficientComputeResources = errors.New("compute resources rule not satisfied")
	errRuleAlreadyProcessed         = errors.New("rule for scope already processed")
)

// ValidationService is a service that validates compute resource rules
type ValidationService struct {
	log    logr.Logger
	driver *vsphere.CloudDriver
}

// NewValidationService creates a new ValidationService
func NewValidationService(log logr.Logger, driver *vsphere.CloudDriver) *ValidationService {
	return &ValidationService{
		log:    log,
		driver: driver,
	}
}

type resourceRequirement struct {
	CPU       resource.Quantity
	Memory    resource.Quantity
	DiskSpace resource.Quantity
}

func buildValidationResult(rule v1alpha1.ComputeResourceRule, validationType string) *types.ValidationRuleResult {
	state := vapi.ValidationSucceeded
	latestCondition := vapi.DefaultValidationCondition()
	latestCondition.Message = "All required compute resources were satisfied"
	latestCondition.ValidationRule = fmt.Sprintf("%s-%s-%s", vapiconstants.ValidationRulePrefix, rule.Scope, rule.EntityName)
	latestCondition.ValidationType = validationType

	return &types.ValidationRuleResult{Condition: &latestCondition, State: &state}
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

// ResourceUsageSummary provides a summary of resource usage
type ResourceUsageSummary struct {
	Used     string
	Free     string
	Capacity string
	Usage    string
}

// ResourceUsage provides resource usage information
type ResourceUsage struct {
	Used     int64
	Free     int64
	Capacity int64
	Usage    float64
	Summary  ResourceUsageSummary
}

// Usage provides memory cpu and storage usage information
type Usage struct {
	Memory  ResourceUsage
	CPU     ResourceUsage
	Storage ResourceUsage
}

// ReconcileComputeResourceValidationRule reconciles the compute resource rule
func (c *ValidationService) ReconcileComputeResourceValidationRule(rule v1alpha1.ComputeResourceRule, finder *find.Finder, driver *vsphere.CloudDriver, seenScopes map[string]bool) (*types.ValidationRuleResult, error) {

	vr := buildValidationResult(rule, constants.ValidationTypeComputeResources)

	key, err := GetScopeKey(rule)
	if err != nil {
		return vr, err
	}
	if seenScopes[key] {
		vr.State = util.Ptr(vapi.ValidationFailed)
		vr.Condition.Message = "Rule for scope already processed"
		vr.Condition.Failures = append(vr.Condition.Failures, fmt.Sprintf("Rule for scope %s already processed", key))
		vr.Condition.Status = corev1.ConditionFalse
		return vr, errRuleAlreadyProcessed
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var res *Usage
	switch rule.Scope {
	case vcenter.Cluster:
		res, err = clusterUsage(ctx, rule, finder)
	case vcenter.ResourcePool:
		res, err = resourcePoolUsage(ctx, rule, finder, driver)
	case vcenter.Host:
		res, err = hostUsage(ctx, rule, finder)
	default:
		err = fmt.Errorf("unsupported scope: %s", rule.Scope)
	}
	if err != nil {
		return vr, err
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

	resourceReq := getResourceRequirements(rule.NodepoolResourceRequirements)
	cpuCapacityAvailable := requestedQuantityAvailable(freeCPU, resourceReq.CPU)
	memoryCapacityAvailable := requestedQuantityAvailable(freeMemory, resourceReq.Memory)
	diskCapacityAvailable := requestedQuantityAvailable(freeStorage, resourceReq.DiskSpace)

	if !cpuCapacityAvailable || !memoryCapacityAvailable || !diskCapacityAvailable {
		vr.State = util.Ptr(vapi.ValidationFailed)
		vr.Condition.Failures = append(vr.Condition.Failures, fmt.Sprintf("Not enough resources available. CPU available: %t, Memory available: %t, Storage available: %t", cpuCapacityAvailable, memoryCapacityAvailable, diskCapacityAvailable))
		vr.Condition.Message = "One or more resource requirements were not satisfied"
		vr.Condition.Status = corev1.ConditionFalse
		return vr, errInsufficientComputeResources
	}

	return vr, nil
}

func clusterUsage(ctx context.Context, rule v1alpha1.ComputeResourceRule, finder *find.Finder) (*Usage, error) {
	var res Usage

	// disk space
	datastores, hosts, err := clusterResources(ctx, finder, rule.EntityName)
	if err != nil {
		return nil, err
	}
	res.Storage.Capacity, res.Storage.Free = getDatastoreInfo(datastores)

	// cpu & memory
	for _, host := range hosts {
		addHostUsage(&res, host)
	}

	return &res, nil
}

func hostUsage(ctx context.Context, rule v1alpha1.ComputeResourceRule, finder *find.Finder) (*Usage, error) {
	var res Usage

	obj, err := finder.HostSystem(ctx, rule.EntityName)
	if err != nil {
		return nil, err
	}
	pc := property.DefaultCollector(obj.Client())

	// cpu & memory
	var hostSystem mo.HostSystem
	if err := pc.RetrieveOne(ctx, obj.Reference(), nil, &hostSystem); err != nil {
		return nil, err
	}
	addHostUsage(&res, hostSystem)

	// disk space
	var datastores []mo.Datastore
	if err := pc.Retrieve(ctx, hostSystem.Datastore, nil, &datastores); err != nil {
		return nil, err
	}
	res.Storage.Capacity, res.Storage.Free = getDatastoreInfo(datastores)

	return &res, nil
}

func resourcePoolUsage(ctx context.Context, rule v1alpha1.ComputeResourceRule, finder *find.Finder, driver *vsphere.CloudDriver) (*Usage, error) {
	var res Usage

	// cpu & memory
	inventoryPath := fmt.Sprintf(constants.ResourcePoolInventoryPath, driver.Datacenter, rule.ClusterName, rule.EntityName)
	if rule.EntityName == constants.ClusterDefaultResourcePoolName {
		inventoryPath = fmt.Sprintf("/%s/host/%s/%s", driver.Datacenter, rule.ClusterName, rule.EntityName)
	}
	resourcePool, virtualMachines, err := GetResourcePoolAndVMs(ctx, inventoryPath, finder)
	if err != nil {
		return nil, err
	}

	res.CPU.Capacity += *resourcePool.Config.CpuAllocation.Limit
	res.Memory.Capacity += *resourcePool.Config.MemoryAllocation.Limit << 20

	for _, vm := range *virtualMachines {
		res.CPU.Used += int64(vm.Summary.QuickStats.OverallCpuUsage)
		res.Memory.Used += int64(vm.Summary.QuickStats.HostMemoryUsage) << 20
	}

	// disk space
	datastores, _, err := clusterResources(ctx, finder, rule.ClusterName)
	if err != nil {
		return nil, err
	}
	res.Storage.Capacity, res.Storage.Free = getDatastoreInfo(datastores)

	return &res, nil
}

func clusterResources(ctx context.Context, finder *find.Finder, path string) ([]mo.Datastore, []mo.HostSystem, error) {
	obj, err := finder.ClusterComputeResource(ctx, path)
	if err != nil {
		return nil, nil, err
	}
	pc := property.DefaultCollector(obj.Client())

	var cluster mo.ClusterComputeResource
	if err := pc.RetrieveOne(ctx, obj.Reference(), []string{"datastore", "host"}, &cluster); err != nil {
		return nil, nil, err
	}

	var datastores []mo.Datastore
	if err := pc.Retrieve(ctx, cluster.Datastore, []string{"summary"}, &datastores); err != nil {
		return nil, nil, err
	}

	var hosts []mo.HostSystem
	if err = pc.Retrieve(ctx, cluster.Host, []string{"summary"}, &hosts); err != nil {
		return nil, nil, err
	}

	return datastores, hosts, nil
}

func addHostUsage(res *Usage, host mo.HostSystem) {
	res.CPU.Capacity += int64(int32(host.Summary.Hardware.NumCpuCores) * host.Summary.Hardware.CpuMhz)
	res.CPU.Used += int64(host.Summary.QuickStats.OverallCpuUsage)

	res.Memory.Capacity += host.Summary.Hardware.MemorySize
	res.Memory.Used += int64(host.Summary.QuickStats.OverallMemoryUsage) << 20
}

func getResourcePoolAndVMs(ctx context.Context, inventoryPath string, finder *find.Finder) (*mo.ResourcePool, *[]mo.VirtualMachine, error) {
	obj, err := getResourcePoolObj(ctx, inventoryPath, finder)
	if err != nil {
		return nil, nil, err
	}
	pc := property.DefaultCollector(obj.Client())

	var resourcePool mo.ResourcePool
	if err := pc.RetrieveOne(ctx, obj.Reference(), nil, &resourcePool); err != nil {
		return nil, nil, err
	}

	var virtualMachines []mo.VirtualMachine
	if err := pc.Retrieve(ctx, resourcePool.Vm, nil, &virtualMachines); err != nil {
		return nil, nil, err
	}

	return &resourcePool, &virtualMachines, nil
}

func getResourcePoolObj(ctx context.Context, inventoryPath string, finder *find.Finder) (*object.ResourcePool, error) {
	return finder.ResourcePool(ctx, inventoryPath)
}

func getDatastoreInfo(datastores []mo.Datastore) (capacity int64, freeSpace int64) {
	for _, datastore := range datastores {
		// skip host local storage
		shared := datastore.Summary.MultipleHostAccess
		if shared != nil && !*shared {
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

func getResourceRequirements(requirements []v1alpha1.NodepoolResourceRequirement) *resourceRequirement {
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
	}
}

func sanitizeStrUnits(resource string, resourceType string) string {
	switch resourceType {
	case "memory", "storage":
		return strings.TrimSuffix(resource, "B")
	}
	return strings.ReplaceAll(resource, "Hz", "")
}

func getTotalQuantity(quantity string, numberOfNodes int) resource.Quantity {
	var totalQuantity resource.Quantity

	for i := 0; i < numberOfNodes; i++ {
		totalQuantity.Add(resource.MustParse(quantity))
	}
	return totalQuantity
}

// GetScopeKey returns a formatted key depending on the scope of a rule
func GetScopeKey(rule v1alpha1.ComputeResourceRule) (string, error) {
	switch rule.Scope {
	case vcenter.Cluster:
		return fmt.Sprintf("%s-%s", rule.Scope, rule.EntityName), nil
	case vcenter.Host:
		return fmt.Sprintf("%s-%s", rule.Scope, rule.EntityName), nil
	case vcenter.ResourcePool:
		return fmt.Sprintf("%s-%s", rule.Scope, rule.ClusterName), nil
	default:
		return "", fmt.Errorf("unsupported scope: %s", rule.Scope)
	}
}
