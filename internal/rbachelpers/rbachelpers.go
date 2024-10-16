// Copyright 2024 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rbachelpers

import (
	v1 "k8s.io/api/rbac/v1"
)

type RBACHelper struct {}

func (r *RBACHelper) IsServiceAccountBoundToRoleBindingList(clusterRoleBindings *v1.ClusterRoleBindingList, serviceAccountName string) bool {
	for _, roleBinding := range clusterRoleBindings.Items {
		if roleBinding.Subjects != nil {
			for _, subject := range roleBinding.Subjects {
				if subject.Kind == "ServiceAccount" && subject.Name == serviceAccountName {
					return true
				}
			}
		}
	}
	return false
}
