// Package vsphere is used to interact with vSphere
package vsphere

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/session/keepalive"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"

	"github.com/validator-labs/validator-plugin-vsphere/api/vcenter"
)

const (
	// KeepAliveIntervalInMinute is the interval in minutes for keep alive in the govmomi vim25 client
	KeepAliveIntervalInMinute = 10

	// K8sDatacenterTagCategory is the tag category for kubernetes-enabled datacenters
	K8sDatacenterTagCategory = "k8s-region"

	// K8sComputeClusterTagCategory is the tag category for kubernetes-enabled compute clusters
	K8sComputeClusterTagCategory = "k8s-zone"
)

var (
	sessionCache        = map[string]Session{}
	sessionMU           sync.Mutex
	restClientLoggedOut = false
)

// Driver is an interface that defines the functions to interact with vSphere
type Driver interface {
	GetClusters(ctx context.Context, datacenter string) ([]string, error)
	GetClustersByTag(ctx context.Context, datacenter, tagCategory string) ([]string, error)
	GetDatacenters(ctx context.Context) ([]string, error)
	GetDatacentersByTag(ctx context.Context, tagCategory string) ([]string, error)
	GetDatastores(ctx context.Context, datacenter string) ([]string, error)
	GetDistributedVirtualPortgroups(ctx context.Context, datacenter string) ([]string, error)
	GetDistributedVirtualSwitches(ctx context.Context, datacenter string) ([]string, error)
	GetHostClusterMapping(ctx context.Context) (map[string]string, error)
	GetHostSystems(ctx context.Context, datacenter, cluster string) ([]vcenter.HostSystem, error)
	GetNetworks(ctx context.Context, datacenter string) ([]string, error)
	GetResourcePools(ctx context.Context, datacenter string, cluster string) ([]*object.ResourcePool, error)
	GetVApps(ctx context.Context) ([]mo.VirtualApp, error)
	GetVMFolders(ctx context.Context, datacenter string) ([]string, error)
	GetVMs(ctx context.Context, dcName string) ([]vcenter.VM, error)
	GetResourceTags(ctx context.Context, resourceType string) (map[string]tags.AttachedTags, error)
	ValidateCredentials() (bool, error)
	ValidateVersion(constraint string) error
}

// ensure that VCenterDriver implements the Driver interface
var _ Driver = &VCenterDriver{}

// VCenterDriver is a struct that implements the Driver interface
type VCenterDriver struct {
	Account    vcenter.Account
	Datacenter string
	Client     *govmomi.Client
	RestClient *rest.Client
	log        logr.Logger
}

// Session is a struct that contains the govmomi and rest clients
type Session struct {
	GovmomiClient *govmomi.Client
	RestClient    *rest.Client
}

// NewVCenterDriver creates a new VCenterDriver
func NewVCenterDriver(account vcenter.Account, datacenter string, log logr.Logger) (*VCenterDriver, error) {
	session, err := GetOrCreateSession(context.TODO(), account, true)
	if err != nil {
		return nil, err
	}

	return &VCenterDriver{
		Account:    account,
		Datacenter: datacenter,
		Client:     session.GovmomiClient,
		RestClient: session.RestClient,
		log:        log,
	}, nil
}

// ValidateCredentials ensures that vCenter account credentials are valid
func (v *VCenterDriver) ValidateCredentials() (bool, error) {
	if _, err := v.getFinder(); err != nil {
		return false, err
	}
	return true, nil
}

// ValidateVersion ensures that the vSphere version satisfies the given constraint
func (v *VCenterDriver) ValidateVersion(constraint string) error {
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
		return fmt.Errorf("vSphere version %s does not satisfy the constraints: %s", vsphereVersion, constraints)
	}
	return nil
}

// GetFinderWithDatacenter returns a finder and the datacenter name
func (v *VCenterDriver) GetFinderWithDatacenter(ctx context.Context, datacenter string) (*find.Finder, string, error) {
	finder, err := v.getFinder()
	if err != nil {
		return nil, "", err
	}
	dc, govErr := finder.DatacenterOrDefault(ctx, datacenter)
	if govErr != nil {
		return nil, "", fmt.Errorf("failed to fetch datacenter: %s. code: %s"+govErr.Error(), http.StatusBadRequest)
	}
	// set the datacenter
	finder.SetDatacenter(dc)

	return finder, dc.Name(), nil
}

func (v *VCenterDriver) getFinder() (*find.Finder, error) {
	if v.Client == nil {
		return nil, fmt.Errorf("failed to fetch govmomi client: %d", http.StatusBadRequest)
	}

	finder := find.NewFinder(v.Client.Client, true)
	return finder, nil
}

// GetOrCreateSession returns the session for the given server, username and password
func GetOrCreateSession(ctx context.Context, account vcenter.Account, refreshRestClient bool) (Session, error) {
	sessionMU.Lock()
	defer sessionMU.Unlock()

	sessionKey := account.Host + account.Username
	currentSession, ok := sessionCache[sessionKey]

	if ok {
		if refreshRestClient && restClientLoggedOut {
			restClient, err := createRestClientWithKeepAlive(ctx, account, currentSession.GovmomiClient)
			if err != nil {
				return currentSession, err
			}
			currentSession.RestClient = restClient
			restClientLoggedOut = false
		}
		return currentSession, nil
	}

	// govmomi client
	govClient, err := createGovmomiClientWithKeepAlive(ctx, sessionKey, account)
	if err != nil {
		return currentSession, err
	}
	currentSession.GovmomiClient = govClient

	// REST client
	restClient, err := createRestClientWithKeepAlive(ctx, account, govClient)
	if err != nil {
		return currentSession, err
	}
	currentSession.RestClient = restClient

	// Cache the current session
	sessionCache[sessionKey] = currentSession

	return currentSession, nil
}

func createGovmomiClientWithKeepAlive(ctx context.Context, sessionKey string, account vcenter.Account) (*govmomi.Client, error) {
	// get vcenter URL
	vCenterURL, err := getVCenterURL(account)
	if err != nil {
		return nil, err
	}

	insecure := true

	soapClient := soap.NewClient(vCenterURL, insecure)
	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, err
	}

	vimClient.UserAgent = "vsphere-validator"

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

func getVCenterURL(account vcenter.Account) (*url.URL, error) {
	// parse vCenter URL
	for _, scheme := range []string{"http://", "https://"} {
		account.Host = strings.TrimPrefix(account.Host, scheme)
	}
	account.Host = fmt.Sprintf("https://%s/sdk", strings.TrimSuffix(account.Host, "/"))

	vCenterURL, err := url.Parse(account.Host)
	if err != nil {
		return nil, errors.Errorf("invalid vCenter server")

	}
	vCenterURL.User = account.Userinfo()

	return vCenterURL, nil
}

// createRestClientWithKeepAlive creates a REST client for operations like get tags
func createRestClientWithKeepAlive(ctx context.Context, account vcenter.Account, govClient *govmomi.Client) (*rest.Client, error) {
	restClient := rest.NewClient(govClient.Client)

	return restClient, restClient.Login(ctx, account.Userinfo())
}

// ClearCache deletes the session from the session cache
func ClearCache(sessionKey string) {
	sessionMU.Lock()
	defer sessionMU.Unlock()
	delete(sessionCache, sessionKey)
}
