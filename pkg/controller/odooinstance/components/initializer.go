/*
 * This file is part of the Odoo-Operator (R) project.
 * Copyright (c) 2018-2018 XOE Corp. SAS
 * Authors: David Arnold, et al.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 *
 * ALTERNATIVE LICENCING OPTION
 *
 * You can be released from the requirements of the license by purchasing
 * a commercial license. Buying such a license is mandatory as soon as you
 * develop commercial activities involving the Odoo-Operator software without
 * disclosing the source code of your own applications. These activities
 * include: Offering paid services to a customer as an ASP, shipping Odoo-
 * Operator with a closed source product.
 *
 */

package components

import (
	"github.com/golang/glog"

	"github.com/blaggacao/ridecell-operator/pkg/components"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	// clusterv1beta1 "github.com/xoe-labs/odoo-operator/pkg/apis/cluster/v1beta1"
	instancev1beta1 "github.com/xoe-labs/odoo-operator/pkg/apis/instance/v1beta1"
)

type initializerComponent struct {
	templatePath string
}

func NewInitializer(templatePath string) *initializerComponent {
	return &initializerComponent{templatePath: templatePath}
}

// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get;update;patch
func (comp *initializerComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{
		&batchv1.Job{},
	}
}

func (_ *initializerComponent) IsReconcilable(ctx *components.ComponentContext) bool {
	instance := ctx.Top.(*instancev1beta1.OdooInstance)
	if instance.Spec.ParentHostname != nil {
		// The copier component is the one that should copy a parent instance
		return false
	}
	if instance.GetStatusCondition(instancev1beta1.OdooInstanceStatusConditionTypeCreated) != nil {
		// The instance is already created (or creating)
		return false
	}
	return true
}

func (comp *initializerComponent) Reconcile(ctx *components.ComponentContext) (reconcile.Result, error) {
	obj, err := ctx.GetTemplate(comp.templatePath, nil)
	if err != nil {
		return reconcile.Result{}, err
	}
	job := obj.(*batchv1.Job)
	instance := ctx.Top.(*instancev1beta1.OdooInstance)

	existing := &batchv1.Job{}
	err = ctx.Get(ctx.Context, types.NamespacedName{Name: job.Name, Namespace: job.Namespace}, existing)
	if err != nil && errors.IsNotFound(err) {
		glog.Infof("[%s/%s] initializer: Creating initializer Job %s/%s\n", instance.Namespace, instance.Name, job.Namespace, job.Name)

		// Setting the creating condition
		condition := instance.NewStatusCondition(instancev1beta1.OdooInstanceStatusConditionTypeCreated, corev1.ConditionFalse, "InitJobCreation",
			"An initializer Job has been launched to create and initialize this database instance.")
		instance.SetStatusCondition(*condition)

		// Launching the job
		err = controllerutil.SetControllerReference(instance, job, ctx.Scheme)
		if err != nil {
			return reconcile.Result{}, err
		}
		err = ctx.Create(ctx.Context, job)
		if err != nil {
			// If this fails, someone else might have started a initializer job between the Get and here, so just try again.
			return reconcile.Result{Requeue: true}, err
		}
		// Job is started, so we're done for now.
		return reconcile.Result{}, nil
	} else if err != nil {
		// Some other real error, bail.
		return reconcile.Result{}, err
	}

	// If we get this far, the job previously started at some point and might be done.
	// Check if the job succeeded.
	if existing.Status.Succeeded > 0 {
		// Success! Update the corresponding OdooInstanceStatusCondition and delete the job.

		glog.Infof("[%s/%s] initializer: Initializer Job succeeded, setting OdooInstanceStatusCondition \"Created\" to 'true'\n", instance.Namespace, instance.Name)
		condition := instance.NewStatusCondition(instancev1beta1.OdooInstanceStatusConditionTypeCreated, corev1.ConditionTrue, "InitJobSuccess",
			"The database instance has been sucessfully created by an initializer Job.")
		instance.SetStatusCondition(*condition)

		glog.V(2).Infof("[%s/%s] initializer: Deleting initializer Job %s/%s\n", instance.Namespace, instance.Name, existing.Namespace, existing.Name)
		err = ctx.Delete(ctx.Context, existing, client.PropagationPolicy(metav1.DeletePropagationBackground))
		if err != nil {
			return reconcile.Result{Requeue: true}, err
		}
	}

	// ... Or if the job failed.
	if existing.Status.Failed > 0 {
		glog.Errorf("[%s/%s] initializer: Initializer Job failed, leaving job %s/%s for debugging purposes\n", instance.Namespace, instance.Name, existing.Namespace, existing.Name)
	}

	// Job is still running, will get reconciled when it finishes.
	return reconcile.Result{}, nil
}
