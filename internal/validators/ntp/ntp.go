package ntp

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/api/v1alpha1"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/internal/constants"
	"github.com/spectrocloud-labs/valid8or-plugin-vsphere/pkg/vsphere"
	v8or "github.com/spectrocloud-labs/valid8or/api/v1alpha1"
	v8orconstants "github.com/spectrocloud-labs/valid8or/pkg/constants"
	"github.com/spectrocloud-labs/valid8or/pkg/types"
	"github.com/spectrocloud-labs/valid8or/pkg/util/ptr"
	"github.com/vmware/govmomi/find"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

type NTPValidationService struct {
	log        logr.Logger
	driver     *vsphere.VSphereCloudDriver
	datacenter string
}

func NewNTPValidationService(log logr.Logger, driver *vsphere.VSphereCloudDriver, datacenter string) *NTPValidationService {
	return &NTPValidationService{
		log:        log,
		driver:     driver,
		datacenter: datacenter,
	}
}

func buildValidationResult(rule v1alpha1.NTPValidationRule, validationType string) *types.ValidationResult {
	state := v8or.ValidationSucceeded
	latestCondition := v8or.DefaultValidationCondition()
	latestCondition.Message = fmt.Sprintf("All required ntp rules were satisfied")
	latestCondition.ValidationRule = fmt.Sprintf("%s-%s", v8orconstants.ValidationRulePrefix, strings.ReplaceAll(rule.Name, " ", "-"))
	latestCondition.ValidationType = validationType

	return &types.ValidationResult{Condition: &latestCondition, State: &state}
}

func (n *NTPValidationService) ReconcileNTPRule(rule v1alpha1.NTPValidationRule, finder *find.Finder) (*types.ValidationResult, error) {
	var err error
	vr := buildValidationResult(rule, constants.ValidationTypeNTP)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	valid, failures, err := n.driver.ValidateHostNTPSettings(ctx, n.driver, finder, n.datacenter, rule.ClusterName, rule.Hosts)
	if !valid {
		vr.Condition.Failures = failures
	}

	if len(vr.Condition.Failures) > 0 {
		vr.State = ptr.Ptr(v8or.ValidationFailed)
		vr.Condition.Message = fmt.Sprintf("One or more ntp rules were not satisfied for rule: %s", rule.Name)
		vr.Condition.Status = corev1.ConditionFalse
		err = fmt.Errorf("one or more ntp rules were not satisfied for rule: %s", rule.Name)
	}

	return vr, err
}
