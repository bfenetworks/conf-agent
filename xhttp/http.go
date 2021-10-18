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

package xhttp

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type HTTPRequest struct {
	Client *http.Client

	Request *http.Request

	Response   *http.Response
	RawContent []byte

	err error
}

func RspCode200Op(h *HTTPRequest) error {
	if statusCode := h.Response.StatusCode; statusCode != 200 {
		return fmt.Errorf("bad StatuCode: %d, Raw: %s", statusCode, h.RawContent)
	}
	return nil
}

type HTTPRequestOp func(*HTTPRequest) error

func RspBodyRawReaderOp(hr *HTTPRequest) error {
	if hr.Response == nil {
		return fmt.Errorf("body is nil")
	}

	hr.RawContent, hr.err = ioutil.ReadAll(hr.Response.Body)
	defer hr.Response.Body.Close()

	return hr.err
}

func RspBodyJSONReader(ds ...interface{}) HTTPRequestOp {
	return func(hr *HTTPRequest) error {
		if hr.err != nil {
			return hr.err
		}
		if hr.RawContent == nil {
			if err := RspBodyRawReaderOp(hr); err != nil {
				return err
			}
		}

		for _, d := range ds {
			if err := json.Unmarshal(hr.RawContent, d); err != nil {
				return fmt.Errorf("json.Unmarshal fail, err: %v,  raw: %v", err, string(hr.RawContent))
			}
		}
		return nil
	}
}

func HTTPRequestHeaderOp(header map[string]string) HTTPRequestOp {
	return func(hr *HTTPRequest) error {
		for k, v := range header {
			hr.Request.Header.Add(k, v)
		}
		return nil
	}
}

func SimpleRequestOp(method, url string, body io.Reader) HTTPRequestOp {
	return func(hr *HTTPRequest) error {
		var err error
		hr.Request, err = http.NewRequest(http.MethodGet, url, body)
		return err
	}
}

func HTTPRequestTimeoutOp(timeout time.Duration) HTTPRequestOp {
	return func(hr *HTTPRequest) error {
		if timeout < 1 {
			return nil
		}

		hr.Client = &http.Client{
			Timeout: timeout,
		}

		return nil
	}
}

var defaultClient = &http.Client{
	Timeout: 10 * time.Second,
}

func NewHTTPRequest() *HTTPRequest {
	return &HTTPRequest{
		Client: defaultClient,
	}
}

func (hr *HTTPRequest) Decorate(ops ...HTTPRequestOp) *HTTPRequest {
	for _, op := range ops {
		if hr.err != nil {
			break
		}

		hr.err = op(hr)
	}
	return hr
}

func (hr *HTTPRequest) Do() *HTTPRequest {
	if hr.err != nil {
		return hr
	}

	hr.Response, hr.err = hr.Client.Do(hr.Request)
	return hr
}

func (hr *HTTPRequest) Err() error {
	if hr.err == nil {
		return nil
	}

	if hr.Request != nil {
		return fmt.Errorf("url: %s, err: %v", hr.Request.URL.String(), hr.err)
	}

	return hr.err
}
