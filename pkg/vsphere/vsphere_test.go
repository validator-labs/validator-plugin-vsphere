package vsphere

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/validator-labs/validator-plugin-vsphere/api/vcenter"
)

func Test_getVCenterURL(t *testing.T) {
	tests := []struct {
		name        string
		account     vcenter.Account
		expectedURL string
		expectError bool
	}{
		{
			name: "Valid HTTPS URL",
			account: vcenter.Account{
				Host:     "vcenter.example.com",
				Username: "admin",
				Password: "password",
			},
			expectedURL: "https://admin:password@vcenter.example.com/sdk",
			expectError: false,
		},
		{
			name: "Valid HTTP URL Converted to HTTPS",
			account: vcenter.Account{
				Host:     "http://vcenter.example.com",
				Username: "admin",
				Password: "password",
			},
			expectedURL: "https://admin:password@vcenter.example.com/sdk",
			expectError: false,
		},
		{
			name: "Invalid URL",
			account: vcenter.Account{
				Host:     "not a url",
				Username: "admin",
				Password: "password",
			},
			expectedURL: "",
			expectError: true,
		},
		{
			name: "Trailing Slash Removed",
			account: vcenter.Account{
				Host:     "vcenter.example.com/",
				Username: "admin",
				Password: "password",
			},
			expectedURL: "https://admin:password@vcenter.example.com/sdk",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultURL, err := getVCenterURL(tt.account)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resultURL)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resultURL)
				assert.Equal(t, tt.expectedURL, resultURL.String())
				assert.Equal(t, tt.account.Username, resultURL.User.Username())
				password, _ := resultURL.User.Password()
				assert.Equal(t, tt.account.Password, password)
			}
		})
	}
}
