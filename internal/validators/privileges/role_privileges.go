package privileges

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/ssoadmin"
	"github.com/vmware/govmomi/sts"
	"github.com/vmware/govmomi/vim25/soap"
	corev1 "k8s.io/api/core/v1"

	"github.com/validator-labs/validator-plugin-vsphere/api/v1alpha1"
	"github.com/validator-labs/validator-plugin-vsphere/internal/constants"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/vsphere"
	vapi "github.com/validator-labs/validator/api/v1alpha1"
	vapiconstants "github.com/validator-labs/validator/pkg/constants"
	"github.com/validator-labs/validator/pkg/types"
	"github.com/validator-labs/validator/pkg/util"
)

var (
	IsAdminAccount                    = isAdminAccount
	GetUserAndGroupPrincipals         = getUserAndGroupPrincipals
	ErrRequiredRolePrivilegesNotFound = errors.New("one or more required role privileges was not found for account")
)

func buildValidationResult(rule v1alpha1.GenericRolePrivilegeValidationRule, validationType string) *types.ValidationRuleResult {
	state := vapi.ValidationSucceeded
	latestCondition := vapi.DefaultValidationCondition()
	latestCondition.Message = fmt.Sprintf("All required %s permissions were found", validationType)
	latestCondition.ValidationRule = fmt.Sprintf("%s-%s", vapiconstants.ValidationRulePrefix, rule.Username)
	latestCondition.ValidationType = validationType

	return &types.ValidationRuleResult{Condition: &latestCondition, State: &state}
}

func setFailureStatus(vr *types.ValidationRuleResult, msg string) {
	vr.State = util.Ptr(vapi.ValidationFailed)
	vr.Condition.Message = msg
	vr.Condition.Status = corev1.ConditionFalse
	return
}

func (s *PrivilegeValidationService) ReconcileRolePrivilegesRule(rule v1alpha1.GenericRolePrivilegeValidationRule, driver *vsphere.VSphereCloudDriver, authManager *object.AuthorizationManager) (*types.ValidationRuleResult, error) {
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vr := buildValidationResult(rule, constants.ValidationTypeRolePrivileges)
	failMsg := fmt.Sprintf("One or more required privileges was not found, or a condition was not met for account: %s", rule.Username)

	privileges, err := getPrivileges(ctx, driver, authManager, rule.Username)
	if err != nil {
		vr.Condition.Failures = append(vr.Condition.Failures, fmt.Sprintf("Failed to get user privileges for %s due to error: %s", rule.Username, err))
		setFailureStatus(vr, failMsg)
		return vr, err
	}

	for _, privilege := range rule.Privileges {
		valid := isValidRule(privilege, privileges)
		if !valid {
			vr.Condition.Failures = append(vr.Condition.Failures, fmt.Sprintf("Privilege: %s, was not found in the user's privileges", privilege))
		}
	}

	if len(vr.Condition.Failures) > 0 {
		setFailureStatus(vr, failMsg)
		err = ErrRequiredRolePrivilegesNotFound
	}

	return vr, err
}

func isValidRule(privilege string, privileges map[string]bool) bool {
	return privileges[privilege]
}

func configureSSOClient(ctx context.Context, driver *vsphere.VSphereCloudDriver) (*ssoadmin.Client, error) {
	vc := driver.Client.Client
	ssoClient, err := ssoadmin.NewClient(ctx, vc)
	if err != nil {
		return nil, err
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
			return nil, cerr
		}

		userInfo := url.UserPassword(driver.VCenterUsername, driver.VCenterPassword)
		req := sts.TokenRequest{
			Certificate: vc.Certificate(),
			Userinfo:    userInfo,
		}

		header.Security, cerr = tokens.Issue(ctx, req)
		if cerr != nil {
			return nil, cerr
		}
	}

	if err = ssoClient.Login(ssoClient.WithHeader(ctx, header)); err != nil {
		return nil, err
	}

	return ssoClient, nil
}

func isAdminAccount(ctx context.Context, driver *vsphere.VSphereCloudDriver) (bool, error) {
	ssoClient, err := configureSSOClient(ctx, driver)
	if err != nil {
		return false, err
	}
	defer ssoClient.Logout(ctx)

	_, err = ssoClient.FindUser(ctx, driver.VCenterUsername)
	if err != nil {
		if strings.Contains(err.Error(), "NoPermission") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func getPrivileges(ctx context.Context, driver *vsphere.VSphereCloudDriver, authManager *object.AuthorizationManager, username string) (map[string]bool, error) {
	isAdmin, err := IsAdminAccount(ctx, driver)
	if err != nil {
		return nil, err
	}

	groupPrincipals := make([]string, 0)
	if isAdmin {
		userPrincipal, groupPrincipals, err := GetUserAndGroupPrincipals(ctx, username, driver)
		if err != nil {
			return nil, err
		}

		privileges, err := vsphere.GetVmwareUserPrivileges(ctx, userPrincipal, groupPrincipals, authManager)
		if err != nil {
			return nil, err
		}

		return privileges, nil
	}

	userPrincipal, err := driver.GetCurrentVmwareUser(ctx)
	if err != nil {
		return nil, err
	}

	if !isSameUser(userPrincipal, username) {
		return nil, errors.New("Not authorized to get privileges for another user from non-admin account")
	}

	privileges, err := vsphere.GetVmwareUserPrivileges(ctx, userPrincipal, groupPrincipals, authManager)
	if err != nil {
		return nil, err
	}

	return privileges, nil
}

// checks if a user principle (VSPHERE.LOCAL\username) matches the username (username@vsphere.local)
// it is only considered a match if the usernames on both are identical and the domains match
func isSameUser(userPrincipal string, username string) bool {
	userPrincipalParts := strings.Split(userPrincipal, "\\")
	usernameParts := strings.Split(username, "@")

	if len(userPrincipalParts) != 2 || len(usernameParts) != 2 {
		return false
	}

	return strings.ToLower(userPrincipalParts[0]) == strings.ToLower(usernameParts[1]) && userPrincipalParts[1] == usernameParts[0]
}

func getUserAndGroupPrincipals(ctx context.Context, username string, driver *vsphere.VSphereCloudDriver) (string, []string, error) {
	var groups []string

	ssoClient, err := configureSSOClient(ctx, driver)
	if err != nil {
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

func getAccountPrivileges(ctx context.Context, driver *vsphere.VSphereCloudDriver) (map[string]bool, error) {
	authManager := object.NewAuthorizationManager(driver.Client.Client)
	if authManager == nil {
		return nil, fmt.Errorf("Error getting authorization manager")
	}

	userName, err := driver.GetCurrentVmwareUser(ctx)
	if err != nil {
		return nil, err
	}

	userPrivileges, err := vsphere.GetVmwareUserPrivileges(ctx, userName, []string{}, authManager)
	if err != nil {
		return nil, err
	}

	return userPrivileges, nil
}
