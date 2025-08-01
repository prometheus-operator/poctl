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

package cmd

import (
	"os"

	"github.com/prometheus-operator/poctl/internal/format"
	"github.com/spf13/cobra"
)

var formatCmd = &cobra.Command{
	Use:   "fmt {file | directory}",
	Short: "Format Prometheus Operator resources",
	Long:  `The format command in poctl formats PromQL expressions found in YAML manifests using the upstream Prometheus formatter.`,
	Args:  cobra.MatchAll(cobra.MinimumNArgs(1)),
	RunE:  runFormat,
}

func init() {
	rootCmd.AddCommand(formatCmd)

	//formatCmd.Flags().BoolP("in-place", "i", false, "Edit the files in place ")
}

func readDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		files = append(files, entry.Name())
	}

	return files, nil
}

func runFormat(_ *cobra.Command, args []string) error {
	var files []string

	for _, arg := range args {
		fi, err := os.Stat(arg)
		if err != nil {
			continue
		}

		if fi.IsDir() {
			dirFiles, err := readDir(arg)
			if err != nil {
				continue
			}

			files = append(files, dirFiles...)
			continue
		}

		files = append(files, arg)
	}

	for _, file := range files {
		if err := processFile(file); err != nil {
			return err
		}
	}

	return nil
}

func processFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	w := os.Stdout
	if err = format.PrintManifest(f, w); err != nil {
		return err
	}

	return nil
}
