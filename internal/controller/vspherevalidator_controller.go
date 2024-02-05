/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	"github.com/vmware/govmomi/object"
	vtags "github.com/vmware/govmomi/vapi/tags"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/spectrocloud-labs/validator-plugin-vsphere/api/v1alpha1"
	"github.com/spectrocloud-labs/validator-plugin-vsphere/internal/constants"
	"github.com/spectrocloud-labs/validator-plugin-vsphere/internal/validators/computeresources"
	"github.com/spectrocloud-labs/validator-plugin-vsphere/internal/validators/ntp"
	"github.com/spectrocloud-labs/validator-plugin-vsphere/internal/validators/privileges"
	"github.com/spectrocloud-labs/validator-plugin-vsphere/internal/validators/tags"
	"github.com/spectrocloud-labs/validator-plugin-vsphere/pkg/vsphere"
	vapi "github.com/spectrocloud-labs/validator/api/v1alpha1"
	"github.com/spectrocloud-labs/validator/pkg/types"
	"github.com/spectrocloud-labs/validator/pkg/util/ptr"
	vres "github.com/spectrocloud-labs/validator/pkg/validationresult"
)

var ErrSecretNameRequired = errors.New("auth.secretName is required")

// VsphereValidatorReconciler reconciles a VsphereValidator object
type VsphereValidatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=validation.spectrocloud.labs,resources=vspherevalidators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=validation.spectrocloud.labs,resources=vspherevalidators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=validation.spectrocloud.labs,resources=vspherevalidators/finalizers,verbs=update

