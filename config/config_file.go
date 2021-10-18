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
	"path"
)

type BasicFile struct {
	// BFECluster is the BFECluster of current instance
	BFECluster string `validate:"required"`
	// ReloadIntervalMs is reload interval in ms
	ReloadIntervalMs int `validate:"min=1"`

	// BFEConfDir is the Dir of BFE Conf
	BFEConfDir string `validate:"required"`
	// BFEMonitorPort is the port of BFE Moinitor, agent will access this port to reload BFE
	BFEMonitorPort int `validate:"required"`
	// BFEReloadTimeoutMs is the timeout of reload BFE request
	BFEReloadTimeoutMs int `validate:"min=1"`

	// ConfServer is api server address
	ConfServer string `validate:"min=1"`
	// ConfTaskHeaders will be carry to api server
	// Authorization should be set
	ConfTaskHeaders map[string]string
	// ConfTaskTimeoutMs is the timeout of conf prober request
	ConfTaskTimeoutMs int `validate:"min=1"`

	// ExtraFileSever is Extra File address
	ExtraFileServer string `validate:"min=1"`
	// ExtraFileTaskHeaders will be carry to extra file server
	// Authorization should be set
	ExtraFileTaskHeaders map[string]string
	// ExtraFileTaskTimeoutMs is the timeout of extra file download request
	ExtraFileTaskTimeoutMs int `validate:"min=1"`
}

type ReloaderConfigFile struct {
	name string
	// BFECluster is the BFECluster of current instance, inherit BasicFile.BFECluster as default value
	BFECluster string `validate:"required"`

	// ConfDir is the reloadr conf dir, BasicFile.BFEConfDir join ConfDir is the conf root dir
	// inherit reloader map's key as default value
	ConfDir string `validate:"min=1"`
	// BFEReloadAPI is the reload api of bfe, with /reload/ prefix all the time
	BFEReloadAPI string `validate:"min=1"`

	// optional, inherit BasicConfig if not set
	BFEReloadTimeoutMs int `validate:"min=1"`
	ReloadIntervalMs   int `validate:"min=1"`

	// CopyFiles is the file/directory which will be copy from default conf dir to newer version conf dir
	// many conf can't fetch from conf file, newer version conf dir show inherit them so bfe can startup aftert stop
	CopyFiles []string

	// NormalFileTasks is the list of NormalFileTask
	// NormalFileTask meaning to conf file and  conf api one to one correspondence
	NormalFileTasks []NormalFileTaskConfigFile
	// MultiKeyFileTasks is the list of MultiKeyFileTask
	// MultiKeyFileTask meaning to conf file and  conf api many to one correspondence
	MultiKeyFileTasks []MultiJSONKeyFileTaskConfigFile
	// ExtraFileTasks is the los of ExtraFile
	// ExtraFile meaning to conf file and  conf api one to one correspondence
	// extra files info can be obtained by parse conf file
	ExtraFileTasks []ExtraFileTaskConfigFile
}

type NormalFileTaskConfigFile struct {
	// ConfAPI use to access to obtain conf file info
	ConfAPI string `validate:"required"`
	// ConfFileName is the local file name of this conf
	ConfFileName string `validate:"required"`

	// optional
	ConfServer        string `validate:"min=1"`
	ConfTaskHeaders   map[string]string
	ConfTaskTimeoutMs int `validate:"min=1"`
}

func (tf *NormalFileTaskConfigFile) merge(basic *BasicFile) {
	if tf.ConfServer == "" {
		tf.ConfServer = basic.ConfServer
	}

	if tf.ConfTaskHeaders == nil {
		tf.ConfTaskHeaders = basic.ConfTaskHeaders
	}

	if tf.ConfTaskTimeoutMs == 0 {
		tf.ConfTaskTimeoutMs = basic.ConfTaskTimeoutMs
	}
}

