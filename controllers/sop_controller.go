/*


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

package controllers

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"strings"
	"time"

	appv1alpha1 "github.com/carlkyrillos/sop-operator/api/v1alpha1"
	"github.com/carlkyrillos/sop-operator/controllers/sopactions"
	"github.com/go-logr/logr"
	logger "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SOPReconciler reconciles a SOP object
type SOPReconciler struct {
	Client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var firstPass = true

// +kubebuilder:rbac:groups=app.integreatly.org,resources=sops,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.integreatly.org,resources=sops/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=app.integreatly.org,resources=sops/finalizers,verbs=update

func (r *SOPReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("sop", req.NamespacedName)

	if firstPass {
		deleteResources(ctx, r.Client)
		firstPass = false
	}

	logger.Info("Beginning SOP reconciliation...")

	instance := &appv1alpha1.SOP{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)

	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// If the status is already complete, return and don't requeue.
	if instance.Status.Phase == "complete" {
		return ctrl.Result{}, nil
	}

	// Otherwise, parse identifier and direct to proper action if possible.
	identifier := instance.Spec.Identifier
	logger.Infof("Found sop: %s", identifier)

	switch identifier {
	case "3scale-backup":
		sopactions.Backup3Scale()
	case "amq-backup":
		sopactions.BackupAMQ()
	case "rhsso-upgrade":
		// Set phase to in progress.
		instance.Status.Phase = "in progress"
		_ = r.Client.Status().Update(context.TODO(), instance)

		// Upgrade RHSSO
		err = sopactions.UpgradeRHSSO(ctx, r.Client)

		if err != nil {
			// An error occurred while upgrading RHSSO. Requeue the request.
			return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
		} else {
			// No errors present. Set phase to complete.
			instance.Status.Phase = "complete"
			_ = r.Client.Status().Update(context.TODO(), instance)
			return ctrl.Result{}, nil
		}
	default:
		logger.Errorf("Unknown sop identifier: %s", identifier)
	}

	return ctrl.Result{}, nil
}

// This is a helper function that deletes a hardcoded set of resources to help during development.
func deleteResources(ctx context.Context, client client.Client) {
	// Find the old daemon set and delete it
	dsKey := types.NamespacedName{
		Name:      "tempds",
		Namespace: "redhat-rhoam-rhsso",
	}
	ds := &appsv1.DaemonSet{}
	_ = client.Get(ctx, dsKey, ds)
	err := client.Delete(ctx, ds)
	if err != nil {
		logger.Error("unable to delete old daemon set")
	}

	// Get all old events and delete them
	eventList := &corev1.EventList{}
	_ = client.List(ctx, eventList)
	for _, event := range eventList.Items {
		if strings.Contains(event.Name, "tempds") {
			err := client.Delete(ctx, &event)
			if err != nil {
				logger.Error("unable to delete old events")
			}
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *SOPReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.SOP{}).
		Complete(r)
}
