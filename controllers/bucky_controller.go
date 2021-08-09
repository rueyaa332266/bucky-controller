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
	corev1 "k8s.io/api/core/v1"
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

	// Create or Update deployment object which match Bucky.Spec.
	deploymentName := "bucky-deployment"
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: req.Namespace,
		},
	}

	// Create or Update deployment object
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, deploy, func() error {
		seleniumNodeNumber := bucky.Spec.SeleniumNodeNumber
		nodeInstanceNumber := bucky.Spec.NodeInstanceNumber
		replicas := int32(1)
		deploy.Spec.Replicas = &replicas

		// set a label for our deployment
		labels := map[string]string{
			"app":        "bucky-deployment",
			"controller": req.Name,
		}

		// set labels to spec.selector for our deployment
		if deploy.Spec.Selector == nil {
			deploy.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
		}

		// set labels to template.objectMeta for our deployment
		if deploy.Spec.Template.ObjectMeta.Labels == nil {
			deploy.Spec.Template.ObjectMeta.Labels = labels
		}

		// set a container for our deployment
		// !!! seleniumNodeNumber for loop the ip
		containers := []corev1.Container{
			{
				Name:  "node-chrome",
				Image: "selenium/node-chrome:latest",
				Env: []corev1.EnvVar{
					{
						Name:  "NODE_MAX_INSTANCES",
						Value: nodeInstanceNumber,
					},
				},
			},
		}

		// set containers to template.spec.containers for our deployment
		if deploy.Spec.Template.Spec.Containers == nil {
			deploy.Spec.Template.Spec.Containers = containers
		}

		// set the owner so that garbage collection can kicks in
		if err := ctrl.SetControllerReference(&bucky, deploy, r.Scheme); err != nil {
			log.Error(err, "unable to set ownerReference from Bucky to Deployment")
			return err
		}

		// end of ctrl.CreateOrUpdate
		return nil
	}); err != nil {

		// error handling of ctrl.CreateOrUpdate
		log.Error(err, "unable to ensure deployment is correct")
		return ctrl.Result{}, err

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
