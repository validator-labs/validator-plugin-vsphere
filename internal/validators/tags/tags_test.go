package tags

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/api/v1alpha1"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/internal/vcsim"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/internal/vsphere"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vapi/rest"
	vtags "github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"k8s.io/klog/v2"
	"net/url"
	"sync"
	"testing"
)

//func TestTagsValidationService_ReconcileRegionZoneTagRules(t *testing.T) {
//	vcSim := vcsim.NewVCSim("admin@vsphere.local")
//
//	vcSim.Start()
//
//	var Log logr.Logger
//	rule := v1alpha1.RegionZoneValidationRule{
//		RegionCategoryName: "k8s-region",
//		ZoneCategoryName:   "k8s-zone",
//		Datacenter:         "DC0",
//		Clusters:           []string{"DC0_C0"},
//	}
//
//	simulator.Test(func(ctx context.Context, client *vim25.Client) {
//		c := rest.NewClient(client)
//		_ = c.Login(ctx, url.UserPassword(vcSim.Driver.VCenterUsername, vcSim.Driver.VCenterPassword))
//
//		m := tags.NewManager(c)
//
//		categoryName := "my-category"
//		categoryID, err := m.CreateCategory(ctx, &tags.Category{
//			AssociableTypes: []string{"Datacenter"},
//			Cardinality:     "SINGLE",
//			Name:            categoryName,
//		})
//		if err != nil {
//			t.Fatal(err)
//		}
//		tagName := "k8s-region"
//		tagID, err := m.CreateTag(ctx, &tags.Tag{CategoryID: categoryID, Name: tagName})
//		if err != nil {
//			t.Fatal(err)
//		}
//
//		dc, err := find.NewFinder(client).Datacenter(ctx, "DC0")
//		if err != nil {
//			t.Fatal(err)
//		}
//
//		err = m.AttachTag(ctx, tagID, dc.Reference())
//		if err != nil {
//			t.Fatal(err)
//		}
//
//		validationService := NewTagsValidationService(Log, vcSim.Driver)
//		t.Log(rule, validationService)
//		_, err = validationService.ReconcileRegionZoneTagRules(rule)
//		if err != nil {
//			t.Fatal(err)
//		}
//	})
//
//}

