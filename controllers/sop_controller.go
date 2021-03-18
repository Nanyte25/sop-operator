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
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=app.integreatly.org,resources=sops,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.integreatly.org,resources=sops/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=app.integreatly.org,resources=sops/finalizers,verbs=update

func (r *SOPReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("sop", req.NamespacedName)

	instance := &appv1alpha1.SOP{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)

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

	// Parse identifier and direct to proper action if possible.
	identifier := instance.Spec.Identifier
	logger.Infof("Found sop: %s", identifier)

	switch identifier {
	case "3scale-backup":
		sopactions.Backup3Scale()
	case "amq-backup":
		sopactions.BackupAMQ()
	default:
		logger.Errorf("Unknown sop identifier: %s", identifier)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SOPReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.SOP{}).
		Complete(r)
}
