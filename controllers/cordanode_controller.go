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
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cordav1 "github.com/orangesys/corda-node-operator/api/v1"
	"github.com/orangesys/corda-node-operator/pkg/utils"
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
		utils.CreateService(node)
		svc, err := utils.GetServce(node)
		if err != nil {
			if errors.IsNotFound(err) {
				log.Info("Wating corda service create.")
				return reconcile.Result{RequeueAfter: time.Second * 5}, nil
			}
			return reconcile.Result{}, err
		}
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			log.Info("Wating cordaNode External IP create.")
			return reconcile.Result{RequeueAfter: time.Second * 5}, nil
		}
		node.Status.ExternalIP = svc.Status.LoadBalancer.Ingress[0].IP

		if err := utils.CreateNodeInfoConfigMap(node); err != nil {
			return ctrl.Result{}, err
		}
		utils.CreateCertsSecret(node)
		utils.CreateCordaNodeDeployment(node)
	}

	return ctrl.Result{}, nil
}

//SetupWithManager ...
func (r *CordaNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cordav1.CordaNode{}).
		Complete(r)
}
