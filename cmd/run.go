// Copyright 2022 Snyk Ltd
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
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/snyk/policy-engine/pkg/data"
	"github.com/snyk/policy-engine/pkg/engine"
	"github.com/snyk/policy-engine/pkg/input"
	"github.com/snyk/policy-engine/pkg/metrics"
	"github.com/snyk/policy-engine/pkg/postprocess"
	"github.com/snyk/policy-engine/pkg/snapshot_testing"
	"github.com/spf13/afero"
	"github.com/spf13/afero/tarfs"
	"github.com/spf13/cobra"
)

var (
	runCmdRules   []string
	runVarFiles   []string
	runCmdWorkers *int
)

var runCmd = &cobra.Command{
	Use:   "run [-d <rules/metadata>...] [-r <rule ID>...] <input> [input...]",
	Short: "Policy Engine",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := cmdLogger()
		snapshot_testing.GlobalRegisterNoop()
		m := metrics.NewLocalMetrics(logger)
		selectedRules := map[string]bool{}
		for _, k := range runCmdRules {
			selectedRules[k] = true
		}
		providers := []data.Provider{
			data.PureRegoLibProvider(),
		}
		for _, path := range rootCmdRegoPaths {
			if isTgz(path) {
				f, err := os.Open(path)
				if err != nil {
					return err
				}
				providers = append(providers, data.TarGzProvider(f))
			} else {
				providers = append(providers, data.LocalProvider(path))
			}
		}
		detector, err := input.DetectorByInputTypes(
			input.Types{input.Auto},
		)
		if err != nil {
			return err
		}
		loader := input.NewLoader(detector)
		fsys := afero.OsFs{}
		for _, p := range args {
			var detectable input.Detectable
			if isTgz(p) {
				f, err := fsys.Open(p)
				if err != nil {
					return err
				}
				gzf, err := gzip.NewReader(f)
				if err != nil {
					return err
				}
				fsys := tarfs.New(tar.NewReader(gzf))
				detectable = &input.Directory{
					Path: ".",
					Fs:   fsys,
				}
			} else {
				var err error
				detectable, err = input.NewDetectable(fsys, p)
				if err != nil {
					return err
				}
			}
			loaded, err := loader.Load(detectable, input.DetectOptions{
				VarFiles: runVarFiles,
			})
			if err != nil {
				return err
			}
			if loaded {
				continue
			}
			if dir, ok := detectable.(*input.Directory); ok {
				walkFunc := func(d input.Detectable, depth int) (bool, error) {
					return loader.Load(d, input.DetectOptions{
						VarFiles: runVarFiles,
					})
				}
				if err := dir.Walk(walkFunc); err != nil {
					return err
				}
			}
		}
		ctx := context.Background()
		for path, errs := range loader.Errors() {
			for _, err := range errs {
				logger.Warn(ctx, fmt.Sprintf("%s: %s", path, err))
			}
		}
		states := loader.ToStates()
		eng, err := engine.NewEngine(ctx, &engine.EngineOptions{
			Providers: providers,
			RuleIDs:   selectedRules,
			Logger:    logger,
			Metrics:   m,
		})
		if err != nil {
			return err
		}
		results := eng.Eval(ctx, &engine.EvalOptions{
			Inputs:  states,
			Workers: *runCmdWorkers,
		})
		postprocess.AddSourceLocs(results, loader)

		bytes, err := json.MarshalIndent(results, "  ", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "%s\n", string(bytes))
		m.Log(ctx)
		return nil
	},
}

func init() {
	runCmdWorkers = runCmd.PersistentFlags().IntP("workers", "w", 0, "Number of workers. When 0 (the default) will use num CPUs + 1.")
	runCmd.PersistentFlags().StringSliceVarP(&runCmdRules, "rule", "r", runCmdRules, "Select specific rules")
	runCmd.PersistentFlags().StringSliceVar(&runVarFiles, "var-file", runVarFiles, "Pass in variable files")
}
