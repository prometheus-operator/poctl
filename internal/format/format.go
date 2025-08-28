// Copyright 2025 The prometheus-operator Authors
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

package format

import (
	"errors"
	"io"

	"github.com/prometheus/prometheus/promql/parser"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

// PrintManifest takes a YAML manifest as input and pretty-prints it.
func PrintManifest(r io.Reader, w io.Writer) error {
	dec := yaml.NewDecoder(r)
	for {
		// parse the YAML manifest.
		obj := &unstructured.Unstructured{
			Object: map[string]any{},
		}

		err := dec.Decode(obj.Object)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return err
		}

		if obj.Object == nil {
			continue
		}

		// Only PrometheusRule objects are formatted, others are printed as-is.
		if obj.GroupVersionKind().Group != monitoring.GroupName || obj.GroupVersionKind().Kind != monv1.PrometheusRuleKind {
			enc := yaml.NewEncoder(w)
			if err = enc.Encode(obj.Object); err != nil {
				return err
			}

			continue
		}

		groups, _, err := unstructured.NestedSlice(obj.Object, "spec", "groups")
		if err != nil {
			return err
		}

		for _, g := range groups {
			group := g.(map[string]any)
			rules, _, err := unstructured.NestedSlice(group, "rules")
			if err != nil {
				return err
			}

			for _, r := range rules {
				rule := r.(map[string]any)

				query, _, err := unstructured.NestedString(rule, "expr")
				if err != nil {
					return err
				}

				expr, err := parser.ParseExpr(query)
				if err != nil {
					return err
				}

				if err = unstructured.SetNestedField(rule, expr.Pretty(0), "expr"); err != nil {
					return err
				}
			}

			if err = unstructured.SetNestedSlice(group, rules, "rules"); err != nil {
				return err
			}
		}

		if err = unstructured.SetNestedSlice(obj.Object, groups, "spec", "groups"); err != nil {
			return err
		}

		// Prints the updated version.
		enc := yaml.NewEncoder(w)
		if err = enc.Encode(obj.Object); err != nil {
			return err
		}
	}

	return nil
}
