package privileges

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/object"
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
}

func (s *PrivilegeValidationService) ReconcileRolePrivilegesRule(rule v1alpha1.GenericRolePrivilegeValidationRule, driver *vsphere.VSphereCloudDriver, authManager *object.AuthorizationManager) (*types.ValidationRuleResult, error) {
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vr := buildValidationResult(rule, constants.ValidationTypeRolePrivileges)
	failMsg := fmt.Sprintf("One or more required privileges was not found, or a condition was not met for account: %s", rule.Username)

	privileges, err := getPrivileges(ctx, driver, authManager, rule.Username, s.log)
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

func getPrivileges(ctx context.Context, driver *vsphere.VSphereCloudDriver, authManager *object.AuthorizationManager, username string, log logr.Logger) (map[string]bool, error) {
	isAdmin, err := vsphere.IsAdminAccount(ctx, driver)
	if err != nil {
		return nil, err
	}

	if isAdmin {
		userPrincipal, groupPrincipals, err := GetUserAndGroupPrincipals(ctx, username, driver, log)
		if err != nil {
			return nil, err
		}

		return vsphere.GetVmwareUserPrivileges(ctx, userPrincipal, groupPrincipals, authManager)
	}

	userPrincipal, err := driver.GetCurrentVmwareUser(ctx)
	if err != nil {
		return nil, err
	}

	if !isSameUser(userPrincipal, username) {
		return nil, errors.New("not authorized to get privileges for another user from non-admin account")
	}

	groupPrincipals := make([]string, 0)
	return vsphere.GetVmwareUserPrivileges(ctx, userPrincipal, groupPrincipals, authManager)
}

// checks if a user principle (VSPHERE.LOCAL\username) matches the username (username@vsphere.local)
// it is only considered a match if the usernames on both are identical and the domains match
func isSameUser(userPrincipal string, username string) bool {
	userPrincipalParts := strings.Split(userPrincipal, `\`)
	usernameParts := strings.Split(username, "@")

	if len(userPrincipalParts) != 2 || len(usernameParts) != 2 {
		return false
	}
	return strings.EqualFold(userPrincipalParts[0], usernameParts[1]) && userPrincipalParts[1] == usernameParts[0]
}

func getUserAndGroupPrincipals(ctx context.Context, username string, driver *vsphere.VSphereCloudDriver, log logr.Logger) (string, []string, error) {
	ssoClient, err := vsphere.ConfigureSSOClient(ctx, driver)
	if err != nil {
		return "", nil, err
	}
	defer func() {
		if err := ssoClient.Logout(ctx); err != nil {
			log.Error(err, "Failed to logout from SSO client")
		}
	}()

	user, err := ssoClient.FindUser(ctx, username)
	if err != nil {
		return "", nil, err
	}

	parentGroups, err := ssoClient.FindParentGroups(ctx, user.Id)
	if err != nil {
		return "", nil, err
	}

	groups := make([]string, 0, len(parentGroups))
	for _, group := range parentGroups {
		groups = append(groups, fmt.Sprintf("%s\\%s", strings.ToUpper(group.Domain), group.Name))
	}

	userPrincipal := fmt.Sprintf("%s\\%s", strings.ToUpper(user.Id.Domain), user.Id.Name)

	return userPrincipal, groups, nil
}