// Reconcile compares the state specified by the VsphereValidator object
// against the actual cluster state, and then perform operations to make
// the cluster state reflect the state specified by the user.
func (r *VsphereValidatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.V(0).Info("Reconciling VsphereValidator", "name", req.Name, "namespace", req.Namespace)

	validator := &v1alpha1.VsphereValidator{}
	if err := r.Get(ctx, req.NamespacedName, validator); err != nil {
		// Ignore not-found errors, since they can't be fixed by an immediate requeue
		if apierrs.IsNotFound(err) {
			r.Log.Error(err, "failed to fetch VsphereValidator", "key", req)
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Initialize Vsphere driver
	var vsphereCloudAccount *vsphere.VsphereCloudAccount
	var res *ctrl.Result
	if validator.Spec.Auth.SecretName == "" {
		r.Log.Error(ErrSecretNameRequired, "failed to reconcile VsphereValidator with empty auth.secretName", "key", req)
		return ctrl.Result{}, ErrSecretNameRequired
	} else {
		vsphereCloudAccount, res = r.secretKeyAuth(req, validator)
		if res != nil {
			return *res, nil
		}
	}

	vsphereCloudDriver, err := vsphere.NewVSphereDriver(vsphereCloudAccount.VcenterServer, vsphereCloudAccount.Username, vsphereCloudAccount.Password, validator.Spec.Datacenter)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Get the authorization manager from the Client
	authManager := object.NewAuthorizationManager(vsphereCloudDriver.Client.Client)
	if authManager == nil {
		return ctrl.Result{}, err
	}

	// Get the current user
	userName, err := vsphereCloudDriver.GetCurrentVmwareUser(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	rolePrivilegeValidationService := privileges.NewPrivilegeValidationService(r.Log, vsphereCloudDriver, validator.Spec.Datacenter, authManager, userName)
	tagValidationService := tags.NewTagsValidationService(r.Log)
	computeResourceValidationService := computeresources.NewComputeResourcesValidationService(r.Log, vsphereCloudDriver)
	ntpValidationService := ntp.NewNTPValidationService(r.Log, vsphereCloudDriver, validator.Spec.Datacenter)

	// Get the active validator's validation result
	vr := &vapi.ValidationResult{}
	nn := ktypes.NamespacedName{
		Name:      validationResultName(validator),
		Namespace: req.Namespace,
	}
	if err := r.Get(ctx, nn, vr); err == nil {
		vres.HandleExistingValidationResult(nn, vr, r.Log)
	} else {
		if !apierrs.IsNotFound(err) {
			r.Log.V(0).Error(err, "unexpected error getting ValidationResult", "name", nn.Name, "namespace", nn.Namespace)
		}
		if err := vres.HandleNewValidationResult(r.Client, buildValidationResult(validator), r.Log); err != nil {
			return ctrl.Result{}, err
		}
	}

	tagsManager := vtags.NewManager(vsphereCloudDriver.RestClient)
	finder, _, err := vsphereCloudDriver.GetFinderWithDatacenter(ctx, vsphereCloudDriver.Datacenter)
	if err != nil {
		return ctrl.Result{}, err
	}

	// NTP validation rules
	for _, rule := range validator.Spec.NTPValidationRules {
		validationResult, err := ntpValidationService.ReconcileNTPRule(rule, finder)
		if err != nil {
			r.Log.V(0).Error(err, "failed to reconcile NTP rule")
		}
		r.safeUpdate(nn, validator, validationResult, err)
		r.Log.V(0).Info("Validated NTP rules")
	}

	// entity privilege validation rules
	for _, rule := range validator.Spec.EntityPrivilegeValidationRules {
		validationResult, err := rolePrivilegeValidationService.ReconcileEntityPrivilegeRule(rule, finder)
		if err != nil {
			r.Log.V(0).Error(err, "failed to reconcile entity privilege rule")
		}
		r.safeUpdate(nn, validator, validationResult, err)
		r.Log.V(0).Info("Validated privileges for account", "user", rule.Username)
	}

	// role privilege validation rules
	for _, rule := range validator.Spec.RolePrivilegeValidationRules {
		validationResult, err := rolePrivilegeValidationService.ReconcileRolePrivilegesRule(rule, vsphereCloudDriver, authManager)
		if err != nil {
			r.Log.V(0).Error(err, "failed to reconcile role privilege rule")
		}
		r.safeUpdate(nn, validator, validationResult, err)
		r.Log.V(0).Info("Validated privileges for account", "user", rule.Username)
	}

	// tag validation rules
	for _, rule := range validator.Spec.TagValidationRules {
		r.Log.V(0).Info("Checking if tags are properly assigned")
		validationResult, err := tagValidationService.ReconcileTagRules(tagsManager, finder, vsphereCloudDriver, rule)
		if err != nil {
			r.Log.V(0).Error(err, "failed to reconcile tag validation rule")
		}
		r.safeUpdate(nn, validator, validationResult, err)
		r.Log.V(0).Info("Validated tags", "entity type", rule.EntityType, "entity name", rule.EntityName, "tag", rule.Tag)
	}

	// computeresources validation rules
	for _, rule := range validator.Spec.ComputeResourceRules {
		validationResult, err := computeResourceValidationService.ReconcileComputeResourceValidationRule(rule, finder, vsphereCloudDriver)
		if err != nil {
			r.Log.V(0).Error(err, "failed to reconcile computeresources validation rule")
		}
		r.safeUpdate(nn, validator, validationResult, err)
		r.Log.V(0).Info("Validated compute resources", "scope", rule.Scope, "entity name", rule.EntityName)
	}

	// requeue after two minutes for re-validation
	r.Log.V(0).Info("Requeuing for re-validation in two minutes.", "name", req.Name, "namespace", req.Namespace)
	return ctrl.Result{RequeueAfter: 2 * time.Minute}, nil
}

func (r *VsphereValidatorReconciler) safeUpdate(nn ktypes.NamespacedName, v *v1alpha1.VsphereValidator, vr *types.ValidationResult, err error) {
	vres.SafeUpdateValidationResult(r.Client, nn, vr, v.Spec.ResultCount(), err, r.Log)
}

func (r *VsphereValidatorReconciler) secretKeyAuth(req ctrl.Request, validator *v1alpha1.VsphereValidator) (*vsphere.VsphereCloudAccount, *reconcile.Result) {
	authSecret := &corev1.Secret{}
	nn := ktypes.NamespacedName{Name: validator.Spec.Auth.SecretName, Namespace: req.Namespace}

	if err := r.Get(context.Background(), nn, authSecret); err != nil {
		if apierrs.IsNotFound(err) {
			r.Log.V(0).Error(err, "auth secret does not exist", "name", validator.Spec.Auth.SecretName, "namespace", req.Namespace)
		} else {
			r.Log.V(0).Error(err, "failed to fetch auth secret")
		}
		r.Log.V(0).Info("Requeuing for re-validation in two minutes.", "name", req.Name, "namespace", req.Namespace)
		return nil, &ctrl.Result{RequeueAfter: time.Second * 120}
	}

	username, ok := authSecret.Data["username"]
	if !ok {
		r.Log.V(0).Info("Auth secret missing username", "name", validator.Spec.Auth.SecretName, "namespace", req.Namespace)
		r.Log.V(0).Info("Requeuing for re-validation in two minutes.", "name", req.Name, "namespace", req.Namespace)
		return nil, &ctrl.Result{RequeueAfter: time.Second * 120}
	}

	password, ok := authSecret.Data["password"]
	if !ok {
		r.Log.V(0).Info("Auth secret missing password", "name", validator.Spec.Auth.SecretName, "namespace", req.Namespace)
		r.Log.V(0).Info("Requeuing for re-validation in two minutes.", "name", req.Name, "namespace", req.Namespace)
		return nil, &ctrl.Result{RequeueAfter: time.Second * 120}
	}

	vcenterServer, ok := authSecret.Data["vcenterServer"]
	if !ok {
		r.Log.V(0).Info("Auth secret missing vcenterServer", "name", validator.Spec.Auth.SecretName, "namespace", req.Namespace)
		r.Log.V(0).Info("Requeuing for re-validation in two minutes.", "name", req.Name, "namespace", req.Namespace)
		return nil, &ctrl.Result{RequeueAfter: time.Second * 120}
	}

	insecureSkipVerify, ok := authSecret.Data["insecureSkipVerify"]
	if !ok {
		r.Log.V(0).Info("Auth secret missing insecureSkipVerify", "name", validator.Spec.Auth.SecretName, "namespace", req.Namespace)
		r.Log.V(0).Info("Requeuing for re-validation in two minutes.", "name", req.Name, "namespace", req.Namespace)
		return nil, &ctrl.Result{RequeueAfter: time.Second * 120}
	}

	skipVerify, err := strconv.ParseBool(string(insecureSkipVerify))
	if err != nil {
		r.Log.V(0).Info("Failure converting insecureSkipVerify to bool", "name", validator.Spec.Auth.SecretName, "namespace", req.Namespace)
		r.Log.V(0).Info("Requeuing for re-validation in two minutes.", "name", req.Name, "namespace", req.Namespace)
		return nil, &ctrl.Result{RequeueAfter: time.Second * 120}
	}

	return &vsphere.VsphereCloudAccount{
		Insecure:      skipVerify,
		Password:      string(password),
		Username:      string(username),
		VcenterServer: string(vcenterServer),
	}, nil
}

func buildValidationResult(validator *v1alpha1.VsphereValidator) *vapi.ValidationResult {
	return &vapi.ValidationResult{
		ObjectMeta: metav1.ObjectMeta{
			Name:      validationResultName(validator),
			Namespace: validator.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: validator.APIVersion,
					Kind:       validator.Kind,
					Name:       validator.Name,
					UID:        validator.UID,
					Controller: ptr.Ptr(true),
				},
			},
		},
		Spec: vapi.ValidationResultSpec{
			Plugin:          constants.PluginCode,
			ExpectedResults: validator.Spec.ResultCount(),
		},
	}
}

func validationResultName(validator *v1alpha1.VsphereValidator) string {
	return fmt.Sprintf("validator-plugin-vsphere-%s", validator.Name)
}

// SetupWithManager sets up the controller with the Manager.
func (r *VsphereValidatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.VsphereValidator{}).
		Complete(r)
}
