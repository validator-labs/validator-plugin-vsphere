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
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ktypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/validator-labs/validator-plugin-vsphere/api/v1alpha1"
	"github.com/validator-labs/validator-plugin-vsphere/api/vcenter"
	"github.com/validator-labs/validator-plugin-vsphere/pkg/validate"
	vapi "github.com/validator-labs/validator/api/v1alpha1"
	vres "github.com/validator-labs/validator/pkg/validationresult"
)

var errCredentialsRequired = errors.New("auth.secretName or auth.cloudAccount is required")

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
	l := r.Log.V(0).WithValues("name", req.Name, "namespace", req.Namespace)

	l.Info("Reconciling VsphereValidator")

	validator := &v1alpha1.VsphereValidator{}
	if err := r.Get(ctx, req.NamespacedName, validator); err != nil {
		// Ignore not-found errors, since they can't be fixed by an immediate requeue
		if apierrs.IsNotFound(err) {
			l.Error(err, "failed to fetch VsphereValidator")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Get the active validator's validation result
	vr := &vapi.ValidationResult{}
	p, err := patch.NewHelper(vr, r.Client)
	if err != nil {
		l.Error(err, "failed to create patch helper")
		return ctrl.Result{}, err
	}
	nn := ktypes.NamespacedName{
		Name:      vres.Name(validator),
		Namespace: req.Namespace,
	}
	if err := r.Get(ctx, nn, vr); err == nil {
		vres.HandleExisting(vr, r.Log)
	} else {
		if !apierrs.IsNotFound(err) {
			l.Error(err, "unexpected error getting ValidationResult")
		}
		if err := vres.HandleNew(ctx, r.Client, p, vres.Build(validator), r.Log); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Millisecond}, nil
	}

	// Always update the expected result count in case the validator's rules have changed
	vr.Spec.ExpectedResults = validator.Spec.ResultCount()

	// Initialize Vsphere driver
	if validator.Spec.Auth.SecretName == "" && validator.Spec.Auth.Account == nil {
		l.Error(errCredentialsRequired, "failed to reconcile VsphereValidator with empty credentials")
		return ctrl.Result{}, errCredentialsRequired
	}
	if validator.Spec.Auth.SecretName != "" {
		if err := r.secretKeyAuth(req, validator); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Validate the rules
	resp := validate.Validate(ctx, validator.Spec, r.Log)

	// Patch the ValidationResult with the latest ValidationRuleResults
	if err := vres.SafeUpdate(ctx, p, vr, resp, r.Log); err != nil {
		return ctrl.Result{}, err
	}

	// requeue after two minutes for re-validation
	l.Info("Requeuing for re-validation in two minutes.")
	return ctrl.Result{RequeueAfter: 2 * time.Minute}, nil
}

func (r *VsphereValidatorReconciler) secretKeyAuth(req ctrl.Request, validator *v1alpha1.VsphereValidator) error {

	authSecret := &corev1.Secret{}
	nn := ktypes.NamespacedName{Name: validator.Spec.Auth.SecretName, Namespace: req.Namespace}

	if err := r.Get(context.Background(), nn, authSecret); err != nil {
		return fmt.Errorf("failed to get secret %s: %w", validator.Spec.Auth.SecretName, err)
	}

	username, ok := authSecret.Data["username"]
	if !ok {
		return errors.New("auth secret missing username")
	}
	password, ok := authSecret.Data["password"]
	if !ok {
		return errors.New("auth secret missing password")
	}
	vcenterServer, ok := authSecret.Data["vcenterServer"]
	if !ok {
		return errors.New("auth secret missing vcenterServer")
	}
	insecureSkipVerify, ok := authSecret.Data["insecureSkipVerify"]
	if !ok {
		return errors.New("auth secret missing insecureSkipVerify")
	}
	skipVerify, err := strconv.ParseBool(string(insecureSkipVerify))
	if err != nil {
		return fmt.Errorf("failed to convert insecureSkipVerify to bool: %w", err)
	}

	validator.Spec.Auth.Account = &vcenter.Account{
		Insecure: skipVerify,
		Username: string(username),
		Password: string(password),
		Host:     string(vcenterServer),
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VsphereValidatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.VsphereValidator{}).
		Complete(r)
}
