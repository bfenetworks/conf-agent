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

package prober

import (
	"context"
	"encoding/json"
	"path"

	"github.com/baidu/conf-agent/config"
	"github.com/baidu/conf-agent/xlog"
)

type MultiKeyFileTask struct {
	config config.MultiJSONKeyFileTaskConfig

	commonConfig commonConfig
}

func NewMultiKeyFileTask(c config.MultiJSONKeyFileTaskConfig) (*MultiKeyFileTask, error) {
	return &MultiKeyFileTask{
		config: c,
		commonConfig: commonConfig{
			BFECluster:      c.BFECluster,
			ConfTaskHeaders: c.ConfTaskHeaders,
			ConfTaskTimeout: c.ConfTaskTimeout,
		},
	}, nil
}

func (task *MultiKeyFileTask) FetchConfFiles(ctx context.Context) ([]*FetchFileResult, error) {
	config := task.config

	localVersion := ""
	for _, fileName := range config.Key2ConfFile {
		version, err := loadLocalVersion(path.Join(config.ConfDir, fileName))
		if err != nil {
			return nil, err
		}

		if version > localVersion {
			localVersion = version
		}
	}

	// obtain config data
	raw, err := obtainRemoteConfig(ctx, task.commonConfig, config.ConfAPI, localVersion)
	if err != nil {
		return nil, err
	}

	// if no newer config, conf server will return null
	if raw == nil {
		return nil, nil
	}

	rawMap := map[string]json.RawMessage{}
	if err = json.Unmarshal(raw, &rawMap); err != nil {
		xlog.Default.Error(xlog.ErrLogFormat(ctx, "obtainRemoteConfig.Unmarshal", err))
		return nil, err
	}
	if rawMap == nil {
		return nil, nil
	}

	var fileList []*FetchFileResult
	for key, fileName := range config.Key2ConfFile {
		fileContent, ok := rawMap[key]
		if !ok {
			xlog.Default.Info(xlog.InfoLogFormat(ctx, "Key2ConfFile key not exist ", key))
			continue
		}

		version, err := calculateVersion(fileContent)
		if err != nil {
			return nil, err
		}

		fileList = append(fileList, &FetchFileResult{
			Name:    fileName,
			Version: version,
			Content: fileContent,
		})
	}

	return fileList, nil

}
