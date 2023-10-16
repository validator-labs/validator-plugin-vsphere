package privileges

import (
	"context"
	"fmt"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/pkg/vsphere"
	"net/url"
	"os"
	"strings"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/ssoadmin"
	"github.com/vmware/govmomi/sts"
	"github.com/vmware/govmomi/vim25/soap"
	corev1 "k8s.io/api/core/v1"

	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/api/v1alpha1"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/internal/constants"
	v8or "github.com/spectrocloud-labs/valid8or/api/v1alpha1"
	v8orconstants "github.com/spectrocloud-labs/valid8or/pkg/constants"
	"github.com/spectrocloud-labs/valid8or/pkg/types"
	"github.com/spectrocloud-labs/valid8or/pkg/util/ptr"
)

var GetUserAndGroupPrincipals = getUserAndGroupPrincipals

func buildValidationResult(rule v1alpha1.GenericRolePrivilegeValidationRule, validationType string) *types.ValidationResult {
	state := v8or.ValidationSucceeded
	latestCondition := v8or.DefaultValidationCondition()
	latestCondition.Message = fmt.Sprintf("All required %s permissions were found", validationType)
	latestCondition.ValidationRule = fmt.Sprintf("%s-%s", v8orconstants.ValidationRulePrefix, rule.Username)
	latestCondition.ValidationType = validationType

	return &types.ValidationResult{Condition: &latestCondition, State: &state}
}

func (s *PrivilegeValidationService) ReconcileRolePrivilegesRule(rule v1alpha1.GenericRolePrivilegeValidationRule, driver *vsphere.VSphereCloudDriver, authManager *object.AuthorizationManager) (*types.ValidationResult, error) {
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vr := buildValidationResult(rule, constants.ValidationTypeRolePrivileges)

	userPrincipal, groupPrincipals, err := GetUserAndGroupPrincipals(ctx, rule.Username, driver)
	if err != nil {
		return nil, err
	}

	privileges, err := vsphere.GetVmwareUserPrivileges(userPrincipal, groupPrincipals, authManager)
	if err != nil {
		return nil, err
	}

	for _, privilege := range rule.Privileges {
		valid := isValidRule(privilege, privileges)
		if !valid {
			vr.Condition.Failures = append(vr.Condition.Failures, fmt.Sprintf("Privilege: %s, was not found in the user's privileges", privilege))
		}
	}

	if len(vr.Condition.Failures) > 0 {
		vr.State = ptr.Ptr(v8or.ValidationFailed)
		vr.Condition.Message = "One or more required privileges was not found, or a condition was not met"
		vr.Condition.Status = corev1.ConditionFalse
		err = fmt.Errorf("one or more required privileges was not found for account: %s", rule.Username)
	}

	return vr, err
}

func isValidRule(privilege string, privileges map[string]bool) bool {
	return privileges[privilege]
}

func getUserAndGroupPrincipals(ctx context.Context, username string, driver *vsphere.VSphereCloudDriver) (string, []string, error) {
	var groups []string
	vc := driver.Client.Client

	ssoClient, err := ssoadmin.NewClient(ctx, vc)
	if err != nil {
		return "", nil, err
	}

	token := os.Getenv("SSO_LOGIN_TOKEN")
	header := soap.Header{
		Security: &sts.Signer{
			Certificate: vc.Certificate(),
			Token:       token,
		},
	}
	if token == "" {
		tokens, cerr := sts.NewClient(ctx, vc)
		if cerr != nil {
			return "", nil, cerr
		}

		userInfo := url.UserPassword(driver.VCenterUsername, driver.VCenterPassword)
		req := sts.TokenRequest{
			Certificate: vc.Certificate(),
			Userinfo:    userInfo,
		}

		header.Security, cerr = tokens.Issue(ctx, req)
		if cerr != nil {
			return "", nil, cerr
		}
	}

	if err = ssoClient.Login(ssoClient.WithHeader(ctx, header)); err != nil {
		return "", nil, err
	}
	defer ssoClient.Logout(ctx)

	user, err := ssoClient.FindUser(ctx, username)
	if err != nil {
		return "", nil, err
	}

	parentGroups, err := ssoClient.FindParentGroups(ctx, user.Id)
	if err != nil {
		return "", nil, err
	}
	for _, group := range parentGroups {
		groups = append(groups, fmt.Sprintf("%s\\%s", strings.ToUpper(group.Domain), group.Name))
	}

	userPrincipal := fmt.Sprintf("%s\\%s", strings.ToUpper(user.Id.Domain), user.Id.Name)

	return userPrincipal, groups, nil
}
