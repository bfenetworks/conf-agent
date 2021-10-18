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

package agent

import (
	"github.com/baidu/conf-agent/conf_reload"
	"github.com/baidu/conf-agent/config"
)

// The Agent keep reloaders.
// Agent Start will start all reloaders
type Agent struct {
	stop chan bool

	reloaders []*conf_reload.Reloader
}

// New create a Agent according to config
func New(rcs []*config.ReloaderConfig) (*Agent, error) {
	agent := &Agent{
		stop: make(chan bool),
	}
	for _, mc := range rcs {
		m, err := conf_reload.NewReloader(mc)
		if err != nil {
			return nil, err
		}

		agent.reloaders = append(agent.reloaders, m)
	}

	return agent, nil
}

func (agent *Agent) Start() {
	for _, reloader := range agent.reloaders {
		go reloader.Start()
	}

	<-agent.stop
}

func (agent *Agent) Stop() {
	agent.stop <- true
}
