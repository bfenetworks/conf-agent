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
	"net/http/httptest"
	"testing"
	"time"

	"github.com/baidu/conf-agent/config"
)

func TestTrigger_TriggerBFEReload(t *testing.T) {
	tests := []struct {
		name string

		BFEReloadAPI     string
		BFEReloadTimeout time.Duration

		tmpConfDir string

		cost time.Duration
		rps  string

		wantErr bool
	}{
		{
			name:             "case1_success",
			BFEReloadTimeout: time.Second,
			rps:              `{"error":null}`,
			tmpConfDir:       "mod_tls",
		},
		{
			name:             "case1_fail_rsp",
			BFEReloadTimeout: time.Second,
			rps:              `{"error":"some error"}`,
			tmpConfDir:       "mod_tls",
			wantErr:          true,
		},
		{
			name:             "case1_fail_timeout",
			BFEReloadTimeout: time.Microsecond,
			rps:              `{"error":"some error"}`,
			tmpConfDir:       "mod_tls",
			wantErr:          true,
			cost:             time.Millisecond,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.cost > 0 {
					time.Sleep(tt.cost)
				}
				fmt.Fprintln(w, tt.rps)
			}))
			defer ts.Close()

			trigger := &Trigger{
				c: config.TriggerConfig{
					BFEReloadAPI: ts.URL,
				},
			}
			if err := trigger.TriggerBFEReload(context.TODO(), tt.tmpConfDir); (err != nil) != tt.wantErr {
				t.Errorf("Trigger.TriggerBFEReload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
