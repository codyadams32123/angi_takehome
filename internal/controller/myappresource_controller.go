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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	groupv1alpha1 "github.com/codyadams32123/angi_takehome/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
)

// MyAppResourceReconciler reconciles a MyAppResource object
type MyAppResourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=group.angi.com,resources=myappresources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=group.angi.com,resources=myappresources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=group.angi.com,resources=myappresources/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MyAppResource object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *MyAppResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("Angi", req.NamespacedName)

	// Grab the instance of MyAppResource
	instance := &groupv1alpha1.MyAppResource{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "MyAppResource not found")
			return ctrl.Result{}, nil
		}
		// Requeue if error reading the object that isn't NotFound
		log.Error(err, "MyAppResource couldn't be read")
		return ctrl.Result{}, err
	}

	// Check for PodInfo deployment
	podInfoFound := &appsv1.Deployment{}

	err = r.Client.Get(ctx, types.NamespacedName{Name: "podinfo" + instance.Name, Namespace: instance.Namespace}, podInfoFound)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyAppResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&groupv1alpha1.MyAppResource{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func createRedis(redis *groupv1alpha1.MyAppResource) error {

	// foundDeployment := &appsv1.Deployment{}
	// err := r.Client.Get(ctx, types.NamespacedName{Name: "redis" + redis.Name}, foundDeployment)
	// deployment := &appsv1.Deployment{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name: "redis" + redis.Name,
	// 		Labels: map[string]string{
	// 			"app": "redis",
	// 			"name": "redis",
	// 		}},
	// 	Spec: appsv1.DeploymentSpec{
	// 		Selector: &metav1.LabelSelector{
	// 			MatchLabels: map[string]string{
	// 				"app": "redis",
	// 				"name": "redis",
	// 			},
	// 		},
	// 	},
	// }

	// service := &corev1.Service{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name: "redis" + redis.Name,
	// 		Labels: map[string]string{
	// 			"app": "redis",
	// 			"name": "redis",
	// 		}},
	// 	Spec: corev1.ServiceSpec{
	// 		Selector: map[string]string{
	// 			"app": "redis",
	// 			"name": "redis",
	// 		},
	// 		Ports: []corev1.ServicePort{
				
	// }

	return nil
}