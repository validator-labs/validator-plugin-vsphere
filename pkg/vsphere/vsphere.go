package vsphere

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/govc/host/service"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/session/keepalive"
	ssoadmintypes "github.com/vmware/govmomi/ssoadmin/types"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/exp/slices"
)

const (
	KeepAliveIntervalInMinute = 10
	DatacenterTagCategory     = "k8s-region"
	ComputeClusterTagCategory = "k8s-zone"
)

var (
	sessionCache        = map[string]Session{}
	sessionMU           sync.Mutex
	restClientLoggedOut = false
)

func NewVSphereDriver(VCenterServer, VCenterUsername, VCenterPassword, datacenter string) (*VSphereCloudDriver, error) {
	session, err := GetOrCreateSession(context.TODO(), VCenterServer, VCenterUsername, VCenterPassword, true)
	if err != nil {
		return nil, err
	}

	return &VSphereCloudDriver{
		VCenterServer:   VCenterServer,
		VCenterUsername: VCenterUsername,
		VCenterPassword: VCenterPassword,
		Datacenter:      datacenter,
		Client:          session.GovmomiClient,
		RestClient:      session.RestClient,
	}, nil
}

func (v *VSphereCloudDriver) GetCurrentVmwareUser(ctx context.Context) (string, error) {
	userSession, err := v.Client.SessionManager.UserSession(ctx)
	if err != nil {
		return "", err
	}

	return userSession.UserName, nil
}

func (v *VSphereCloudDriver) GetVSphereVMFolders(ctx context.Context, datacenter string) ([]string, error) {
	finder, dc, err := v.getFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return nil, err
	}

	fos, err := finder.FolderList(ctx, "*")
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to fetch vSphere folders for Datacenter %s", datacenter))
	}

	prefix := fmt.Sprintf("/%s/vm/", dc)
	folders := make([]string, 0)
	for _, fo := range fos {
		inventoryPath := fo.InventoryPath
		//get vm folders, items with path prefix '/{Datacenter}/vm'
		if strings.HasPrefix(inventoryPath, prefix) {
			folder := strings.TrimPrefix(inventoryPath, prefix)
			//skip spectro folders & sub-folders
			if !strings.HasPrefix(folder, "spc-") &&
				!strings.Contains(folder, "/spc-") {
				folders = append(folders, folder)
			}
		}
	}

	sort.Strings(folders)
	return folders, nil
}

func (v *VSphereCloudDriver) GetVSphereDatacenters(ctx context.Context) ([]string, error) {
	finder, err := v.getFinder()
	if err != nil {
		return nil, err
	}

	dcs, err := finder.DatacenterList(ctx, "*")
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch vSphere datacenters")
	}

	if len(dcs) == 0 {
		return nil, errors.New("No datacenters found")
	}

	client := dcs[0].Client()
	tags, categoryId, err := v.getTagsAndCategory(ctx, client, "Datacenter", DatacenterTagCategory)
	if err != nil {
		return nil, err
	}

	datacenters := make([]string, 0)
	for _, dc := range dcs {
		if v.ifTagHasCategory(tags[dc.Reference().Value].Tags, categoryId) {
			dcName := strings.TrimPrefix(dc.InventoryPath, "/")
			datacenters = append(datacenters, dcName)
		}
	}

	if len(datacenters) == 0 {
		return nil, errors.Errorf("No datacenter with tag category %s found", DatacenterTagCategory)
	}

	sort.Strings(datacenters)
	return datacenters, nil
}

func (v *VSphereCloudDriver) GetVSphereClusters(ctx context.Context, datacenter string) ([]string, error) {
	finder, dc, err := v.getFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return nil, err
	}

	ccrs, err := finder.ClusterComputeResourceList(ctx, "*")
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch vSphere clusters")
	}

	if len(ccrs) == 0 {
		return nil, errors.New("No compute clusters found")
	}

	client := ccrs[0].Client()

	tags, categoryId, err := v.getTagsAndCategory(ctx, client, "ClusterComputeResource", ComputeClusterTagCategory)
	if err != nil {
		return nil, err
	}

	clusters := make([]string, 0)
	for _, ccr := range ccrs {
		if v.ifTagHasCategory(tags[ccr.Reference().Value].Tags, categoryId) {
			prefix := fmt.Sprintf("/%s/host/", dc)
			cluster := strings.TrimPrefix(ccr.InventoryPath, prefix)
			clusters = append(clusters, cluster)
		}
	}

	if len(clusters) == 0 {
		return nil, errors.Errorf("No compute clusters with tag category %s found", ComputeClusterTagCategory)
	}

	sort.Strings(clusters)
	return clusters, nil
}

