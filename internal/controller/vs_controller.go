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
	"strings"

	networkingv1a3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// VirtualServicdReconciler reconciles a Guestbook object
type VirtualServicdReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *VirtualServicdReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Reconciling VirtualService", "name", req.NamespacedName)

	vs := networkingv1a3.VirtualService{}
	if err := r.Get(ctx, req.NamespacedName, &vs); err != nil {
		logger.Error(err, "unable to fetch DestinationRule")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	for _, gwName := range vs.Spec.Gateways {
		// try to resolve gateway namespaced name
		gwNN := resolveGatewayName(gwName, req.NamespacedName)
		gw := networkingv1a3.Gateway{}
		if err := r.Get(ctx, gwNN, &gw); err != nil {
			logger.Error(err, "unable to fetch Gateway")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VirtualServicdReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1a3.VirtualService{}).
		Complete(r)
}

// resolveGatewayName uses metadata information to resolve a reference
// to shortname of the gateway to FQDN
// copy from https://github.com/istio/istio/blob/d2770ba97eb2c3f19e7013ca4f3c05075dbc8433/pilot/pkg/model/config.go#L249
func resolveGatewayName(gwname string, meta types.NamespacedName) types.NamespacedName {
	// New way of binding to a gateway in remote namespace
	// is ns/name. Old way is either FQDN or short name
	if !strings.Contains(gwname, "/") {
		if !strings.Contains(gwname, ".") {
			// we have a short name. Resolve to a gateway in same namespace
			return types.NamespacedName{
				Namespace: meta.Namespace,
				Name:      gwname,
			}
		} else {
			// parse namespace from FQDN. This is very hacky, but meant for backward compatibility only
			// This is a legacy FQDN format. Transform name.ns.svc.cluster.local -> ns/name
			i := strings.Index(gwname, ".")
			fqdn := strings.Index(gwname[i+1:], ".")
			if fqdn == -1 {
				return types.NamespacedName{
					Namespace: gwname[i+1:],
					Name:      gwname[:i],
				}
			} else {
				return types.NamespacedName{
					Namespace: gwname[i+1 : i+1+fqdn],
					Name:      gwname[:i],
				}
			}
		}
	} else {
		// remove the . from ./gateway and substitute it with the namespace name
		i := strings.Index(gwname, "/")
		if gwname[:i] == "." {
			return types.NamespacedName{
				Namespace: meta.Namespace,
				Name:      gwname[i+1:],
			}
		}
	}

	return types.NamespacedName{
		Namespace: meta.Namespace,
		Name:      gwname,
	}
}
