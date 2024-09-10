package vsphere

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-logr/logr"

	"github.com/validator-labs/validator-plugin-vsphere/pkg/vcsim"
)

func TestGetDistributedVirtualPortgroups(t *testing.T) {
	vcSim := vcsim.NewVCSim("admin@vsphere.local", 8452, logr.Logger{})
	vcSim.Start()
	defer vcSim.Shutdown()

	driver, err := NewVCenterDriver(vcSim.Account, vcSim.Options.Datacenter, logr.Logger{})
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"DVS0-DVUplinks-9",
		"DC0_DVPG0",
	}

	result, err := driver.GetDistributedVirtualPortgroups(context.Background(), vcSim.Options.Datacenter)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("GetDistributedVirtualPortgroups() got %s != expected %s", result, expected)
	}
}

func TestGetDistributedVirtualSwitches(t *testing.T) {
	vcSim := vcsim.NewVCSim("admin@vsphere.local", 8453, logr.Logger{})
	vcSim.Start()
	defer vcSim.Shutdown()

	driver, err := NewVCenterDriver(vcSim.Account, vcSim.Options.Datacenter, logr.Logger{})
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"DVS0",
	}

	result, err := driver.GetDistributedVirtualSwitches(context.Background(), vcSim.Options.Datacenter)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("GetDistributedVirtualSwitches() got %s != expected %s", result, expected)
	}
}

func TestGetNetworks(t *testing.T) {
	vcSim := vcsim.NewVCSim("admin@vsphere.local", 8454, logr.Logger{})
	vcSim.Start()
	defer vcSim.Shutdown()

	driver, err := NewVCenterDriver(vcSim.Account, vcSim.Options.Datacenter, logr.Logger{})
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"VM Network",
	}

	result, err := driver.GetNetworks(context.Background(), vcSim.Options.Datacenter)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("GetNetworks() got %s != expected %s", result, expected)
	}
}

func TestGetOpaqueNetworks(t *testing.T) {
	vcSim := vcsim.NewVCSim("admin@vsphere.local", 8455, logr.Logger{})
	vcSim.Start()
	defer vcSim.Shutdown()

	driver, err := NewVCenterDriver(vcSim.Account, vcSim.Options.Datacenter, logr.Logger{})
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{}

	result, err := driver.GetOpaqueNetworks(context.Background(), vcSim.Options.Datacenter)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("GetOpaqueNetworks() got %s != expected %s", result, expected)
	}
}

func TestGetNetworkTypeByName(t *testing.T) {
	vcSim := vcsim.NewVCSim("admin@vsphere.local", 8456, logr.Logger{})
	vcSim.Start()
	defer vcSim.Shutdown()

	driver, err := NewVCenterDriver(vcSim.Account, vcSim.Options.Datacenter, logr.Logger{})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "VM Network",
			expected: "Network",
		},
		{
			name:     "DVS0",
			expected: "Distributed Switch",
		},
		{
			name:     "DC0_DVPG0",
			expected: "Distributed Port Group",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := driver.GetNetworkTypeByName(context.Background(), vcSim.Options.Datacenter, tc.name)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("GetNetworkTypeByName() got %s != expected %s", result, tc.expected)
			}
		})
	}
}
