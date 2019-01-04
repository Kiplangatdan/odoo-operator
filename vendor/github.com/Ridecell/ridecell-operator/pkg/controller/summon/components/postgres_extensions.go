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

	"github.com/pkg/errors"
	postgresv1 "github.com/zalando-incubator/postgres-operator/pkg/apis/acid.zalan.do/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	dbv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/db/v1beta1"
	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/components"
)

type postgresExtensionsComponent struct{}

func NewPostgresExtensions() *postgresExtensionsComponent {
	return &postgresExtensionsComponent{}
}

func (comp *postgresExtensionsComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{
		&dbv1beta1.PostgresExtension{},
	}
}

func (_ *postgresExtensionsComponent) IsReconcilable(ctx *components.ComponentContext) bool {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)
	if instance.Status.PostgresStatus != postgresv1.ClusterStatusRunning {
		// Database not ready yet.
		return false
	}
	return true
}

func (_ *postgresExtensionsComponent) Reconcile(ctx *components.ComponentContext) (reconcile.Result, error) {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)

	var existingPostgis *dbv1beta1.PostgresExtension
	var existingPostgisToplogy *dbv1beta1.PostgresExtension

	// Create the postgis extension.
	res, _, err := ctx.CreateOrUpdate("postgres_extensions/postgis.yml.tpl", nil, func(goalObj, existingObj runtime.Object) error {
		goal := goalObj.(*dbv1beta1.PostgresExtension)
		existingPostgis = existingObj.(*dbv1beta1.PostgresExtension)
		// Copy the Spec over.
		existingPostgis.Spec = goal.Spec
		return nil
	})
	if err != nil {
		return res, errors.Wrap(err, "Error from postgis extension")
	}

	// Create the postgis_topology extension.
	res, _, err = ctx.CreateOrUpdate("postgres_extensions/postgis_topology.yml.tpl", nil, func(goalObj, existingObj runtime.Object) error {
		goal := goalObj.(*dbv1beta1.PostgresExtension)
		existingPostgisToplogy = existingObj.(*dbv1beta1.PostgresExtension)
		// Copy the Spec over.
		existingPostgisToplogy.Spec = goal.Spec
		return nil
	})
	if err != nil {
		return res, errors.Wrap(err, "Error from postgis_topology extension")
	}

	// Figure out status-y things.
	if existingPostgis.Status.Status == dbv1beta1.StatusError {
		// Postgis error'd, grab its message and error the whole thing.
		instance.Status.Status = summonv1beta1.StatusError
		instance.Status.PostgresExtensionStatus = summonv1beta1.StatusError
		instance.Status.Message = fmt.Sprintf("postgis: %s", existingPostgis.Status.Message)
	} else if existingPostgisToplogy.Status.Status == dbv1beta1.StatusError {
		// Postgis_topology, same as above but with a different error message (hopefully).
		instance.Status.Status = summonv1beta1.StatusError
		instance.Status.PostgresExtensionStatus = summonv1beta1.StatusError
		instance.Status.Message = fmt.Sprintf("postgis_topology: %s", existingPostgisToplogy.Status.Message)
	} else if existingPostgis.Status.Status == dbv1beta1.StatusReady && existingPostgisToplogy.Status.Status == dbv1beta1.StatusReady {
		// Both are ready, we're good to go!
		instance.Status.PostgresExtensionStatus = summonv1beta1.StatusReady
	} else {
		// Something else, probably still initial setup where upstream status is still "".
		instance.Status.PostgresExtensionStatus = ""
	}

	return reconcile.Result{}, nil
}
