package vsphere

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func TestGetHostSystem(t *testing.T) {
	tests := []struct {
		name           string
		hostNameObj    *types.ManagedObjectReference
		hostSystems    []mo.HostSystem
		expectedResult *mo.HostSystem
	}{
		{
			name: "Found Host System",
			hostNameObj: &types.ManagedObjectReference{
				Value: "host-123",
			},
			hostSystems: []mo.HostSystem{
				{
					Summary: types.HostListSummary{
						Host: &types.ManagedObjectReference{
							Value: "host-456",
						},
					},
				},
				{
					Summary: types.HostListSummary{
						Host: &types.ManagedObjectReference{
							Value: "host-123",
						},
					},
				},
			},
			expectedResult: &mo.HostSystem{
				Summary: types.HostListSummary{
					Host: &types.ManagedObjectReference{Value: "host-123"},
				},
			},
		},
		{
			name: "Not Found Host System",
			hostNameObj: &types.ManagedObjectReference{
				Value: "host-123",
			},
			hostSystems: []mo.HostSystem{
				{
					Summary: types.HostListSummary{
						Host: &types.ManagedObjectReference{
							Value: "host-456",
						},
					},
				},
				{
					Summary: types.HostListSummary{
						Host: &types.ManagedObjectReference{
							Value: "host-789",
						},
					},
				},
			},
		},
		{
			name:        "Nil Input",
			hostNameObj: nil,
			hostSystems: []mo.HostSystem{
				{
					Summary: types.HostListSummary{
						Host: &types.ManagedObjectReference{
							Value: "host-456",
						},
					},
				},
				{
					Summary: types.HostListSummary{
						Host: &types.ManagedObjectReference{
							Value: "host-789",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getHostSystem(tt.hostNameObj, tt.hostSystems)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestValidateHostNTPServers(t *testing.T) {
	testCases := []struct {
		name          string
		expectedErr   error
		hostsDateInfo []HostDateInfo
	}{
		{
			name:        "all valid case",
			expectedErr: nil,
			hostsDateInfo: []HostDateInfo{
				{
					HostName:   "host0",
					NtpServers: []string{"ntp.a.com", "ntp.b.com", "ntp.c.com"},
				},
				{
					HostName:   "host1",
					NtpServers: []string{"ntp.e.com", "ntp.c.com", "ntp.z.com"},
				},
				{
					HostName:   "host2",
					NtpServers: []string{"ntp.c.com"},
				},
				{
					HostName:   "host3",
					NtpServers: []string{"ntp.x.com", "ntp.y.com", "ntp.c.com"},
				},
				{
					HostName:   "host4",
					NtpServers: []string{"ntp.l.com", "ntp.m.com", "ntp.c.com"},
				},
			},
		},
		{
			name:        "first server invalid",
			expectedErr: fmt.Errorf("some of the hosts has differently configured NTP servers"),
			hostsDateInfo: []HostDateInfo{
				{
					HostName:   "host0",
					NtpServers: []string{"ntp.a.com", "ntp.b.com"},
				},
				{
					HostName:   "host1",
					NtpServers: []string{"ntp.e.com", "ntp.c.com", "ntp.z.com"},
				},
				{
					HostName:   "host0",
					NtpServers: []string{"ntp.c.com"},
				},
				{
					HostName:   "host0",
					NtpServers: []string{"ntp.x.com", "ntp.y.com", "ntp.c.com"},
				},
				{
					HostName:   "host0",
					NtpServers: []string{"ntp.l.com", "ntp.m.com", "ntp.c.com"},
				},
			},
		},
		{
			name:        "all invalid servers",
			expectedErr: fmt.Errorf("some of the hosts has differently configured NTP servers"),
			hostsDateInfo: []HostDateInfo{
				{
					HostName:   "host0",
					NtpServers: []string{"ntp.a.com", "ntp.b.com"},
				},
				{
					HostName:   "host1",
					NtpServers: []string{"ntp.e.com"},
				},
				{
					HostName:   "host0",
					NtpServers: []string{"ntp.c.com"},
				},
				{
					HostName:   "host0",
					NtpServers: []string{"ntp.x.com"},
				},
				{
					HostName:   "host0",
					NtpServers: []string{"ntp.l.com"},
				},
			},
		},
		{
			name:        "last server invalid",
			expectedErr: fmt.Errorf("some of the hosts has differently configured NTP servers"),
			hostsDateInfo: []HostDateInfo{
				{
					HostName:   "host0",
					NtpServers: []string{"ntp.a.com", "ntp.b.com", "ntp.c.com"},
				},
				{
					HostName:   "host1",
					NtpServers: []string{"ntp.e.com", "ntp.c.com", "ntp.z.com"},
				},
				{
					HostName:   "host2",
					NtpServers: []string{"ntp.c.com"},
				},
				{
					HostName:   "host3",
					NtpServers: []string{"ntp.x.com", "ntp.y.com", "ntp.c.com"},
				},
				{
					HostName:   "host4",
					NtpServers: []string{"ntp.l.com", "ntp.m.com", "ntp.n.com"},
				},
			},
		},
	}
	for _, tc := range testCases {
		err := validateHostNTPServers(tc.hostsDateInfo)
		if err != nil && !reflect.DeepEqual(err.Error(), tc.expectedErr.Error()) {
			t.Errorf("Expected %v but got %v", tc.expectedErr, err)
		}
		if err == nil && tc.expectedErr != nil {
			t.Errorf("Expected error but got no error")
		}
	}
}
