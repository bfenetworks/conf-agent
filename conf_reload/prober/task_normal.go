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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"

	"github.com/baidu/conf-agent/config"
	"github.com/baidu/conf-agent/xhttp"
	"github.com/baidu/conf-agent/xlog"
)

type NormalFileTask struct {
	config config.NormalFileTaskConfig

	commonConfig commonConfig
}

func NewNormalFileTask(c config.NormalFileTaskConfig) (*NormalFileTask, error) {
	return &NormalFileTask{
		config: c,
		commonConfig: commonConfig{
			BFECluster:      c.BFECluster,
			ConfTaskHeaders: c.ConfTaskHeaders,
			ConfTaskTimeout: c.ConfTaskTimeout,
		},
	}, nil
}

func (task *NormalFileTask) FetchConfFiles(ctx context.Context) ([]*FetchFileResult, error) {
	config := task.config
	fileName := config.ConfFileName

	localVersion, err := loadLocalVersion(path.Join(config.ConfDir, fileName))
	if err != nil {
		return nil, err
	}

	// obtain config data
	raw, err := obtainRemoteConfig(ctx, task.commonConfig, config.ConfAPI, localVersion)
	if err != nil {
		return nil, err
	}

	// if no newer config, conf server will return null
	if raw == nil || string(raw) == `null` {
		return nil, nil
	}

	version, err := calculateVersion(raw)
	if err != nil {
		return nil, err
	}

	return []*FetchFileResult{
		{
			Name:    fileName,
			Version: version,
			Content: raw,
		},
	}, nil
}

func obtainRemoteConfig(ctx context.Context, config commonConfig, apiURL, localVersion string) ([]byte, error) {
	/* response data look like:
	{
		"ErrNum": 200,
		"Data": {

		}
	}
	*/
	rsp := &struct {
		ErrNum int
		Data   json.RawMessage
	}{}

	params := url.Values{}
	params.Add("version", localVersion)
	params.Add("bfe_cluster", config.BFECluster)
	requestURL := apiURL + "?" + params.Encode()

	req := xhttp.NewHTTPRequest().
		Decorate(
			xhttp.HTTPRequestTimeoutOp(config.ConfTaskTimeout),
			xhttp.SimpleRequestOp(http.MethodGet, requestURL, nil),
			xhttp.HTTPRequestHeaderOp(config.ConfTaskHeaders)).
		Do().
		Decorate(
			xhttp.RspBodyRawReaderOp,
			xhttp.RspCode200Op,
			xhttp.RspBodyJSONReader(&rsp),
		)

	if err := req.Err(); err != nil {
		return nil, err
	}

	xlog.Default.Debug(
		xlog.InfoLogFormat(ctx, "obtainRemoteConfig", "url: ", requestURL, " fileContent: ", string(req.RawContent)))

	return rsp.Data, nil
}

func loadLocalVersion(fileName string) (string, error) {
	bs, err := ioutil.ReadFile(fileName)
	if os.IsNotExist(err) {
		return "", nil
	}

	version, err := calculateVersion(bs)
	if err != nil {
		return "", fmt.Errorf("bad file content, file: %s, err: %v", fileName, err)
	}

	return version, nil
}

var regNumber = regexp.MustCompile("[^0-9]")

func justKeepNumber(s string) string {
	return regNumber.ReplaceAllString(s, "")
}

func calculateVersion(fileContent []byte) (string, error) {
	if len(fileContent) == 0 || bytes.Equal(fileContent, []byte("null")) {
		return "", nil
	}

	// all conf file content look like {"Version": "xxx", ....}
	tmp := struct {
		Version string
	}{}
	if err := json.Unmarshal(fileContent, &tmp); err != nil {
		return "", err
	}

	version := justKeepNumber(tmp.Version)
	if version == "" {
		version = "00000000000000"
	}

	return version, nil
}
