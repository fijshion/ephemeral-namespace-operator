/*
Copyright 2021.

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

	clowder "github.com/RedHatInsights/clowder/apis/cloud.redhat.com/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ClowdenvironmentReconciler reconciles a Clowdenvironment object
type ClowdenvironmentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=cloud.redhat.com,resources=clowdenvironments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud.redhat.com,resources=clowdenvironments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud.redhat.com,resources=clowdenvironments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Clowdenvironment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ClowdenvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	env := clowder.ClowdEnvironment{}
	if err := r.Client.Get(ctx, req.NamespacedName, &env); err != nil {
		r.Log.Error(err, "Error retrieving clowdenv", "env-name", env.Name)
		return ctrl.Result{}, err
	}

	r.Log.Info("Reconciling clowdenv", "env-name", env.Name)

	if ready, _ := VerifyClowdEnvReady(env); ready {
		if err := CreateFrontendEnv(ctx, r.Client, env.Status.TargetNamespace, env); err != nil {
			r.Log.Error(err, "Error creating frontend env", "ns-name", env.Status.TargetNamespace)
			UpdateAnnotations(ctx, r.Client, map[string]string{"status": "error"}, env.Status.TargetNamespace)
		}
		UpdateAnnotations(ctx, r.Client, map[string]string{"status": "ready"}, env.Status.TargetNamespace)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClowdenvironmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	return ctrl.NewControllerManagedBy(mgr).
		For(&clowder.ClowdEnvironment{}).
		WithEventFilter(poolFilter(ctx, r.Client)).
		Complete(r)
}

func poolFilter(ctx context.Context, cl client.Client) predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newObject := e.ObjectNew.(*clowder.ClowdEnvironment)

			return isOwnedByPool(ctx, cl, newObject.Status.TargetNamespace)
		},
		CreateFunc: func(e event.CreateEvent) bool {
			object := e.Object.(*clowder.ClowdEnvironment)

			return isOwnedByPool(ctx, cl, object.Status.TargetNamespace)
		},
	}
}

func isOwnedByPool(ctx context.Context, cl client.Client, nsName string) bool {
	ns, err := GetNamespace(ctx, cl, nsName)
	if err != nil {
		return false
	}
	for _, owner := range ns.GetOwnerReferences() {
		if owner.Kind == "Pool" {
			return true
		}
	}

	return false
}
