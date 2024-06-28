package tags

import (
	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/find"
	_ "github.com/vmware/govmomi/vapi/simulator"
	vtags "github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/mo"
	v1 "k8s.io/api/core/v1"

	"github.com/validator-labs/validator-plugin-vsphere/api/v1alpha1"
	tags "github.com/validator-labs/validator-plugin-vsphere/internal/validators/tags"
	"github.com/validator-labs/validator-plugin-vsphere/internal/vcsim"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/vsphere"
	"github.com/validator-labs/validator-plugin-vsphere/tests/utils/test"
	"github.com/validator-labs/validator/pkg/types"
)

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
	{
		ID:              "urn:vmomi:InventoryServiceCategory:4adb4e4b-8aee-4beb-8f6c-66d22d76abcd:GLOBAL",
		Name:            "owner",
		Description:     "",
		Cardinality:     "SINGLE",
		AssociableTypes: []string{"com.vmware.content.library.Item"},
		UsedBy:          []string{},
	},
}
var attachedTags = []vtags.AttachedTags{
	{
		ObjectID: nil,
		TagIDs:   []string{"urn:vmomi:InventoryServiceTag:b4f0bd2c-1d62-4af6-ae41-bef79caf5f21:GLOBAL"},
		Tags: []vtags.Tag{
			{
				ID:          "urn:vmomi:InventoryServiceTag:b4f0bd2c-1d62-4af6-ae41-bef79caf5f21:GLOBAL",
				Description: "",
				Name:        "usdc",
				CategoryID:  "urn:vmomi:InventoryServiceCategory:552dfe88-38ab-4c76-8791-14a2156a5f3f:GLOBAL",
				UsedBy:      nil,
			},
		},
	},
	{
		ObjectID: nil,
		TagIDs:   []string{"urn:vmomi:InventoryServiceTag:e886a5b2-73cd-488e-85be-9c8b1bc740eb:GLOBAL"},
		Tags: []vtags.Tag{
			{
				ID:          "urn:vmomi:InventoryServiceTag:e886a5b2-73cd-488e-85be-9c8b1bc740eb:GLOBAL",
				Description: "",
				Name:        "zone1",
				CategoryID:  "urn:vmomi:InventoryServiceCategory:167242af-7e93-41ed-8704-52791115e1a8:GLOBAL",
				UsedBy:      nil,
			},
		},
	},
	{
		ObjectID: nil,
		TagIDs:   []string{"urn:vmomi:InventoryServiceTag:e886a5b2-73cd-488e-85be-9c8b1bc740eb:GLOBAL"},
		Tags: []vtags.Tag{
			{
				ID:          "urn:vmomi:InventoryServiceTag:e886a5b2-73cd-488e-85be-9c8b1bc740eb:GLOBAL",
				Description: "",
				Name:        "owner",
				CategoryID:  "urn:vmomi:InventoryServiceCategory:4adb4e4b-8aee-4beb-8f6c-66d22d76abcd:GLOBAL",
				UsedBy:      nil,
			},
		},
	},
}

func Execute() error {
	testCtx := test.NewTestContext()
	return test.Flow(testCtx).
		Test(NewTagValidationTest("vali8or-plugin-tags-integration-test")).
		TearDown().Audit()
}

type TagValidationTest struct {
	*test.BaseTest
	log logr.Logger
}

func NewTagValidationTest(description string) *TagValidationTest {
	return &TagValidationTest{
		log:      logr.Logger{},
		BaseTest: test.NewBaseTest("vsphere-plugin", description, nil),
	}
}

func (t *TagValidationTest) Execute(ctx *test.TestContext) (tr *test.TestResult) {
	t.log.Info("Executing Test", "name", t.GetName(), "description", t.GetDescription())
	if tr := t.PreRequisite(ctx); tr.IsFailed() {
		return tr
	}

	if result := t.testTagsOnObjects(ctx); result.IsFailed() {
		return result
	}

	return test.Success()
}

func (t *TagValidationTest) testTagsOnObjects(ctx *test.TestContext) (tr *test.TestResult) {
	vcSim := ctx.Get("vcsim")
	vsphereCloudAccount := vcSim.(*vcsim.VCSimulator).GetTestVsphereAccount()

	vsphereCloudDriver, err := vsphere.NewVSphereDriver(
		vsphereCloudAccount.VcenterServer, vsphereCloudAccount.Username,
		vsphereCloudAccount.Password, "DC0", t.log,
	)
	if err != nil {
		return tr
	}

	tm := vtags.NewManager(vsphereCloudDriver.RestClient)
	finder := find.NewFinder(vsphereCloudDriver.Client.Client)

	var log logr.Logger
	tagService := tags.NewTagsValidationService(log)

	rules := []v1alpha1.TagValidationRule{
		{
			Name:       "Datacenter validation rule",
			EntityType: "Datacenter",
			EntityName: "DC0",
			Tag:        "k8s-region",
		},
		{
			Name:       "Cluster validation rule",
			EntityType: "Cluster",
			EntityName: "DC0_C0",
			Tag:        "k8s-zone",
		},
		{
			Name:        "Host validation rule",
			ClusterName: "DC0_C0",
			EntityType:  "Host",
			EntityName:  "DC0_C0_H0",
			Tag:         "owner",
		},
	}

	testCases := []struct {
		name             string
		expectedErr      bool
		validationResult types.ValidationRuleResult
		categories       []vtags.Category
		attachedTags     []vtags.AttachedTags
		expectedStatus   v1.ConditionStatus
	}{
		{
			name:             "DataCenter and Cluster tags Exist",
			expectedErr:      false,
			validationResult: types.ValidationRuleResult{},
			categories:       categories,
			attachedTags:     attachedTags,
			expectedStatus:   v1.ConditionTrue,
		},
		{
			name:             "Empty categories and attachedTags",
			expectedErr:      true,
			validationResult: types.ValidationRuleResult{},
			categories:       []vtags.Category{},
			attachedTags:     []vtags.AttachedTags{},
			expectedStatus:   v1.ConditionFalse,
		},
	}
	for _, tc := range testCases {
		tags.GetCategories = func(manager *vtags.Manager) ([]vtags.Category, error) {
			return tc.categories, nil
		}
		tags.GetAttachedTagsOnObjects = func(tagsManager *vtags.Manager, refs []mo.Reference) ([]vtags.AttachedTags, error) {
			return tc.attachedTags, nil
		}

		for _, rule := range rules {
			vr, err := tagService.ReconcileTagRules(tm, finder, vsphereCloudDriver, rule)
			if vr.Condition.Status != tc.expectedStatus {
				test.Failure("Expected status is not equal to condition status")
			}
			if err == nil && tc.expectedErr {
				test.Failure("Expected error but got no error")
			}
		}
	}

	return test.Success()
}

func (t *TagValidationTest) PreRequisite(ctx *test.TestContext) (tr *test.TestResult) {
	t.log.Info("Executing PreRequisites", "name", t.GetName(), "description", t.GetDescription())

	// setup vCenter simulator
	vcSim := vcsim.NewVCSim("admin@vsphere.local", 8450, t.log)
	vcSim.Start()
	ctx.Put("vcsim", vcSim)

	return test.Success()
}

func (t *TagValidationTest) TearDown(ctx *test.TestContext) {
	t.log.Info("Executing TearDown", "name", t.GetName(), "description", t.GetDescription())

	// shut down vCenter simulator
	vcSimulator := ctx.Get("vcsim")
	vcSimulator.(*vcsim.VCSimulator).Shutdown()
}
