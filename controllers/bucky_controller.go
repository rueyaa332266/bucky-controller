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
	"strconv"
	"strings"

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

// +kubebuilder:rbac:groups=buckycontroller.k8s.io,resources=buckies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=buckycontroller.k8s.io,resources=buckies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=developments,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
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

	// set a label for deployment and service
	labels := map[string]string{
		"app":        "bucky-deployment",
		"controller": req.Name,
	}

	// Create or Update deployment object
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, deploy, func() error {
		seleniumNodeNumber := bucky.Spec.SeleniumNodeNumber
		nodeInstanceNumber := bucky.Spec.NodeInstanceNumber
		BuckyCoreImage := bucky.Spec.BuckyCoreImage
		BuckyCommand := bucky.Spec.BuckyCommand
		replicas := int32(1)
		deploy.Spec.Replicas = &replicas
		var rootuser int64 = 0

		// set labels to spec.selector for our deployment
		if deploy.Spec.Selector == nil {
			deploy.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
		}

		// set labels to template.objectMeta for our deployment
		if deploy.Spec.Template.ObjectMeta.Labels == nil {
			deploy.Spec.Template.ObjectMeta.Labels = labels
		}

		// set a container for our deployment
		containers := []corev1.Container{
			{
				Name:  "selenium-hub",
				Image: "selenium/hub:latest",
			},
			{
				Name:            "bucky-test-script",
				Image:           BuckyCoreImage,
				ImagePullPolicy: "IfNotPresent",
				Command:         strings.Split(BuckyCommand, " "),
				Env: []corev1.EnvVar{
					{
						Name:  "E2E_PARALLEL_NUM",
						Value: strconv.Itoa(nodeInstanceNumber),
					},
				},
				SecurityContext: &corev1.SecurityContext{RunAsUser: &rootuser, RunAsGroup: &rootuser},
			},
		}

		// same ip now
		for i := 0; i < seleniumNodeNumber; i++ {
			containers = append(containers, corev1.Container{
				Name:  "node-chrome-" + strconv.Itoa(i+1),
				Image: "selenium/node-chrome:latest",
				VolumeMounts: []corev1.VolumeMount{{
					Name:      "dshm",
					MountPath: "/dev/shm",
				}},
				Env: []corev1.EnvVar{
					{
						Name:  "NODE_MAX_INSTANCES",
						Value: strconv.Itoa(nodeInstanceNumber),
					},
					{
						Name:  "HUB_HOST",
						Value: "localhost",
					},
					{
						Name:  "HUB_PORT",
						Value: "4444",
					},
					{
						Name:  "NODE_PORT",
						Value: strconv.Itoa(5555 + i),
					},
				},
			})
		}

		volumes := []corev1.Volume{
			{
				Name: "dshm",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium: "Memory",
					},
				},
			},
		}

		// set containers to template.spec.containers for our deployment
		deploy.Spec.Template.Spec.Containers = containers

		// set containers to template.spec.containers for our deployment
		deploy.Spec.Template.Spec.Volumes = volumes

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
		Owns(&corev1.Service{}).
		Complete(r)
}
