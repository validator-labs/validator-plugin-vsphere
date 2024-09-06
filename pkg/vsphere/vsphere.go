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
	// DatacenterTagCategory is the tag category for datacenter
	DatacenterTagCategory = "k8s-region"
	// ComputeClusterTagCategory is the tag category for compute cluster
	ComputeClusterTagCategory = "k8s-zone"
)

var (
	sessionCache        = map[string]Session{}
	sessionMU           sync.Mutex
	restClientLoggedOut = false
)

// Driver is an interface that defines the functions to interact with vSphere
type Driver interface {
	GetVSphereVMFolders(ctx context.Context, datacenter string) ([]string, error)
	GetVSphereDatacenters(ctx context.Context) ([]string, error)
	GetVSphereClusters(ctx context.Context, datacenter string) ([]string, error)
	GetVSphereHostSystems(ctx context.Context, datacenter, cluster string) ([]HostSystem, error)
	IsValidVSphereCredentials() (bool, error)
	ValidateVsphereVersion(constraint string) error
	GetHostClusterMapping(ctx context.Context) (map[string]string, error)
	GetVSphereVms(ctx context.Context, dcName string) ([]VM, error)
	GetResourcePools(ctx context.Context, datacenter string, cluster string) ([]*object.ResourcePool, error)
	GetVapps(ctx context.Context) ([]mo.VirtualApp, error)
	GetResourceTags(ctx context.Context, resourceType string) (map[string]tags.AttachedTags, error)
}

// ensure that CloudDriver implements the Driver interface
var _ Driver = &CloudDriver{}

// CloudDriver is a struct that implements the Driver interface
type CloudDriver struct {
	VCenterServer   string
	VCenterUsername string
	VCenterPassword string
	Datacenter      string
	Client          *govmomi.Client
	RestClient      *rest.Client
	log             logr.Logger
}

// Session is a struct that contains the govmomi and rest clients
type Session struct {
	GovmomiClient *govmomi.Client
	RestClient    *rest.Client
}

// NewVSphereDriver creates a new instance of CloudDriver
func NewVSphereDriver(account vcenter.Account, datacenter string, log logr.Logger) (*CloudDriver, error) {
	session, err := GetOrCreateSession(context.TODO(), account.Host, account.Username, account.Password, true)
	if err != nil {
		return nil, err
	}

	return &CloudDriver{
		VCenterServer:   account.Host,
		VCenterUsername: account.Username,
		VCenterPassword: account.Password,
		Datacenter:      datacenter,
		Client:          session.GovmomiClient,
		RestClient:      session.RestClient,
		log:             log,
	}, nil
}

// IsValidVSphereCredentials checks if the vSphere credentials are valid
func (v *CloudDriver) IsValidVSphereCredentials() (bool, error) {
	_, err := v.getFinder()
	if err != nil {
		return false, err
	}

	return true, nil
}

// ValidateVsphereVersion validates the vSphere version satisfies the given constraint
func (v *CloudDriver) ValidateVsphereVersion(constraint string) error {
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

// GetFinderWithDatacenter returns a finder and the datacenter name
func (v *CloudDriver) GetFinderWithDatacenter(ctx context.Context, datacenter string) (*find.Finder, string, error) {
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

func (v *CloudDriver) getFinder() (*find.Finder, error) {
	if v.Client == nil {
		return nil, fmt.Errorf("failed to fetch govmomi client: %d", http.StatusBadRequest)
	}

	finder := find.NewFinder(v.Client.Client, true)
	return finder, nil
}

// GetOrCreateSession returns the session for the given server, username and password
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
	vCenterURL, err := getVCenterURL(server, username, password)
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

func getVCenterURL(vCenterServer string, vCenterUsername string, vCenterPassword string) (*url.URL, error) {
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

// ClearCache deletes the session from the session cache
func ClearCache(sessionKey string) {
	sessionMU.Lock()
	defer sessionMU.Unlock()
	delete(sessionCache, sessionKey)
}
