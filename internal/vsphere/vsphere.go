package vsphere

import (
	"context"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/api/v1alpha1"
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
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

const KeepAliveIntervalInMinute = 10

var sessionCache = map[string]Session{}
var sessionMU sync.Mutex
var restClientLoggedOut = false

type VMwareRolePrivilege struct {
	rule       v1alpha1.RolePrivilegeValidationRule
	Privileges map[string]bool
}

type VsphereCloudAccount struct {

	// Insecure is a flag that controls whether to validate the vSphere server's certificate.
	Insecure bool `json:"insecure"`

	// password
	// Required: true
	Password string `json:"password"`

	// username
	// Required: true
	Username string `json:"username"`

	// VcenterServer is the address of the vSphere endpoint
	// Required: true
	VcenterServer string `json:"vcenterServer"`
}

type Session struct {
	GovmomiClient *govmomi.Client
	RestClient    *rest.Client
}

type RulesEngine struct {
	Driver *VSphereCloudDriver
	Rules  []v1alpha1.RolePrivilegeValidationRule
}

type RegionZoneCategoryExistsInput struct {
	Datacenter         string
	Cluster            []string
	RegionCategoryName string
	ZoneCategoryName   string
}

type VSphereCloudDriver struct {
	VCenterServer   string
	VCenterUsername string
	VCenterPassword string
	Client          *govmomi.Client
	RestClient      *rest.Client
}

func NewVSphereDriver(logger logr.Logger, VCenterServer string, VCenterUsername string, VCenterPassword string) (*VSphereCloudDriver, error) {
	session, err := GetOrCreateSession(context.TODO(), VCenterServer, VCenterUsername, VCenterPassword, true)
	if err != nil {
		logger.V(1).Error(err, "failed to create govmomi session")
		return nil, err
	}

	return &VSphereCloudDriver{
		VCenterServer:   VCenterServer,
		VCenterUsername: VCenterUsername,
		VCenterPassword: VCenterPassword,
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

func (v *VSphereCloudDriver) GetVmwareUserPrivileges(userName string, authManager *object.AuthorizationManager) (map[string]bool, error) {
	// Get the current user's roles
	authRoles, err := authManager.RoleList(context.TODO())
	if err != nil {
		return nil, err
	}

	// create a map to store privileges for current user
	privileges := make(map[string]bool)

	// Print the roles
	for _, authRole := range authRoles {
		// print permissions for every role
		permissions, err := authManager.RetrieveRolePermissions(context.TODO(), authRole.RoleId)
		if err != nil {
			return nil, err
		}
		for _, perm := range permissions {
			// if current user has the role, append all user privileges to privileges slice.
			if perm.Principal == userName {
				for _, priv := range authRole.Privilege {
					privileges[priv] = true
				}
			}
		}
	}
	return privileges, nil
}

func ToSlice(m map[string]bool) []interface{} {
	values := make([]interface{}, 0, len(m))
	for k, v := range m {
		if v {
			values = append(values, k)
		}
	}
	return values
}

func (v *VMwareRolePrivilege) ValidateVMwareRolePrivilege() bool {
	data := map[string]interface{}{
		"vmware_user_privileges": ToSlice(v.Privileges),
	}
	for _, expr := range v.rule.Expressions {
		expression, err := govaluate.NewEvaluableExpression(expr)
		if err != nil {
			return false
		} else {
			result, err := expression.Evaluate(data)
			if err != nil {
				return false
			} else {
				if result == false {
					return false
				}
			}
		}
	}
	return true
}

func IsValidRule(rule v1alpha1.RolePrivilegeValidationRule, privileges map[string]bool) bool {
	// convert the keys of the map to a slice of strings
	keys := make([]string, 0, len(privileges))
	for k := range privileges {
		keys = append(keys, k)
	}

	// sort the slice of keys
	sort.Strings(keys)

	if rule.IsEnabled {
		switch rule.RuleType {
		case "VMwareRolePrivilege":
			rolePrivilegeRule := VMwareRolePrivilege{}
			rolePrivilegeRule.rule = rule
			rolePrivilegeRule.Privileges = privileges
			return rolePrivilegeRule.ValidateVMwareRolePrivilege()
		}
	}

	return false
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
			restClient, err := createRestClientWithKeepAlive(ctx, sessionKey, username, password, currentSession.GovmomiClient)
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
	restClient, err := createRestClientWithKeepAlive(ctx, sessionKey, username, password, govClient)
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

func createRestClientWithKeepAlive(ctx context.Context, sessionKey, username, password string, govClient *govmomi.Client) (*rest.Client, error) {
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

func RegionZoneCategoryExists(tagsManager *tags.Manager, finder *find.Finder, input RegionZoneCategoryExistsInput) (*bool, error) {
	isTrue, isFalse := true, false
	regionCategoryID, zoneCategoryID := "", ""

	cats, err := tagsManager.GetCategories(context.TODO())
	if err != nil {
		return &isFalse, err
	}
	var regionZoneTags []tags.Category
	for _, category := range cats {
		switch category.Name {
		case input.RegionCategoryName:
			regionCategoryID = category.ID
			regionZoneTags = append(regionZoneTags, category)
		case input.ZoneCategoryName:
			zoneCategoryID = category.ID
			regionZoneTags = append(regionZoneTags, category)
		}
	}

	if len(regionZoneTags) < 2 {
		return &isFalse, nil
	}

	// check if datacenter has region tag
	list, err := finder.ManagedObjectList(context.TODO(), fmt.Sprintf("/%s", input.Datacenter))
	if err != nil {
		return nil, err
	}

	// return early if no can't find the managedobject list
	if len(list) == 0 {
		return nil, nil
	}
	var refs []mo.Reference
	refs = append(refs, list[0].Object.Reference())
	attachedTags, err := tagsManager.GetAttachedTagsOnObjects(context.TODO(), refs)
	if err != nil {
		return nil, err
	}
	isDatacenterTaggedWithRegion := false

	for _, attachedTag := range attachedTags {
		for _, tagName := range attachedTag.Tags {
			if tagName.CategoryID == regionCategoryID {
				isDatacenterTaggedWithRegion = true
				break
			}
		}
	}

	// check if all compute clusters has zone tag
	areComputeClustersTaggedWithZone := true
	for _, cluster := range input.Cluster {
		list, err = finder.ManagedObjectList(context.TODO(), fmt.Sprintf("/%s/host/%s", input.Datacenter, cluster))
		if err != nil {
			return nil, err
		}
		// return early if no can't find the managedobject list
		if len(list) == 0 {
			return nil, nil
		}
		refs = nil
		refs = append(refs, list[0].Object.Reference())
		attachedTags, err := tagsManager.GetAttachedTagsOnObjects(context.TODO(), refs)
		if err != nil {
			return nil, err
		}
		found := false
		for _, tag := range attachedTags {
			if found {
				break
			}
			for _, tagName := range tag.Tags {
				if tagName.CategoryID == zoneCategoryID {
					found = true
					break
				}
			}
		}
		areComputeClustersTaggedWithZone = areComputeClustersTaggedWithZone && found
	}

	if areComputeClustersTaggedWithZone && isDatacenterTaggedWithRegion && len(regionZoneTags) >= 2 {
		return &isTrue, nil
	}

	return &isFalse, nil
}