type ExtraFileTaskConfigFile struct {
	NormalFileTaskConfigFile

	// ExtraFileJSONPaths is the list of json path
	// the rule of obtain extra files name be defined by json path
	// see https://goessner.net/articles/JsonPath/
	ExtraFileJSONPaths []string

	// optional
	ExtraFileServer        string `validate:"min=1"`
	ExtraFileTaskHeaders   map[string]string
	ExtraFileTaskTimeoutMs int `validate:"min=1"`
}

func (tf *ExtraFileTaskConfigFile) merge(basic *BasicFile) {
	tf.NormalFileTaskConfigFile.merge(basic)

	if tf.ExtraFileServer == "" {
		tf.ExtraFileServer = basic.ExtraFileServer
	}

	if tf.ExtraFileTaskHeaders == nil {
		tf.ExtraFileTaskHeaders = basic.ExtraFileTaskHeaders
	}

	if tf.ExtraFileTaskTimeoutMs == 0 {
		tf.ExtraFileTaskTimeoutMs = basic.ExtraFileTaskTimeoutMs
	}
}

type MultiJSONKeyFileTaskConfigFile struct {
	// ConfAPI use to access to obtain conf file info
	ConfAPI string `validate:"required"`
	// Key2ConfFile is a map define the relation of the key of conf object and local file name
	Key2ConfFile map[string]string

	// optional
	ConfServer        string `validate:"min=1"`
	ConfTaskHeaders   map[string]string
	ConfTaskTimeoutMs int `validate:"min=1"`
}

func (tf *MultiJSONKeyFileTaskConfigFile) merge(basic *BasicFile) {
	if tf.ConfServer == "" {
		tf.ConfServer = basic.ConfServer
	}

	if tf.ConfTaskHeaders == nil {
		tf.ConfTaskHeaders = basic.ConfTaskHeaders
	}

	if tf.ConfTaskTimeoutMs == 0 {
		tf.ConfTaskTimeoutMs = basic.ConfTaskTimeoutMs
	}
}

type LoggerConfig struct {
	LogDir      string `validate:"required,min=1"`
	LogName     string `validate:"required,min=1"`
	LogLevel    string `validate:"required,oneof=DEBUG TRACE INFO WARNING ERROR CRITICAL"`
	RotateWhen  string `validate:"required,oneof=M H D MIDNIGHT"` // rotate time
	BackupCount int    `validate:"required,min=1"`                // backup files
	Format      string `validate:"required,min=1"`
	StdOut      bool
}

type ConfigFile struct {
	Basic  BasicFile
	Logger LoggerConfig `validate:"required"`

	Reloaders map[string]*ReloaderConfigFile `validate:"required,dive,min=1"`
}

func (reloader *ReloaderConfigFile) merge(basic *BasicFile) error {
	name := reloader.name

	if reloader.BFECluster == "" {
		reloader.BFECluster = basic.BFECluster
	}

	taskCount := len(reloader.MultiKeyFileTasks) + len(reloader.NormalFileTasks) + len(reloader.ExtraFileTasks)
	if taskCount == 0 {
		return fmt.Errorf("reloader %s should has at least one task", name)
	}

	for i, task := range reloader.NormalFileTasks {
		task.merge(basic)
		reloader.NormalFileTasks[i] = task
	}
	for i, task := range reloader.MultiKeyFileTasks {
		task.merge(basic)
		reloader.MultiKeyFileTasks[i] = task
	}
	for i, task := range reloader.ExtraFileTasks {
		task.merge(basic)
		reloader.ExtraFileTasks[i] = task
	}

	if reloader.ConfDir == "" {
		reloader.ConfDir = path.Join(basic.BFEConfDir, name)
	}
	if reloader.BFEReloadAPI == "" {
		reloader.BFEReloadAPI = "/reload/" + name
	}
	if reloader.BFEReloadTimeoutMs == 0 {
		reloader.BFEReloadTimeoutMs = basic.BFEReloadTimeoutMs
	}
	if reloader.ReloadIntervalMs == 0 {
		reloader.ReloadIntervalMs = basic.ReloadIntervalMs
	}

	return nil
}