func initSimulator(t *testing.T) (*simulator.Model, *Session, *simulator.Server) {
	model := simulator.VPX()
	model.Host = 0
	err := model.Create()
	if err != nil {
		t.Fatal(err)
	}
	model.Service.TLS = new(tls.Config)
	model.Service.RegisterEndpoints = true

	server := model.Service.NewServer()
	pass, _ := server.URL.User.Password()

	authSession, err := GetOrCreate(
		context.TODO(),
		server.URL.Host, "",
		server.URL.User.Username(), pass, true)
	if err != nil {
		t.Fatal(err)
	}

	// create folder
	folders, err := authSession.Datacenter.Folders(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	_, err = folders.VmFolder.CreateFolder(context.TODO(), "custom-folder")
	if err != nil {
		t.Fatal(err)
	}

	return model, authSession, server
}

type datacenter struct {
	context.Context
	Ref types.ManagedObjectReference
	Obj *object.VirtualMachine
}

var categories = []vtags.Category{
	{
		ID:              "urn:vmomi:InventoryServiceCategory:552dfe88-38ab-4c76-8791-14a2156a5f3f:GLOBAL",
		Name:            "k8s-region",
		Description:     "",
		Cardinality:     "SINGLE",
		AssociableTypes: []string{"Datacenter", "Folder"},
		UsedBy:          []string{},
	},
	{
		ID:              "urn:vmomi:InventoryServiceCategory:167242af-7e93-41ed-8704-52791115e1a8:GLOBAL",
		Name:            "k8s-zone",
		Description:     "",
		Cardinality:     "SINGLE",
		AssociableTypes: []string{"Datacenter", "ClusterComputeResource", "HostSystem", "Folder"},
		UsedBy:          []string{},
	},
	{
		ID:              "urn:vmomi:InventoryServiceCategory:4adb4e4b-8aee-4beb-8f6c-66d22d768cbc:GLOBAL",
		Name:            "AVICLUSTER_UUID",
		Description:     "",
		Cardinality:     "SINGLE",
		AssociableTypes: []string{"com.vmware.content.library.Item"},
		UsedBy:          []string{},
	},
}

func TestCheckAttachedTag(t *testing.T) {
	model, session, server := initSimulator(t)
	defer model.Remove()
	defer server.Close()

	vcSim := vcsim.NewVCSim("admin@vsphere.local")
	vcSim.Start()
	vsphereCloudAccount := vcSim.GetTestVsphereAccount()

	vsphereCloudDriver, err := vsphere.NewVSphereDriver(vsphereCloudAccount.VcenterServer, vsphereCloudAccount.Username, vsphereCloudAccount.Password, "DC0")
	if err != nil {
		return
	}

	rule := v1alpha1.RegionZoneValidationRule{
		RegionCategoryName: "k8s-region",
		ZoneCategoryName:   "k8s-zone",
		Datacenter:         "DC0",
		Clusters:           []string{"DC0_C0"},
	}

	var log logr.Logger

	validationService := NewTagsValidationService(log)
	GetCategories = func(manager *vtags.Manager) ([]vtags.Category, error) {
		return categories, nil
	}
	managedObj := simulator.Map.Any("VirtualMachine").(*simulator.VirtualMachine)
	managedObjRef := object.NewVirtualMachine(session.Client.Client, managedObj.Reference()).Reference()

	vm := &datacenter{
		Context: context.TODO(),
		Obj:     object.NewVirtualMachine(session.Client.Client, managedObjRef),
		Ref:     managedObjRef,
	}

	tagName := "k8s-region"
	nonAttachedTagName := "nonAttachedTag"

	tagsMgr := vtags.NewManager(vsphereCloudDriver.RestClient)

	id, err := tagsMgr.CreateCategory(context.TODO(), &vtags.Category{
		AssociableTypes: []string{"Datacenter"},
		Cardinality:     "SINGLE",
		Name:            "CLUSTERID_CATEGORY",
	})
	if err != nil {
		return
	}

	_, err = tagsMgr.CreateTag(context.TODO(), &vtags.Tag{
		CategoryID: id,
		Name:       tagName,
	})
	if err != nil {
		return
	}

	if err := tagsMgr.AttachTag(context.TODO(), tagName, vm.Ref); err != nil {
		return
	}

	_, err = tagsMgr.CreateTag(context.TODO(), &vtags.Tag{
		CategoryID: id,
		Name:       nonAttachedTagName,
	})
	if err != nil {
		return
	}

	testCases := []struct {
		name    string
		findTag bool
		tagName string
	}{
		{
			name:    "Successfully find a tag",
			findTag: true,
			tagName: tagName,
		},
		{
			name:    "Return true if a tag doesn't exist",
			tagName: "non existent tag",
			findTag: true,
		},
		{
			name:    "Fail to find a tag",
			tagName: nonAttachedTagName,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := session.WithRestClient(context.TODO(), func(c *rest.Client) error {
				tagsMgr := vtags.NewManager(c)
				finder := find.NewFinder(session.Client.Client, true)

				vr, err := validationService.ReconcileRegionZoneTagRules(tagsMgr, finder, rule)
				if err != nil {
					return fmt.Errorf("Not expected error %v", err)
				}

				t.Log(vr)

				return nil
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Session is a vSphere session with a configured Finder.
type Session struct {
	*govmomi.Client
	Finder     *find.Finder
	Datacenter *object.Datacenter

	username string
	password string
}

var sessionMU sync.Mutex
var sessionCache = map[string]Session{}

func GetOrCreate(
	ctx context.Context,
	server, datacenter, username, password string, insecure bool) (*Session, error) {

	sessionMU.Lock()
	defer sessionMU.Unlock()

	sessionKey := server + username + datacenter
	if session, ok := sessionCache[sessionKey]; ok {
		if ok, _ := session.SessionManager.SessionIsActive(ctx); ok {
			return &session, nil
		}
	}

	soapURL, err := soap.ParseURL(server)
	if err != nil {
		return nil, fmt.Errorf("error parsing vSphere URL %q: %w", server, err)
	}
	if soapURL == nil {
		return nil, fmt.Errorf("error parsing vSphere URL %q", server)
	}

	soapURL.User = url.UserPassword(username, password)

	client, err := govmomi.NewClient(ctx, soapURL, insecure)
	if err != nil {
		return nil, fmt.Errorf("error setting up new vSphere SOAP client: %w", err)
	}

	session := Session{
		Client:   client,
		username: username,
		password: password,
	}

	session.UserAgent = "machineAPIvSphereProvider"
	session.Finder = find.NewFinder(session.Client.Client, false)

	dc, err := session.Finder.DatacenterOrDefault(ctx, datacenter)
	if err != nil {
		return nil, fmt.Errorf("unable to find datacenter %q: %w", datacenter, err)
	}
	session.Datacenter = dc
	session.Finder.SetDatacenter(dc)

	// Cache the session.
	sessionCache[sessionKey] = session

	return &session, nil
}

func (s *Session) WithRestClient(ctx context.Context, f func(c *rest.Client) error) error {
	c := rest.NewClient(s.Client.Client)

	user := url.UserPassword(s.username, s.password)
	if err := c.Login(ctx, user); err != nil {
		return err
	}

	defer func() {
		if err := c.Logout(ctx); err != nil {
			klog.Errorf("Failed to logout: %v", err)
		}
	}()

	return f(c)
}
