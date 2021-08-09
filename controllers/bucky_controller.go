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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	buckycontrollerv1alpha1 "github.com/rueyaa332266/bucky-controller/api/v1alpha1"
)

// BuckyReconciler reconciles a Bucky object
type BuckyReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=buckycontroller.k8s.io,resources=buckys,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=buckycontroller.k8s.io,resources=buckys/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=developments,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *BuckyReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("bucky", req.NamespacedName)

	// Load Bucky by name
	var bucky buckycontrollerv1alpha1.Bucky
	log.Info("fetching Bucky Resource")
	if err := r.Get(ctx, req.NamespacedName, &bucky); err != nil {
		log.Error(err, "unable to fetch Bucky")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return ctrl.Result{}, nil
}

var (
	deploymentOwnerKey = ".metadata.controller"
	apiGVStr           = buckycontrollerv1alpha1.GroupVersion.String()
)

func (r *BuckyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(&appsv1.Deployment{}, deploymentOwnerKey, func(rawObj runtime.Object) []string {
		// grab the deployment object, extract the owner...
		deployment := rawObj.(*appsv1.Deployment)
		owner := metav1.GetControllerOf(deployment)
		if owner == nil {
			return nil
		}
		// ...make sure it's a Bucky...
		if owner.APIVersion != apiGVStr || owner.Kind != "Bucky" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&buckycontrollerv1alpha1.Bucky{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
