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

package v1beta1

// OdooClusterStatusConditionType ...
type OdooClusterStatusConditionType string

const (
	// OdooClusterStatusConditionTypeCreated ...
	OdooClusterStatusConditionTypeCreated OdooClusterStatusConditionType = "Created"
	// OdooClusterStatusConditionTypeReconciled ...
	OdooClusterStatusConditionTypeReconciled OdooClusterStatusConditionType = "Reconciled"
	// OdooClusterStatusConditionTypeAppSecretLoaned ...
	OdooClusterStatusConditionTypeAppSecretLoaned OdooClusterStatusConditionType = "AppSecretLoaned"
	// OdooClusterStatusConditionTypePullSecretLoaned ...
	OdooClusterStatusConditionTypePullSecretLoaned OdooClusterStatusConditionType = "PullSecretLoaned"
	// OdooClusterStatusConditionTypeErrored ...
	OdooClusterStatusConditionTypeErrored OdooClusterStatusConditionType = "Errored"
)

// OdooVersionStatusConditionType ...
type OdooVersionStatusConditionType string

const (
	// OdooVersionStatusConditionTypeDeployed ...
	OdooVersionStatusConditionTypeDeployed OdooVersionStatusConditionType = "Deployed"
	// OdooVersionStatusConditionTypeApplied ...
	OdooVersionStatusConditionTypeApplied OdooVersionStatusConditionType = "Applied"
	// OdooVersionStatusConditionTypeReconciled ...
	OdooVersionStatusConditionTypeReconciled OdooVersionStatusConditionType = "Reconciled"
)