func (v *VSphereCloudDriver) GetVSphereHostSystems(ctx context.Context, datacenter, cluster string) ([]VSphereHostSystem, error) {
	finder, _, err := v.getFinderWithDatacenter(ctx, datacenter)
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

func (v *VSphereCloudDriver) IsValidVSphereCredentials(ctx context.Context) (bool, error) {
	_, err := v.getFinder()
	if err != nil {
		return false, err
	}

	return true, nil
}

func (v *VSphereCloudDriver) ValidateVsphereVersion(constraint string) error {
	vsphereVersion := v.Client.ServiceContent.About.Version
	vn, err := version.NewVersion(vsphereVersion)
	if err != nil {
		return err
	}
	constraints, err := version.NewConstraint(constraint)
	if err != nil {
		return err
	}
	if !constraints.Check(vn) {
		return fmt.Errorf("vSphere version %s does not satisfies the constraints %s", vsphereVersion, constraints)
	}
	return nil
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

func (v *VSphereCloudDriver) GetResourcePools(ctx context.Context, datacenter string, cluster string) ([]*object.ResourcePool, error) {
	path := fmt.Sprintf("/%s/host/%s/Resources/*", datacenter, cluster)

	if cluster == "" {
		path = fmt.Sprintf("/%s/host/*", datacenter)
	}

	rps, err := v.getResourcePools(ctx, datacenter, path)
	if err != nil {
		return nil, err
	}

	return rps, nil
}

func (v *VSphereCloudDriver) GetVapps(ctx context.Context) ([]mo.VirtualApp, error) {
	m := view.NewManager(v.Client.Client)

	containerView, err := m.CreateContainerView(ctx, v.Client.Client.ServiceContent.RootFolder, []string{"VirtualApp"}, true)
	if err != nil {
		return nil, err
	}
	var vApps []mo.VirtualApp
	err = containerView.Retrieve(ctx, []string{"VirtualApp"}, nil, &vApps)
	if err != nil {
		return nil, err
	}

	return vApps, nil
}

func (v *VSphereCloudDriver) GetResourceTags(ctx context.Context, resourceType string) (map[string]tags.AttachedTags, error) {
	tags, err := v.getResourceTags(ctx, v.Client.Client, resourceType)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

func (v *VSphereCloudDriver) ValidateUserPrivilegeOnEntities(ctx context.Context, authManager *object.AuthorizationManager, datacenter string, finder *find.Finder, entityName, entityType string, privileges []string, userName, clusterName string) (isValid bool, failures []string, err error) {
	var folder *object.Folder
	var cluster *object.ClusterComputeResource
	var host *object.HostSystem
	var vapp *object.VirtualApp
	var resourcePool *object.ResourcePool
	var vm *object.VirtualMachine

	var moID types.ManagedObjectReference

	switch entityType {
	case "folder":
		_, folder, err = v.GetFolderIfExists(ctx, finder, datacenter, entityName)
		if err != nil {
			return false, failures, err
		}
		moID = folder.Reference()
	case "resourcepool":
		_, resourcePool, err = v.GetResourcePoolIfExists(ctx, finder, datacenter, clusterName, entityName)
		if err != nil {
			return false, failures, err
		}
		moID = resourcePool.Reference()
	case "vapp":
		_, vapp, err = v.GetVAppIfExists(ctx, finder, datacenter, entityName)
		if err != nil {
			return false, failures, err
		}
		moID = vapp.Reference()
	case "vm":
		_, vm, err = v.GetVMIfExists(ctx, finder, datacenter, clusterName, entityName)
		if err != nil {
			return false, failures, err
		}
		moID = vm.Reference()
	case "host":
		_, host, err = v.GetHostIfExists(ctx, finder, datacenter, clusterName, entityName)
		if err != nil {
			return false, failures, err
		}
		moID = host.Reference()
	case "cluster":
		_, cluster, err = v.GetClusterIfExists(ctx, finder, datacenter, entityName)
		if err != nil {
			return false, failures, err
		}
		moID = cluster.Reference()
	}

	userPrincipal := getUserPrincipalFromUsername(userName)
	privilegeResult, err := authManager.FetchUserPrivilegeOnEntities(ctx, []types.ManagedObjectReference{moID}, userPrincipal)
	if err != nil {
		return false, failures, err
	}

	privilegesMap := make(map[string]bool)
	for _, result := range privilegeResult {
		for _, privilege := range result.Privileges {
			privilegesMap[privilege] = true
		}
	}

	for _, privilege := range privileges {
		if _, ok := privilegesMap[privilege]; !ok {
			err = fmt.Errorf("some entity privileges were not found for user: %s", userName)
			failures = append(failures, fmt.Sprintf("user: %s does not have privilege: %s on entity type: %s with name: %s", userName, privilege, entityType, entityName))
		}
	}

	if len(failures) == 0 {
		isValid = true
	}

	return isValid, failures, nil
}

func GetOrCreateSession(
	ctx context.Context,
	server, username, password string, refreshRestClient bool) (Session, error) {

	sessionMU.Lock()
	defer sessionMU.Unlock()

	sessionKey := server + username
	currentSession, ok := sessionCache[sessionKey]

	if ok {
		if refreshRestClient && restClientLoggedOut {
			//Rest Client
			restClient, err := createRestClientWithKeepAlive(ctx, username, password, currentSession.GovmomiClient)
			if err != nil {
				return currentSession, err
			}
			currentSession.RestClient = restClient
			restClientLoggedOut = false
		}
		return currentSession, nil
	}

	// govmomi Client
	govClient, err := createGovmomiClientWithKeepAlive(ctx, sessionKey, server, username, password)
	if err != nil {
		return currentSession, err
	}

	//Rest Client
	restClient, err := createRestClientWithKeepAlive(ctx, username, password, govClient)
	if err != nil {
		return currentSession, err
	}

	currentSession.GovmomiClient = govClient
	currentSession.RestClient = restClient

	// Cache the currentSession.
	sessionCache[sessionKey] = currentSession
	return currentSession, nil
}

func createGovmomiClientWithKeepAlive(ctx context.Context, sessionKey, server, username, password string) (*govmomi.Client, error) {
	//get vcenter URL
	vCenterURL, err := getVCenterUrl(server, username, password)
	if err != nil {
		return nil, err
	}

	insecure := true

	soapClient := soap.NewClient(vCenterURL, insecure)
	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, err
	}

	vimClient.UserAgent = "spectro-palette"

	c := &govmomi.Client{
		Client:         vimClient,
		SessionManager: session.NewManager(vimClient),
	}

	send := func() error {
		ctx := context.Background()
		_, err := methods.GetCurrentTime(ctx, vimClient.RoundTripper)
		if err != nil {
			ClearCache(sessionKey)
		}
		return err
	}

	// this starts the keep alive handler when Login is called, and stops the handler when Logout is called
	// it'll also stop the handler when send() returns error, so we wrap around the default send()
	// with err check to clear cache in case of error
	vimClient.RoundTripper = keepalive.NewHandlerSOAP(vimClient.RoundTripper, KeepAliveIntervalInMinute*time.Minute, send)

	// Only login if the URL contains user information.
	if vCenterURL.User != nil {
		err = c.Login(ctx, vCenterURL.User)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func getVCenterUrl(vCenterServer string, vCenterUsername string, vCenterPassword string) (*url.URL, error) {
	// parse vcenter URL
	for _, scheme := range []string{"http://", "https://"} {
		vCenterServer = strings.TrimPrefix(vCenterServer, scheme)
	}
	vCenterServer = fmt.Sprintf("https://%s/sdk", strings.TrimSuffix(vCenterServer, "/"))

	vCenterURL, err := url.Parse(vCenterServer)
	if err != nil {
		return nil, errors.Errorf("invalid vCenter server")

	}
	vCenterURL.User = url.UserPassword(vCenterUsername, vCenterPassword)

	return vCenterURL, nil
}

func createRestClientWithKeepAlive(ctx context.Context, username, password string, govClient *govmomi.Client) (*rest.Client, error) {
	// create RestClient for operations like get tags
	restClient := rest.NewClient(govClient.Client)

	err := restClient.Login(ctx, url.UserPassword(username, password))
	if err != nil {
		return nil, err
	}

	return restClient, nil
}

func ClearCache(sessionKey string) {
	sessionMU.Lock()
	defer sessionMU.Unlock()
	delete(sessionCache, sessionKey)
}

func (v *VSphereCloudDriver) CreateVSphereVMFolder(ctx context.Context, datacenter string, folders []string) error {
	finder, _, err := v.getFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return err
	}

	for _, folder := range folders {
		folderExists, _, err := v.GetFolderIfExists(ctx, finder, datacenter, folder)
		if folderExists {
			continue
		}

		dir := path.Dir(folder)
		name := path.Base(folder)

		if dir == "" {
			dir = "/"
		}

		folder, err := finder.Folder(ctx, dir)
		if err != nil {
			return fmt.Errorf("error fetching folder: %s. Code:%d", err.Error(), http.StatusInternalServerError)
		}

		if _, err := folder.CreateFolder(ctx, name); err != nil {
			return fmt.Errorf("error creating folder: %s. Code:%d", err.Error(), http.StatusInternalServerError)
		}
	}

	return nil
}

func (v *VSphereCloudDriver) getFinderWithDatacenter(ctx context.Context, datacenter string) (*find.Finder, string, error) {
	finder, err := v.getFinder()
	if err != nil {
		return nil, "", err
	}
	dc, govErr := finder.DatacenterOrDefault(ctx, datacenter)
	if govErr != nil {
		return nil, "", fmt.Errorf("failed to fetch datacenter: %s. code: %d", govErr.Error(), http.StatusBadRequest)
	}
	//set the datacenter
	finder.SetDatacenter(dc)

	return finder, dc.Name(), nil
}

func (v *VSphereCloudDriver) getFinder() (*find.Finder, error) {
	if v.Client == nil {
		return nil, fmt.Errorf("failed to fetch govmomi client: %d", http.StatusBadRequest)
	}

	finder := find.NewFinder(v.Client.Client, true)
	return finder, nil
}

func (v *VSphereCloudDriver) getHostSystems(ctx context.Context, v1 *view.ContainerView) ([]mo.HostSystem, error) {
	var hs []mo.HostSystem
	e := v1.Retrieve(ctx, []string{"HostSystem"}, []string{"summary", "name", "parent"}, &hs)
	if e != nil {
		return nil, errors.Wrap(e, "failed to get host systems")
	}
	return hs, nil
}

func (v *VSphereCloudDriver) getTagsAndCategory(ctx context.Context, client *vim25.Client, resourceType, tagCategory string) (map[string]tags.AttachedTags, string, error) {
	categoryId, e := v.getCategoryId(ctx, client, tagCategory)
	if e != nil {
		return nil, "", e
	}

	if categoryId == "" {
		return nil, "", errors.Errorf("No tag with category type %s is created", tagCategory)
	}

	tags, e := v.getResourceTags(ctx, client, resourceType)
	if e != nil {
		return nil, "", e
	}
	if len(tags) == 0 {
		return nil, "", errors.Errorf("No tag is attached to resource %s", resourceType)
	}

	return tags, categoryId, e
}

func (v *VSphereCloudDriver) getResourceTags(ctx context.Context, client *vim25.Client, resourceType string) (map[string]tags.AttachedTags, error) {
	t, err := v.getTagManager(ctx, client)
	if err != nil {
		return nil, err
	}
	m, err := view.NewManager(client).CreateContainerView(ctx, client.ServiceContent.RootFolder, []string{resourceType}, true)
	if err != nil {
		return nil, err
	}

	resource, err := m.Find(ctx, []string{resourceType}, property.Match{})
	if err != nil {
		return nil, err
	}

	refs := make([]mo.Reference, len(resource))
	for i := range resource {
		refs[i] = resource[i]
	}
	attachedTags, err := t.GetAttachedTagsOnObjects(ctx, refs)
	if err != nil {
		return nil, err
	}

	tags := make(map[string]tags.AttachedTags)
	for _, t := range attachedTags {
		tags[t.ObjectID.Reference().Value] = t
	}
	return tags, nil
}

func (v *VSphereCloudDriver) getCategoryId(ctx context.Context, client *vim25.Client, name string) (string, error) {

	t, err := v.getTagManager(ctx, client)
	if err != nil {
		return "", err
	}
	categories, err := t.GetCategories(ctx)
	if err != nil {
		return "", err
	}
	for _, category := range categories {
		if category.Name == name {
			return category.ID, nil
		}
	}
	return "", nil
}

func (v *VSphereCloudDriver) ifTagHasCategory(tags []tags.Tag, categoryId string) bool {
	for _, tag := range tags {
		if tag.CategoryID == categoryId {
			return true
		}
	}
	return false
}

func (v *VSphereCloudDriver) getTagManager(ctx context.Context, client *vim25.Client) (*tags.Manager, error) {
	c := rest.NewClient(client)
	err := c.Login(ctx, url.UserPassword(v.VCenterUsername, v.VCenterPassword))
	if err != nil {
		return nil, err
	}

	return tags.NewManager(c), nil
}

func (v *VSphereCloudDriver) getVmClient(ctx context.Context, dcName string) (*find.Finder, *view.ContainerView, *vim25.Client, error) {
	finder, _, err := v.getFinderWithDatacenter(ctx, dcName)
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

func getNameFromInventory(inventoryPath string) string {
	arr := strings.Split(inventoryPath, "/")
	return arr[len(arr)-1]
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

func getVmCluster(clusterName string, clusters []*object.ClusterComputeResource) *object.ClusterComputeResource {
	for _, cluster := range clusters {
		if cluster.ComputeResource.Reference().Value == clusterName {
			return cluster
		}
	}
	return nil
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

func (v *VSphereCloudDriver) getResourcePools(ctx context.Context, datacenter, path string) ([]*object.ResourcePool, error) {
	finder, _, err := v.getFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get finder with datacenter")
	}

	rps, err := finder.ResourcePoolList(ctx, path)
	if err != nil {
		return nil, err
	}

	return rps, nil
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

	defer v1.Destroy(ctx)

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

func (v *VSphereCloudDriver) FolderExists(ctx context.Context, finder *find.Finder, datacenter, folderName string) (bool, error) {

	if _, err := finder.Folder(ctx, folderName); err != nil {
		return false, nil
	}
	return true, nil
}

func (v *VSphereCloudDriver) GetFolderNameByID(ctx context.Context, datacenter, id string) (string, error) {
	finder, dc, err := v.getFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return "", err
	}

	fos, govErr := finder.FolderList(ctx, "*")
	if govErr != nil {
		return "", fmt.Errorf("failed to fetch vSphere folders. Datacenter: %s, Error: %s", datacenter, govErr.Error())
	}

	prefix := fmt.Sprintf("/%s/vm/", dc)
	for _, fo := range fos {
		inventoryPath := fo.InventoryPath
		//get vm folders, items with path prefix '/{Datacenter}/vm'
		if strings.HasPrefix(inventoryPath, prefix) {
			folderName := strings.TrimPrefix(inventoryPath, prefix)
			//skip spectro folders & sub-folders
			if !strings.HasPrefix(folderName, "spc-") && !strings.Contains(folderName, "/spc-") {
				if fo.Reference().Value == id {
					return folderName, nil
				}
			}
		}
	}

	return "", fmt.Errorf("unable to find folder with id: %s", id)
}

func (v *VSphereCloudDriver) GetFinderWithDatacenter(ctx context.Context, datacenter string) (*find.Finder, string, error) {
	finder, err := v.getFinder()
	if err != nil {
		return nil, "", err
	}
	dc, govErr := finder.DatacenterOrDefault(ctx, datacenter)
	if govErr != nil {
		return nil, "", fmt.Errorf("failed to fetch datacenter: %s. code: %s"+govErr.Error(), http.StatusBadRequest)
	}
	//set the datacenter
	finder.SetDatacenter(dc)

	return finder, dc.Name(), nil
}

func GetVmwareUserPrivileges(ctx context.Context, userPrincipal string, groupPrincipals []string, authManager *object.AuthorizationManager) (map[string]bool, error) {
	groupPrincipalMap := make(map[string]bool)
	for _, principal := range groupPrincipals {
		groupPrincipalMap[principal] = true
	}

	// Get the current user's roles
	authRoles, err := authManager.RoleList(ctx)
	if err != nil {
		return nil, err
	}

	// create a map to store privileges for current user
	privileges := make(map[string]bool)

	// Print the roles
	for _, authRole := range authRoles {
		// print permissions for every role
		permissions, err := authManager.RetrieveRolePermissions(ctx, authRole.RoleId)
		if err != nil {
			return nil, err
		}
		for _, perm := range permissions {
			if perm.Principal == userPrincipal || groupPrincipalMap[perm.Principal] {
				for _, priv := range authRole.Privilege {
					privileges[priv] = true
				}
			}
		}
	}
	return privileges, nil
}

func (v *VSphereCloudDriver) GetVSphereResourcePools(ctx context.Context, datacenter string, cluster string) (resourcePools []string, err error) {
	finder, dc, err := v.getFinderWithDatacenter(ctx, datacenter)
	if err != nil {
		return nil, err
	}

	searchPath := fmt.Sprintf("/%s/host/%s/Resources/*", dc, cluster)
	pools, govErr := finder.ResourcePoolList(ctx, searchPath)
	if govErr != nil {
		//ignore NotFoundError, to allow selection of "Resources" as the default option for rs pool
		if _, ok := govErr.(*find.NotFoundError); !ok {
			return nil, fmt.Errorf("failed to fetch vSphere resource pools. datacenter: %s, code: %d", datacenter, http.StatusBadRequest)
		}
	}

	for i := 0; i < len(pools); i++ {
		pool := pools[i]
		prefix := fmt.Sprintf("/%s/host/%s/Resources/", dc, cluster)
		poolPath := strings.TrimPrefix(pool.InventoryPath, prefix)
		resourcePools = append(resourcePools, poolPath)
		childPoolSearchPath := fmt.Sprintf("/%s/host/%s/Resources/%s/*", dc, cluster, poolPath)
		childPools, err := finder.ResourcePoolList(ctx, childPoolSearchPath)
		if err == nil {
			pools = append(pools, childPools...)
		}
	}

	sort.Strings(resourcePools)
	return resourcePools, nil
}

func (v *VSphereCloudDriver) getClusterDatastores(ctx context.Context, finder *find.Finder, datacenter string, cluster mo.ClusterComputeResource) (datastores []string, err error) {
	dsMobjRefs := cluster.Datastore

	for i := range dsMobjRefs {
		inventoryPath := ""
		dsObjRef, err := finder.ObjectReference(ctx, dsMobjRefs[i])
		if err != nil {
			return nil, fmt.Errorf("error: %s, code: %d", err.Error(), http.StatusBadRequest)
		}
		if dsObjRef != nil {
			ref := dsObjRef
			switch ref.(type) {
			case *object.Datastore:
				n := dsObjRef.(*object.Datastore)
				inventoryPath = n.InventoryPath
			default:
				continue
			}

			if inventoryPath != "" {
				prefix := fmt.Sprintf("/%s/datastore/", datacenter)
				datastore := strings.TrimPrefix(inventoryPath, prefix)
				datastores = append(datastores, datastore)
			}
		}
	}

	sort.Strings(datastores)
	return datastores, nil
}

func (v *VSphereCloudDriver) getClusterComputeResources(ctx context.Context, finder *find.Finder) ([]*object.ClusterComputeResource, error) {
	ccrs, err := finder.ClusterComputeResourceList(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("failed to get compute cluster resources: %s", err.Error())
	}
	return ccrs, nil
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

func (info *HostDateInfo) servers() []string {
	return info.NtpConfig.Server
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

func getUserPrincipalFromPrincipalID(id ssoadmintypes.PrincipalId) string {
	return fmt.Sprintf("%s\\%s", strings.ToUpper(id.Domain), id.Name)
}

func getUserPrincipalFromUsername(username string) string {
	splitStr := strings.Split(username, "@")
	return fmt.Sprintf("%s\\%s", strings.ToUpper(splitStr[1]), splitStr[0])
}
