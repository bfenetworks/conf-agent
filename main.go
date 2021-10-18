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

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/baidu/conf-agent/agent"
	"github.com/baidu/conf-agent/config"
	"github.com/baidu/conf-agent/xlog"
	"github.com/baidu/conf-agent/version"
)

var (
	help     *bool   = flag.Bool("h", false, "to show help")
	showVer  *bool   = flag.Bool("v", false, "to show version")
	confDir  *string = flag.String("c", "./conf/", "API configure dir")
	confFile *string = flag.String("cf", "conf-agent.toml", "API configure file")
)

func main() {
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		return
	}
	if *showVer {
		fmt.Printf("version %s\n", version.Version)
		return
	}

	exit := func(err error) {
		fmt.Println(err)
		os.Exit(-1)
	}

	conf, err := config.Init(filepath.Join(*confDir, *confFile))
	if err != nil {
		exit(err)
	}

	if err := xlog.Init(conf.Logger); err != nil {
		exit(err)
	}

	agent, err := agent.New(conf.Reloaders)
	if err != nil {
		exit(err)
	}

	agent.Start()
}
