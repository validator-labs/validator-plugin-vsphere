package vsphere

import (
	"fmt"
	"reflect"
	"testing"
)

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
		if !reflect.DeepEqual(err, tc.expectedErr) {
			t.Errorf("expected error (%v), got (%v)", tc.expectedErr, err)
		}
	}
}
