package vsphere

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-logr/logr"

	"github.com/validator-labs/validator-plugin-vsphere/pkg/vcsim"
)

func TestGetDatastores(t *testing.T) {
	vcSim := vcsim.NewVCSim("admin@vsphere.local", 8451, logr.Logger{})
	vcSim.Start()
	defer vcSim.Shutdown()

	driver, err := NewVCenterDriver(vcSim.Account, vcSim.Options.Datacenter, logr.Logger{})
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"LocalDS_0",
		"LocalDS_1",
	}

	result, err := driver.GetDatastores(context.Background(), vcSim.Options.Datacenter)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("GetDatastores() got %s != expected %s", result, expected)
	}
}
