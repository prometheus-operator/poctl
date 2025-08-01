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
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/golden"
)

func TestPrintManifest(t *testing.T) {
	for _, tc := range []struct {
		input    string
		ruleFile string

		golden string
	}{
		{
			input:  "rule.yaml",
			golden: "rule.golden",
		},
		{
			input:  "not_a_rule.yaml",
			golden: "not_a_rule.golden",
		},
		{
			input: "invalid.yaml",
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", tc.input))
			if err != nil {
				t.Fatalf("failed to open input file: %s", err)
				return
			}

			w := &bytes.Buffer{}
			if err = PrintManifest(f, w); err != nil {
				if tc.golden != "" {
					t.Fatalf("expected no error, got %s", err)
				}
				return
			}

			golden.Assert(t, w.String(), tc.golden)
		})
	}
}
