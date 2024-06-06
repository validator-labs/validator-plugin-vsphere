package vsphere

import (
	"context"
	"fmt"
	"golang.org/x/exp/slices"
	"time"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/govc/host/service"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type VSphereHostSystem struct {
	Name      string
	Reference string
}

type HostDateInfo struct {
	HostName   string
	NtpServers []string
	types.HostDateTimeInfo
	Service       *types.HostService
	Current       *time.Time
	ClientStatus  string
	ServiceStatus string
}

func (v *VSphereCloudDriver) GetHostIfExists(ctx context.Context, finder *find.Finder, datacenter, clusterName, hostName string) (bool, *object.HostSystem, error) {
	path := fmt.Sprintf("/%s/host/%s/%s", datacenter, clusterName, hostName)
	// Handle datacenter level hosts
	if clusterName == "" {
		path = fmt.Sprintf("/%s/host/%s", datacenter, hostName)
	}
	host, err := finder.HostSystem(ctx, path)
	if err != nil {
		return false, nil, err
	}
	return true, host, nil
}

func (v *VSphereCloudDriver) GetVSphereHostSystems(ctx context.Context, datacenter, cluster string) ([]VSphereHostSystem, error) {
	finder, _, err := v.GetFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/%s/host/%s", datacenter, cluster)
	if cluster == "" {
		path = fmt.Sprintf("/%s/host/*", datacenter)
	}

	hss, err := finder.HostSystemList(ctx, path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch vSphere host systems")
	}
	if len(hss) == 0 {
		return nil, errors.New("No host systems found")
	}

	hostSystems := make([]VSphereHostSystem, 0)
	for _, hs := range hss {
		hostSystems = append(hostSystems, VSphereHostSystem{
			Name:      hs.Name(),
			Reference: hs.Reference().String(),
		})
	}

	return hostSystems, nil
}

func (v *VSphereCloudDriver) GetHostClusterMapping(ctx context.Context) (map[string]string, error) {
	m := view.NewManager(v.Client.Client)
	pc := property.DefaultCollector(v.Client.Client)
	var hostClusterMapping = make(map[string]string)

	containerView, err := m.CreateContainerView(ctx, v.Client.Client.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating containerview for hostsystems")
	}

	hosts, msgErr := v.getHostSystems(ctx, containerView)
	if msgErr != nil {
		return nil, msgErr
	}

	for _, host := range hosts {
		var cluster mo.ManagedEntity
		err = pc.RetrieveOne(ctx, *host.Parent, []string{"name"}, &cluster)
		if err != nil {
			return nil, err
		}
		hostClusterMapping[host.Name] = cluster.Name
	}

	return hostClusterMapping, nil
}

func (v *VSphereCloudDriver) getHostSystems(ctx context.Context, v1 *view.ContainerView) ([]mo.HostSystem, error) {
	var hs []mo.HostSystem
	e := v1.Retrieve(ctx, []string{"HostSystem"}, []string{"summary", "name", "parent"}, &hs)
	if e != nil {
		return nil, errors.Wrap(e, "failed to get host systems")
	}
	return hs, nil
}

func (info *HostDateInfo) servers() []string {
	return info.NtpConfig.Server
}

func getHostSystem(hostNameObj *types.ManagedObjectReference, hostSystems []mo.HostSystem) *mo.HostSystem {
	if hostNameObj == nil {
		return nil
	}
	for _, host := range hostSystems {
		if host.Summary.Host.Value == hostNameObj.Value {
			return &host
		}
	}
	return nil
}

func (v *VSphereCloudDriver) ValidateHostNTPSettings(ctx context.Context, finder *find.Finder, datacenter, clusterName string, hosts []string) (bool, []string, error) {
	var hostsDateInfo []HostDateInfo
	var failures []string

	for _, host := range hosts {
		_, hostObj, err := v.GetHostIfExists(ctx, finder, datacenter, clusterName, host)
		if err != nil {
			return false, nil, err
		}

		s, err := hostObj.ConfigManager().DateTimeSystem(ctx)
		if err != nil {
			return false, nil, err
		}

		var hs mo.HostDateTimeSystem
		if err = s.Properties(ctx, s.Reference(), nil, &hs); err != nil {
			return false, nil, err
		}

		ss, err := hostObj.ConfigManager().ServiceSystem(ctx)
		if err != nil {
			return false, nil, err
		}

		services, err := ss.Service(ctx)
		if err != nil {
			return false, nil, err
		}

		res := &HostDateInfo{HostDateTimeInfo: hs.DateTimeInfo}

		for i, service := range services {
			if service.Key == "ntpd" {
				res.Service = &services[i]
				break
			}
		}

		if res.Service == nil {
			failures = append(failures, fmt.Sprintf("Host: %s has no NTP service operating on it", host))
			return false, failures, fmt.Errorf("host: %s has no NTP service operating on it", host)
		}

		res.Current, err = s.Query(ctx)
		if err != nil {
			return false, nil, err
		}

		res.ClientStatus = service.Policy(*res.Service)
		res.ServiceStatus = service.Status(*res.Service)
		res.HostName = host
		res.NtpServers = res.servers()

		hostsDateInfo = append(hostsDateInfo, *res)
	}

	for _, dateInfo := range hostsDateInfo {
		if dateInfo.ClientStatus != "Enabled" {
			failureMsg := fmt.Sprintf("NTP client status is disabled or unknown for host: %s", dateInfo.HostName)
			failures = append(failures, failureMsg)
		}

		if dateInfo.ServiceStatus != "Running" {
			failureMsg := fmt.Sprintf("NTP service status is stopped or unknown for host: %s", dateInfo.HostName)
			failures = append(failures, failureMsg)
		}
	}

	err := validateHostNTPServers(hostsDateInfo)
	if err != nil {
		failures = append(failures, fmt.Sprintf("%s", err.Error()))
	}

	if len(failures) > 0 {
		return false, failures, err
	}

	return true, failures, nil
}

func validateHostNTPServers(hostsDateInfo []HostDateInfo) error {
	var intersectionList []string
	for i := 0; i < len(hostsDateInfo)-1; i++ {
		if intersectionList == nil {
			intersectionList = intersection(hostsDateInfo[i].NtpServers, hostsDateInfo[i+1].NtpServers)
		} else {
			intersectionList = intersection(intersectionList, hostsDateInfo[i+1].NtpServers)
		}

		if intersectionList == nil {
			return fmt.Errorf("some of the hosts has differently configured NTP servers")
		}
	}

	return nil
}

func intersection(listA []string, listB []string) []string {
	var intersect []string
	for _, element := range listA {
		if slices.Contains(listB, element) {
			intersect = append(intersect, element)
		}
	}

	if len(intersect) == 0 {
		return nil
	}
	return intersect
}
