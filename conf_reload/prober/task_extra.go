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
	"fmt"
	"net/http"
	"strings"

	"github.com/baidu/conf-agent/config"
	"github.com/baidu/conf-agent/xhttp"
	"github.com/baidu/conf-agent/xlog"
	"github.com/ohler55/ojg/oj"
)

type ExtraFileTask struct {
	config config.ExtraFileTaskConfig

	normalFileTask *NormalFileTask
}

func NewExtraFileTask(c config.ExtraFileTaskConfig) (*ExtraFileTask, error) {
	np, err := NewNormalFileTask(c.NormalFileTaskConfig)
	if err != nil {
		return nil, err
	}

	return &ExtraFileTask{
		config: c,

		normalFileTask: np,
	}, nil
}

func (task *ExtraFileTask) FetchConfFiles(ctx context.Context) ([]*FetchFileResult, error) {
	fileList, err := task.normalFileTask.FetchConfFiles(ctx)
	if err != nil {
		return nil, err
	}

	if len(fileList) == 0 {
		return fileList, err
	}

	// analysis file content, obtain extra files
	extraFiles, err := task.obtainExtraFiles(ctx, fileList[0].Content)
	if err != nil {
		return nil, err
	}

	for remotePath, localPath := range extraFiles {
		fileContent, err := task.obtainExtraFile(ctx, remotePath)
		if err != nil {
			return nil, err
		}

		fileList = append(fileList, &FetchFileResult{
			Name:    localPath,
			Content: fileContent,
		})
	}

	return fileList, err
}

// convert {module}_{version}/xxxx to {module}/xxxx and xxxx
func removeDirVersionInfo(fileName string) (remotePath, localPath string, err error) {
	slashIndex := strings.Index(fileName, "/")
	if slashIndex == -1 {
		return "", "", fmt.Errorf("want format {module}_{version}/xxxx")
	}

	moduleWithVersion := fileName[:slashIndex]
	underlineIndex := strings.LastIndex(moduleWithVersion, "_")
	if underlineIndex == -1 {
		return "", "", fmt.Errorf("want format {module}_{version}/xxxx")
	}

	if slashIndex == underlineIndex+1 {
		return "", "", fmt.Errorf("want format {module}_{version}/xxxx")
	}

	return moduleWithVersion[:underlineIndex] + fileName[slashIndex:], fileName[slashIndex+1:], nil
}

func (prober *ExtraFileTask) obtainExtraFiles(ctx context.Context, fileContent []byte) (map[string]string, error) {
	jsonData, err := oj.Parse(fileContent)
	if err != nil {
		err = fmt.Errorf("parse fail, content: %s, err: %v", string(fileContent), err)
		xlog.Default.Error(xlog.ErrLogFormat(ctx, "TaskExtraFile.parse", err))

		return nil, err
	}

	remotePath2localPath := map[string]string{}
	for _, pattern := range prober.config.JSONPaths {
		results := pattern.Get(jsonData)

		for _, result := range results {
			fileName := fmt.Sprintf("%v", result)
			remote, local, err := removeDirVersionInfo(fileName)
			if err != nil {
				return nil, err
			}

			remotePath2localPath[remote] = local
		}
	}

	return remotePath2localPath, nil
}

func (prober *ExtraFileTask) obtainExtraFile(ctx context.Context, fileName string) (raw []byte, err error) {
	config := prober.config

	req := xhttp.NewHTTPRequest().
		Decorate(
			xhttp.SimpleRequestOp(http.MethodGet, config.ExtraFileServer+fileName, nil),
			xhttp.HTTPRequestTimeoutOp(config.ExtraFileTaskTimeout),
			xhttp.HTTPRequestHeaderOp(config.ExtraFileTaskHeaders)).
		Do().
		Decorate(
			xhttp.RspBodyRawReaderOp,
			xhttp.RspCode200Op,
		)

	err = req.Err()
	if err != nil {
		return
	}

	raw = req.RawContent

	return
}
