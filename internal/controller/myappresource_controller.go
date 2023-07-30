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

	"strconv"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	groupv1alpha1 "github.com/codyadams32123/angi_takehome/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MyAppResourceReconciler reconciles a MyAppResource object
type MyAppResourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=group.angi.com,resources=myappresources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=group.angi.com,resources=myappresources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=group.angi.com,resources=myappresources/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services/status,verbs=get;update;patch

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

	// Create Redis first. Makes it less likely errors experienced in podInfo deployment
	if instance.Spec.Redis.Enabled {
		redisDeploymentFound := &appsv1.Deployment{}
		err = r.Client.Get(ctx, types.NamespacedName{Name: "redis-" + instance.Name, Namespace: instance.Namespace}, redisDeploymentFound)

		if err != nil && errors.IsNotFound(err) {
			//Creating Deployment as it doesn't existt
			log.Info("Creating Redis Deployment")
			redisDeployment := createRedisDeployment(instance)
			ctrl.SetControllerReference(instance, redisDeployment, r.Scheme)
			err := r.Client.Create(ctx, redisDeployment)
			if err != nil {
				log.Error(err, "Failed to create Redis Deployment")
				return ctrl.Result{}, err
			}
			log.Info("Created Redis Deployment")
		}

		redisServiceFound := &corev1.Service{}
		err = r.Client.Get(ctx, types.NamespacedName{Name: "redis-" + instance.Name, Namespace: instance.Namespace}, redisServiceFound)

		if err != nil && errors.IsNotFound(err) {
			//Creating service as it doesn't exist
			redisService := createRedisService(instance)
			ctrl.SetControllerReference(instance, redisService, r.Scheme)
			err := r.Client.Create(ctx, redisService)
			if err != nil {
				log.Error(err, "Failed to create Redis Service")
				return ctrl.Result{}, err
			}
		}
		log.Info("Created Redis Service")

	}

	// Check for PodInfo deployment
	podInfoFound := &appsv1.Deployment{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: "podinfo-" + instance.Name, Namespace: instance.Namespace}, podInfoFound)

	if err != nil && errors.IsNotFound(err) {
		//Creating if it doesn't exist
		log.Info("Creating PodInfo Deployment")
		podInfoDeployment := createPodInfoDeployment(instance)
		ctrl.SetControllerReference(instance, podInfoDeployment, r.Scheme)
		err := r.Client.Create(ctx, podInfoDeployment)
		if err != nil {
			log.Error(err, "Failed to create PodInfo Deployment")
			return ctrl.Result{}, err
		}
		// err = r.Status().Update(ctx, instance, &client.UpdateOptions{})
		log.Info("Created PodInfo Deployment")
	} else if err != nil {
		//Requeue if failed to read object for another reason
		log.Error(err, "Failed to get PodInfo Deployment")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyAppResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&groupv1alpha1.MyAppResource{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

/*
Generate PodInfo Deployment and apply to cluster
*/
func createPodInfoDeployment(podInfo *groupv1alpha1.MyAppResource) *appsv1.Deployment {

	cpuRequest, err := resource.ParseQuantity(podInfo.Spec.Resources.CpuRequest)
	if err != nil {
		return nil
	}

	memLimit, err := resource.ParseQuantity(podInfo.Spec.Resources.MemoryLimit)
	if err != nil {
		return nil
	}

	cpuLimit := cpuRequest.DeepCopy()
	memRequest := memLimit.DeepCopy()

	cpuLimit.Set(cpuLimit.Value() * 2)
	memRequest.Set(memRequest.Value() / 2)

	println("Cpu Limit value is: " + strconv.FormatInt(cpuLimit.Value(), 10))
	println("Memory Request: " + strconv.FormatInt(memRequest.Value(), 10))

	podInfoDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "podinfo-" + podInfo.Name,
			Labels: map[string]string{
				"app":  "podinfo",
				"name": "podinfo-" + podInfo.Name,
			},
			Namespace: podInfo.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  "podinfo",
					"name": "podinfo-" + podInfo.Name,
				},
			},
			Replicas: &podInfo.Spec.ReplicaCount,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  "podinfo",
						"name": "podinfo-" + podInfo.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "podinfo-" + podInfo.Name,
							Image: podInfo.Spec.Image.Repository + ":" + podInfo.Spec.Image.Tag,
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    cpuRequest,
									corev1.ResourceMemory: memRequest,
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    cpuLimit,
									corev1.ResourceMemory: memLimit,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "PODINFO_UI_COLOR",
									Value: podInfo.Spec.Ui.Color,
								},
								{
									Name:  "PODINFO_UI_MESSAGE",
									Value: podInfo.Spec.Ui.Message,
								},
							},
						},
					},
				},
			}},
	}

	if podInfo.Spec.Redis.Enabled {
		podInfoDeployment.Spec.Template.Spec.Containers[0].Env = append(podInfoDeployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
			Name:  "PODINFO_CACHE_SERVER",
			Value: "tcp://redis-" + podInfo.Name + ":6379",
		})
	}
	return podInfoDeployment
}

func createRedisDeployment(myapp *groupv1alpha1.MyAppResource) *appsv1.Deployment {

	var replica int32 = 1
	redisDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "redis-" + myapp.Name,
			Labels: map[string]string{
				"app":  "redis",
				"name": "redis-" + myapp.Name,
			},
			Namespace: myapp.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":  "redis",
					"name": "redis-" + myapp.Name,
				},
			},
			Replicas: &replica,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":  "redis",
						"name": "redis-" + myapp.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "redis-" + myapp.Name,
							Image: "redis:latest",
							Ports: []corev1.ContainerPort{
								{
									Name:          "redis",
									ContainerPort: 6379,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}

	return redisDeployment
}

func createRedisService(myapp *groupv1alpha1.MyAppResource) *corev1.Service {
	redisService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "redis-" + myapp.Name,
			Labels: map[string]string{
				"app":  "redis",
				"name": "redis-" + myapp.Name,
			},
			Namespace: myapp.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":  "redis",
				"name": "redis-" + myapp.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:     "redis",
					Port:     6379,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}
	return redisService
}
