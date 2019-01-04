/*
Copyright 2018 Ridecell, Inc.

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

package components

import (
	"fmt"
	"io/ioutil"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	secretsv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/secrets/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/components"
)

const inClusterNamespacePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

type pullSecretComponent struct{}

func NewSecret() *pullSecretComponent {
	return &pullSecretComponent{}
}

func (comp *pullSecretComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{
		&corev1.Secret{},
	}
}

func (_ *pullSecretComponent) IsReconcilable(_ *components.ComponentContext) bool {
	// Secrets have no dependencies, always reconcile.
	return true
}

func (comp *pullSecretComponent) Reconcile(ctx *components.ComponentContext) (reconcile.Result, error) {
	instance := ctx.Top.(*secretsv1beta1.PullSecret)

	operatorNamespace := os.Getenv("NAMESPACE")
	if operatorNamespace == "" {
		var err error
		operatorNamespace, err = getInClusterNamespace()
		if err != nil {
			instance.Status.Status = secretsv1beta1.StatusError
			return reconcile.Result{}, err
		}
	}

	target := &corev1.Secret{}
	err := ctx.Get(ctx.Context, types.NamespacedName{Name: instance.Spec.PullSecretName, Namespace: operatorNamespace}, target)
	if err != nil {
		if errors.IsNotFound(err) {
			instance.Status.Status = secretsv1beta1.StatusErrorSecretNotFound
		} else {
			instance.Status.Status = secretsv1beta1.StatusError
		}
		return reconcile.Result{Requeue: true}, err
	}

	fetchTarget := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: instance.Spec.PullSecretName, Namespace: instance.Namespace}}
	_, err = controllerutil.CreateOrUpdate(ctx.Context, ctx, fetchTarget, func(existingObj runtime.Object) error {
		existing := existingObj.(*corev1.Secret)
		// Set owner ref.
		err := controllerutil.SetControllerReference(instance, existing, ctx.Scheme)
		if err != nil {
			instance.Status.Status = secretsv1beta1.StatusError
			return err
		}
		// Sync important fields.
		existing.ObjectMeta.Labels = target.ObjectMeta.Labels
		existing.ObjectMeta.Annotations = target.ObjectMeta.Annotations
		existing.Type = target.Type
		existing.Data = target.Data
		return nil
	})
	if err != nil {
		instance.Status.Status = secretsv1beta1.StatusError
		return reconcile.Result{Requeue: true}, err
	}

	instance.Status.Status = secretsv1beta1.StatusReady
	return reconcile.Result{}, nil
}

func getInClusterNamespace() (string, error) {
	// Check whether the namespace file exists.
	// If not, we are not running in cluster so can't guess the namespace.
	_, err := os.Stat(inClusterNamespacePath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("not running in-cluster, please specify $NAMESPACE")
	} else if err != nil {
		return "", fmt.Errorf("error checking namespace file: %v", err)
	}

	// Load the namespace file and return itss content
	namespace, err := ioutil.ReadFile(inClusterNamespacePath)
	if err != nil {
		return "", fmt.Errorf("error reading namespace file: %v", err)
	}
	return string(namespace), nil
}
