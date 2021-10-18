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

package trigger

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/baidu/conf-agent/config"
	"github.com/baidu/conf-agent/xhttp"
	"github.com/baidu/conf-agent/xlog"
)

type Trigger struct {
	c config.TriggerConfig
}

func NewTrigger(c config.TriggerConfig) (*Trigger, error) {
	return &Trigger{
		c: c,
	}, nil
}

func (trigger *Trigger) TriggerBFEReload(ctx context.Context, version string) error {
	confDir := trigger.c.ConfDir + "_" + version

	query := url.Values{}
	query.Add("path", confDir)
	api := fmt.Sprintf("%s?%s", trigger.c.BFEReloadAPI, query.Encode())

	// succ relaod rsp: {"error":null}
	rsp := &struct {
		Error string `json:"error"`
	}{}

	req := xhttp.NewHTTPRequest().
		Decorate(
			xhttp.HTTPRequestTimeoutOp(trigger.c.BFEReloadTimeout),
			xhttp.SimpleRequestOp(http.MethodGet, api, nil),
		).
		Do().
		Decorate(
			xhttp.RspBodyRawReaderOp,
			xhttp.RspCode200Op,
			xhttp.RspBodyJSONReader(rsp),
		)

	err := req.Err()
	if err != nil {
		xlog.Default.Error(xlog.ErrLogFormat(ctx, "reload_bfe", err))
		return err
	}

	if rsp.Error != "" {
		err = fmt.Errorf("reload fail, rsp: %s", string(req.RawContent))
		xlog.Default.Error(xlog.ErrLogFormat(ctx, "reload_bfe", err))
		return err
	}

	return nil
}
