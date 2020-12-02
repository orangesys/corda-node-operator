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
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cordav1 "orangesys.io/cordanode/api/v1"
)

// CordaNodeReconciler reconciles a CordaNode object
type CordaNodeReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const (
	cordaNodeFinalizer string = "cordanode.finalizers.corda.orangesys.io"
	searchKey          string = ".metadata.controller"
)

// +kubebuilder:rbac:groups=corda.orangesys.io,resources=cordanodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=corda.orangesys.io,resources=cordanodes/status,verbs=get;update;patch

func (r *CordaNodeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("CordaNode", req.NamespacedName)

	node := &cordav1.CordaNode{}

	if err := r.Get(ctx, req.NamespacedName, node); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if node.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(node.ObjectMeta.Finalizers, cordaNodeFinalizer) {
			node.ObjectMeta.Finalizers = append(node.ObjectMeta.Finalizers, cordaNodeFinalizer)
			if err := r.Update(ctx, node); err != nil {
				return ctrl.Result{}, err
			}
		}

		podLabels := map[string]string{
			"app": req.Name,
		}
		dp := appv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.Name,
				Namespace: req.Namespace,
			},
			Spec: appv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: podLabels,
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: podLabels,
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:            req.Name,
							Image:           node.Spec.Image,
							ImagePullPolicy: "Always",
							Ports: []corev1.ContainerPort{{
								ContainerPort: 80,
							}},
						}},
					},
				},
			},
		}
		oldDeployment := &appv1.Deployment{}
		if err := r.Get(ctx, req.NamespacedName, oldDeployment); err != nil {
			if client.IgnoreNotFound(err) == nil {
				if err := r.Create(ctx, &dp); err != nil {
					log.Error(err, "Create Node Failed")
					return ctrl.Result{}, err
				}
			} else {
				return ctrl.Result{}, err
			}
		}

		if err := r.Update(ctx, &dp); err != nil {
			return ctrl.Result{}, err
		}

	} else {
		if containsString(node.ObjectMeta.Finalizers, cordaNodeFinalizer) {
			log.Info("Start delete deployment")
			deployment := &appv1.Deployment{}
			if err := r.Get(ctx, req.NamespacedName, deployment); client.IgnoreNotFound(err) != nil {
				return ctrl.Result{}, err
			}
			if err := r.Delete(ctx, deployment); err != nil {
				return ctrl.Result{}, err
			}
			node.ObjectMeta.Finalizers = removeString(node.ObjectMeta.Finalizers, cordaNodeFinalizer)
			if err := r.Update(ctx, node); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *CordaNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// if err := mgr.GetFieldIndexer().IndexField(&cordav1.CordaNode{}, searchKey, func(rawObj runtime.Object) []string {
	// 	node := rawObj.(*appv1.Deployment)
	// 	owner := metav1.GetControllerOf(node)
	// 	if owner == nil {
	// 		return nil
	// 	}
	// 	if owner.APIVersion != cordav1.GroupVersion.String() || owner.Kind != "CordaNode" {
	// 		return nil
	// 	}

	// 	return []string{owner.Name}
	// }); err != nil {
	// 	return err
	// }

	return ctrl.NewControllerManagedBy(mgr).
		For(&cordav1.CordaNode{}).
		Complete(r)
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
