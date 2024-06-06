package vsphere

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVCenterUrl(t *testing.T) {
	tests := []struct {
		name            string
		vCenterServer   string
		vCenterUsername string
		vCenterPassword string
		expectedURL     string
		expectError     bool
	}{
		{
			name:            "Valid HTTPS URL",
			vCenterServer:   "vcenter.example.com",
			vCenterUsername: "admin",
			vCenterPassword: "password",
			expectedURL:     "https://admin:password@vcenter.example.com/sdk",
			expectError:     false,
		},
		{
			name:            "Valid HTTP URL Converted to HTTPS",
			vCenterServer:   "http://vcenter.example.com",
			vCenterUsername: "admin",
			vCenterPassword: "password",
			expectedURL:     "https://admin:password@vcenter.example.com/sdk",
			expectError:     false,
		},
		{
			name:            "Invalid URL",
			vCenterServer:   "not a url",
			vCenterUsername: "admin",
			vCenterPassword: "password",
			expectedURL:     "",
			expectError:     true,
		},
		{
			name:            "Trailing Slash Removed",
			vCenterServer:   "vcenter.example.com/",
			vCenterUsername: "admin",
			vCenterPassword: "password",
			expectedURL:     "https://admin:password@vcenter.example.com/sdk",
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultURL, err := getVCenterUrl(tt.vCenterServer, tt.vCenterUsername, tt.vCenterPassword)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resultURL)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resultURL)
				assert.Equal(t, tt.expectedURL, resultURL.String())
				assert.Equal(t, tt.vCenterUsername, resultURL.User.Username())
				password, _ := resultURL.User.Password()
				assert.Equal(t, tt.vCenterPassword, password)
			}
		})
	}
}
