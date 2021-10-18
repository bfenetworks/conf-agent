// Copyright (c) 2021 The BFE Authors.
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

package config

import (
	"fmt"
	"sort"
	"time"

	"github.com/go-playground/validator"
	"github.com/ohler55/ojg/jp"
)

type Config struct {
	Reloaders []*ReloaderConfig
	Logger    *LoggerConfig
}

type ReloaderConfig struct {
	Name string

	ConfDir        string
	ReloadInterval time.Duration

	Trigger TriggerConfig

	CopyFiles []string

	NormalFileTasks       []*NormalFileTaskConfig
	MultiJSONKeyFileTasks []*MultiJSONKeyFileTaskConfig
	ExtraFileFileTasks    []*ExtraFileTaskConfig
}

type NormalFileTaskConfig struct {
	BFECluster string

	ConfDir      string
	ConfAPI      string
	ConfFileName string

	ConfTaskHeaders map[string]string
	ConfTaskTimeout time.Duration
}

func newNormalFileTaskConfig(cf NormalFileTaskConfigFile, rcf ReloaderConfigFile) *NormalFileTaskConfig {
	return &NormalFileTaskConfig{
		BFECluster: rcf.BFECluster,
		ConfDir:    rcf.ConfDir,

		ConfAPI:      cf.ConfServer + cf.ConfAPI,
		ConfFileName: cf.ConfFileName,

		ConfTaskHeaders: cf.ConfTaskHeaders,
		ConfTaskTimeout: time.Duration(cf.ConfTaskTimeoutMs) * time.Millisecond,
	}
}

type MultiJSONKeyFileTaskConfig struct {
	BFECluster string

	ConfDir      string
	ConfAPI      string
	Key2ConfFile map[string]string

	ConfTaskHeaders map[string]string
	ConfTaskTimeout time.Duration
}

func newMultiJSONKeyFileTaskConfig(cf MultiJSONKeyFileTaskConfigFile, rcf ReloaderConfigFile) *MultiJSONKeyFileTaskConfig {
	return &MultiJSONKeyFileTaskConfig{
		BFECluster: rcf.BFECluster,

		ConfDir:      rcf.ConfDir,
		ConfAPI:      cf.ConfServer + cf.ConfAPI,
		Key2ConfFile: cf.Key2ConfFile,

		ConfTaskHeaders: cf.ConfTaskHeaders,
		ConfTaskTimeout: time.Duration(cf.ConfTaskTimeoutMs) * time.Millisecond,
	}
}

type ExtraFileTaskConfig struct {
	NormalFileTaskConfig

	ExtraFileServer      string
	ExtraFileTaskHeaders map[string]string
	ExtraFileTaskTimeout time.Duration

	// see https://goessner.net/articles/JsonPath/
	JSONPaths []jp.Expr `json:"-"`
}

type TriggerConfig struct {
	BFEReloadAPI     string
	BFEReloadTimeout time.Duration
	ConfDir          string
}

func newExtraFileTaskConfig(cf ExtraFileTaskConfigFile, rcf ReloaderConfigFile) (*ExtraFileTaskConfig, error) {
	patterns := []jp.Expr{}
	for _, path := range cf.ExtraFileJSONPaths {
		pattern, err := jp.ParseString(path)
		if err != nil {
			return nil, fmt.Errorf("ExtraFileJSONPaths %s compile fail, err: %v", path, err)
		}

		patterns = append(patterns, pattern)
	}

	return &ExtraFileTaskConfig{
		NormalFileTaskConfig: *newNormalFileTaskConfig(cf.NormalFileTaskConfigFile, rcf),

		ExtraFileServer:      cf.ExtraFileServer,
		ExtraFileTaskHeaders: cf.ExtraFileTaskHeaders,
		ExtraFileTaskTimeout: time.Duration(cf.ExtraFileTaskTimeoutMs) * time.Millisecond,

		JSONPaths: patterns,
	}, nil
}

func newReloaderConfig(rcf *ReloaderConfigFile, basic BasicFile) (*ReloaderConfig, error) {
	rc := &ReloaderConfig{
		Name:           rcf.name,
		ReloadInterval: time.Duration(rcf.ReloadIntervalMs) * time.Millisecond,
		ConfDir:        rcf.ConfDir,

		Trigger: TriggerConfig{
			BFEReloadAPI:     fmt.Sprintf("http://127.0.0.1:%d%s", basic.BFEMonitorPort, rcf.BFEReloadAPI),
			BFEReloadTimeout: time.Duration(rcf.BFEReloadTimeoutMs) * time.Millisecond,
			ConfDir:          rcf.ConfDir,
		},
		CopyFiles: rcf.CopyFiles,
	}

	for _, task := range rcf.NormalFileTasks {
		rc.NormalFileTasks = append(rc.NormalFileTasks, newNormalFileTaskConfig(task, *rcf))
	}

	for _, task := range rcf.MultiKeyFileTasks {
		rc.MultiJSONKeyFileTasks = append(rc.MultiJSONKeyFileTasks, newMultiJSONKeyFileTaskConfig(task, *rcf))
	}

	for _, task := range rcf.ExtraFileTasks {
		t, err := newExtraFileTaskConfig(task, *rcf)
		if err != nil {
			return nil, err
		}
		rc.ExtraFileFileTasks = append(rc.ExtraFileFileTasks, t)
	}

	return rc, nil
}

func Init(configFile string) (*Config, error) {
	config := &ConfigFile{
		Basic: BasicFile{
			BFEMonitorPort:     8421,
			BFEReloadTimeoutMs: 1500,
			BFEConfDir:         "/home/work/bfe/conf",

			ConfTaskTimeoutMs: 2500,

			ExtraFileTaskTimeoutMs: 2500,

			ReloadIntervalMs: 10000,
		},
	}

	if err := LoadConf(configFile, config); err != nil {
		return nil, err
	}

	for name, reloader := range config.Reloaders {
		reloader.name = name
		if err := reloader.merge(&config.Basic); err != nil {
			return nil, err
		}
	}

	if err := validator.New().Struct(config); err != nil {
		return nil, err
	}

	reloaders := []*ReloaderConfig{}
	for _, reloader := range config.Reloaders {
		rc, err := newReloaderConfig(reloader, config.Basic)
		if err != nil {
			return nil, err
		}

		reloaders = append(reloaders, rc)
	}
	sort.Slice(reloaders, func(i, j int) bool {
		return reloaders[i].Name < reloaders[j].Name
	})

	return &Config{
		Reloaders: reloaders,
		Logger:    &config.Logger,
	}, nil
}
