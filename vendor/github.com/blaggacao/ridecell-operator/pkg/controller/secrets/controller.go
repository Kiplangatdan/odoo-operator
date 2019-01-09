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

package secrets

import (
	secretsv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/secrets/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/Ridecell/ridecell-operator/pkg/components"
	secretscomponents "github.com/Ridecell/ridecell-operator/pkg/controller/secrets/components"
)

// Add creates a new PullSecret Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	_, err := components.NewReconciler("pull-secrets-controller", mgr, &secretsv1beta1.PullSecret{}, nil, []components.Component{
		secretscomponents.NewDefaults(),
		secretscomponents.NewSecret(),
	})
	return err
}
